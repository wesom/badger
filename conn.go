package badger

import (
	"errors"
	"net"
	"sync"

	"go.uber.org/zap"
)

// Abstract connect
type AbsConn interface {
	ReadMessage() ([]byte, error)

	WriteMessage(data []byte) error

	Close() error

	LocalAddr() net.Addr

	RemoteAddr() net.Addr
}

// Connection represents a wrapper connection
type Connection struct {
	logger   *zap.Logger
	h        EventHandler
	id       uint64
	absconn  AbsConn
	output   chan []byte
	wg       sync.WaitGroup
	mu       sync.RWMutex
	property map[string]interface{}
	closed   bool
}

// NewConnection return a new connection
func NewConnection(absconn AbsConn, id uint64, logger *zap.Logger, h EventHandler) *Connection {
	c := &Connection{
		logger:   logger,
		h:        h,
		id:       id,
		absconn:  absconn,
		output:   make(chan []byte, 128),
		property: make(map[string]interface{}),
		closed:   false,
	}
	return c
}

func (c *Connection) ConnID() uint64 {
	return c.id
}

func (c *Connection) LocalAddr() net.Addr {
	return c.absconn.LocalAddr()
}

func (c *Connection) RemoteAddr() net.Addr {
	return c.absconn.RemoteAddr()
}

func (c *Connection) SetProperty(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.property[key] = value
}

func (c *Connection) GetProperty(key string) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if value, ok := c.property[key]; ok {
		return value, nil
	}
	return nil, errors.New("no property data found")
}

func (c *Connection) RemoveProperty(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.property, key)
}

func (c *Connection) Int64(key string) int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.property[key]
	if !ok {
		return 0
	}
	value, ok := v.(int64)
	if !ok {
		return 0
	}
	return value
}

func (c *Connection) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return
	}
	c.output <- nil
}

func (c *Connection) Write(buffer []byte) {
	if buffer == nil {
		panic("buffer must be not nil")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return
	}
	c.output <- buffer
}

func (c *Connection) release() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.closed {
		close(c.output)
		c.closed = true
	}
}

// Start create goroutines for reading and writing
func (c *Connection) Start() {
	c.wg.Add(2)
	go c.readLoop()
	go c.writeLoop()
	c.wg.Wait()
}

func (c *Connection) readLoop() {
	defer func() {
		if err := recover(); err != nil {
			c.logger.Error("readLoop catch panic", zap.Any("err", err))
		}
		// notify writeLoop to quit
		c.release()
		c.wg.Done()
		c.logger.Info("readLoop quit", zap.Uint64("connID", c.id))
	}()

	for {
		data, err := c.absconn.ReadMessage()
		if err != nil {
			c.logger.Error("read message error", zap.Uint64("connID", c.id), zap.Error(err))
			return
		}
		c.h.OnMessage(c, data)
	}
}

func (c *Connection) writeLoop() {
	defer func() {
		if err := recover(); err != nil {
			c.logger.Error("writeLoop catch panic", zap.Any("err", err))
		}
		c.release()
		c.absconn.Close()
		c.wg.Done()
		c.logger.Info("writeLoop quit", zap.Uint64("connID", c.id))
	}()

	for buffer := range c.output {
		if buffer == nil {
			c.logger.Error("writeLoop receive active quit", zap.Uint64("connID", c.id))
			return
		}
		if err := c.absconn.WriteMessage(buffer); err != nil {
			c.logger.Error("write message error", zap.Uint64("connID", c.id), zap.Error(err))
			return
		}
	}
}
