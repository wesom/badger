package websocket

import (
	"context"
	"net"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/wesom/badger/logging"
)

const (
	maxMessageSize = 4096

	bufferedSize = 128
)

// Connection represents a client connection to websocket server
type Connection struct {
	conn   *websocket.Conn
	id     uint64
	output chan []byte
	ctx    context.Context
	cancel context.CancelFunc
	once   sync.Once
	wg     sync.WaitGroup
}

// NewConnection return a new client connection
func NewConnection(id uint64, conn *websocket.Conn) *Connection {
	ctx, cancel := context.WithCancel(context.Background())
	c := &Connection{
		conn:   conn,
		id:     id,
		output: make(chan []byte, bufferedSize),
		ctx:    ctx,
		cancel: cancel,
	}
	return c
}

// Start create goroutines for reading and writing
func (c *Connection) Start() {
	c.wg.Add(2)
	go func() {
		readLoop(c)
		c.wg.Done()
	}()
	go func() {
		writeLoop(c)
		c.wg.Done()
	}()
}

// Close closes a client connection gracefully
func (c *Connection) Close() {
	c.once.Do(func() {
		go func() {
			c.cancel()
			c.conn.Close()
			c.wg.Wait()
			close(c.output)
		}()
	})
}

// ID return connection net id
func (c *Connection) ID() uint64 {
	return c.id
}

// RemoteAddr return the remote address
func (c *Connection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// LocalAddr return the local address
func (c *Connection) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *Connection) Write(buffer []byte) {
	c.output <- buffer
}

func readLoop(c *Connection) {
	defer func() {
		c.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	for {
		select {
		case <-c.ctx.Done():
			logging.Logger().Debugf("handleLoop quit clientId %d", c.id)
			return
		default:
			_, _, err := c.conn.ReadMessage()
			if err != nil {
				logging.Logger().Errorf("readMessage err: %d %v", c.id, err)
				return
			}
		}
	}
}

func writeLoop(c *Connection) {
	defer func() {
		c.Close()
	}()
	for {
		select {
		case <-c.ctx.Done():
			logging.Logger().Debugf("writeLoop quit clientId %d", c.id)
			return
		case buffer, ok := <-c.output:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.BinaryMessage, buffer); err != nil {
				logging.Logger().Errorf("wrapperwrite binary err %v", err)
				return
			}
		}
	}
}
