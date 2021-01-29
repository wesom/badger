package websocket

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/gorilla/websocket"
)

const (
	maxMessageSize = 4096

	bufferedSize = 128
)

type wmessage struct {
	Conn *Connection
	data []byte
	name string
}

func (m *wmessage) Key() uint64 {
	return m.Conn.SessionID
}

func (m *wmessage) Name() string {
	return m.name
}

// Connection represents a client connection to websocket server
type Connection struct {
	SessionID uint64
	belong    *wsServer
	conn      *websocket.Conn
	output    chan []byte
	ctx       context.Context
	cancel    context.CancelFunc
	mu        sync.RWMutex
	isClosed  bool
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
	go c.readLoop()
	go c.writeLoop()
}

// Close closes a client connection gracefully
func (c *Connection) Close() {
	go func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		if c.isClosed {
			return
		}
		c.belong.cmgr.Remove(c.SessionID)
		c.cancel()
		c.conn.Close()
		c.wg.Wait()
		close(c.output)
		c.isClosed = true
	}()
}

// RemoteAddr return the remote address
func (c *Connection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// LocalAddr return the local address
func (c *Connection) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *Connection) Write(buffer []byte) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.isClosed {
		return fmt.Errorf("write closed, session: %d ", c.SessionID)
	}
	select {
	case c.output <- buffer:
	default:
		return fmt.Errorf("write full buffer, session: %d", c.SessionID)
	}
	return nil
}

func (c *Connection) readLoop() {
	defer func() {
		c.wg.Done()
		c.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	for {
		select {
		case <-c.ctx.Done():
			c.belong.options.Logger.Debugf("handleLoop quit clientId %s", c.SessionID)
			return
		default:
			_, data, err := c.conn.ReadMessage()
			if err != nil {
				c.belong.options.Logger.Errorf("readMessage err: %s %v", c.SessionID, err)
				return
			}
			if c.belong.options.Disp == nil {
				return
			}
			msg := &wmessage{
				Conn: c,
				data: data,
			}
			c.belong.options.Disp.Delivery(msg)
		}
	}
}

func (c *Connection) writeLoop() {
	defer func() {
		c.wg.Done()
		c.Close()
	}()
	for {
		select {
		case <-c.ctx.Done():
			c.belong.options.Logger.Debugf("writeLoop quit clientId %s", c.SessionID)
			return
		case buffer, ok := <-c.output:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.BinaryMessage, buffer); err != nil {
				c.belong.options.Logger.Errorf("wrapperwrite binary err %v", err)
				return
			}
		}
	}
}
