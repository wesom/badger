package badger

import (
	"net"
	"sync"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type imessage struct {
	messageType int
	data        []byte
}

// Connection represents a wrapper connection
type Connection struct {
	s          *WsGateWay
	id         uint64
	wsconn     *websocket.Conn
	output     chan imessage
	outputDone chan struct{}
	wg         sync.WaitGroup
	mu         sync.RWMutex
	stopFlag   bool
}

// NewConnection return a new connection
func newConnection(conn *websocket.Conn, id uint64, s *WsGateWay) *Connection {
	c := &Connection{
		s:          s,
		id:         id,
		wsconn:     conn,
		output:     make(chan imessage, 128),
		outputDone: make(chan struct{}),
		stopFlag:   false,
	}
	return c
}

func (c *Connection) ConnID() uint64 {
	return c.id
}

func (c *Connection) LocalAddr() net.Addr {
	return c.wsconn.LocalAddr()
}

func (c *Connection) RemoteAddr() net.Addr {
	return c.wsconn.RemoteAddr()
}

func (c *Connection) closed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.stopFlag
}

func (c *Connection) stop() {
	c.mu.Lock()
	isstoped := c.stopFlag
	c.stopFlag = true
	c.mu.Unlock()

	if !isstoped {
		c.wsconn.Close()
		c.outputDone <- struct{}{}
	}
}

func (c *Connection) write(m imessage) {
	if c.closed() {
		c.s.onError(c, ErrConnClosed)
		return
	}
	select {
	case c.output <- m:
	default:
		c.s.onError(c, ErrBufferFull)
	}
}

func (c *Connection) Close() {
	c.write(imessage{messageType: websocket.CloseMessage, data: []byte{}})
}

func (c *Connection) WriteTextMessage(text []byte) {
	c.write(imessage{messageType: websocket.TextMessage, data: text})
}

func (c *Connection) WriteBinaryMessage(buffer []byte) {
	c.write(imessage{messageType: websocket.BinaryMessage, data: buffer})
}

// Start create goroutines for reading and writing
func (c *Connection) start() {
	c.wg.Add(2)
	go c.readLoop()
	go c.writeLoop()
	c.wg.Wait()
}

func (c *Connection) readLoop() {
	defer func() {
		if err := recover(); err != nil {
			c.s.opts.Logger.Error("readLoop catch panic", zap.Any("err", err))
		}
		c.stop()
		c.wg.Done()
		c.s.opts.Logger.Info("readLoop quit", zap.Uint64("connID", c.id))
	}()

	for {
		t, msg, err := c.wsconn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.s.opts.Logger.Error("read message unexpected error", zap.Uint64("connID", c.id), zap.Error(err))
			}
			break
		}
		switch t {
		case websocket.TextMessage:
			c.s.onTextMessage(c, msg)
		case websocket.BinaryMessage:
			c.s.onBinaryMessage(c, msg)
		}
	}
}

func (c *Connection) writeLoop() {
	defer func() {
		if err := recover(); err != nil {
			c.s.opts.Logger.Error("writeLoop catch panic", zap.Any("err", err))
		}
		c.stop()
		c.wg.Done()
		c.s.opts.Logger.Info("writeLoop quit", zap.Uint64("connID", c.id))
	}()

	for {
		select {
		case msg := <-c.output:
			if msg.messageType == websocket.CloseMessage {
				c.s.opts.Logger.Info("writeLoop active close", zap.Uint64("connID", c.id))
				return
			}
			if err := c.wsconn.WriteMessage(msg.messageType, msg.data); err != nil {
				c.s.opts.Logger.Error("write message error", zap.Uint64("connID", c.id), zap.Error(err))
				return
			}
		case <-c.outputDone:
			c.s.opts.Logger.Info("receive exit signal from readloop", zap.Uint64("connID", c.id))
			return
		}
	}

}
