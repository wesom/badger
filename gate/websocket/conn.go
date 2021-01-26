package websocket

import (
	"context"
	"net"
	"sync"

	"github.com/gorilla/websocket"
)

const (
	maxMessageSize = 4096

	bufferedSize = 128
)

// Connection represents a client connection to websocket server
type Connection struct {
	SessionID uint64
	belong    *wsServer
	conn      *websocket.Conn
	output    chan []byte
	ctx       context.Context
	cancel    context.CancelFunc
	once      sync.Once
	wg        sync.WaitGroup
}

// NewConnection return a new client connection
func NewConnection(conn *websocket.Conn, srv *wsServer) *Connection {
	ctx, cancel := context.WithCancel(context.Background())
	c := &Connection{
		SessionID: srv.options.UIDFn(),
		belong:    srv,
		conn:      conn,
		output:    make(chan []byte, bufferedSize),
		ctx:       ctx,
		cancel:    cancel,
	}
	srv.cmgr.Add(c)
	return c
}

// Start create goroutines for reading and writing
func (c *Connection) Start() {
	c.wg.Add(2)
	go func() {
		c.readLoop()
		c.wg.Done()
	}()
	go func() {
		c.writeLoop()
		c.wg.Done()
	}()
}

// Close closes a client connection gracefully
func (c *Connection) Close() {
	c.once.Do(func() {
		go func() {
			c.belong.cmgr.Remove(c.SessionID)
			c.cancel()
			c.conn.Close()
			c.wg.Wait()
			close(c.output)
		}()
	})
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

func (c *Connection) readLoop() {
	defer func() {
		c.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	for {
		select {
		case <-c.ctx.Done():
			// logger.Debugf("handleLoop quit clientId %s", c.SessionID)
			return
		default:
			_, _, err := c.conn.ReadMessage()
			if err != nil {
				// logger.Errorf("readMessage err: %s %v", c.SessionID, err)
				return
			}
		}
	}
}

func (c *Connection) writeLoop() {
	defer func() {
		c.Close()
	}()
	for {
		select {
		case <-c.ctx.Done():
			// logger.Debugf("writeLoop quit clientId %s", c.SessionID)
			return
		case buffer, ok := <-c.output:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.BinaryMessage, buffer); err != nil {
				// logger.Errorf("wrapperwrite binary err %v", err)
				return
			}
		}
	}
}
