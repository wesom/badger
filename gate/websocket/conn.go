package websocket

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/wesom/badger/dispatch"
)

const (
	maxMessageSize = 4096

	bufferedSize = 128
)

type wmessage struct {
	conn *WsConnection
	data []byte
	name string
}

func (m *wmessage) Conn() dispatch.Conn {
	return m.conn
}

func (m *wmessage) Key() uint64 {
	return m.conn.ID()
}

func (m *wmessage) Name() string {
	return m.name
}

// WsConnection represents a client connection to websocket server
type WsConnection struct {
	id       uint64
	belong   *WsServer
	conn     *websocket.Conn
	output   chan []byte
	ctx      context.Context
	cancel   context.CancelFunc
	mu       sync.Mutex
	isClosed bool
	wg       sync.WaitGroup
}

// NewConnection return a new client connection
func NewConnection(conn *websocket.Conn, srv *WsServer) *WsConnection {
	ctx, cancel := context.WithCancel(context.Background())
	c := &WsConnection{
		id:     srv.options.UIDFn(),
		belong: srv,
		conn:   conn,
		output: make(chan []byte, bufferedSize),
		ctx:    ctx,
		cancel: cancel,
	}
	srv.cmgr.Add(c)
	return c
}

// Start create goroutines for reading and writing
func (c *WsConnection) Start() {
	c.wg.Add(2)
	go c.readLoop()
	go c.writeLoop()
}

// Close closes a client connection gracefully
func (c *WsConnection) Close() {
	go func() {
		c.mu.Lock()
		if c.isClosed {
			c.mu.Unlock()
			return
		}
		c.conn.Close()
		close(c.output)
		c.isClosed = true
		c.mu.Unlock()
		c.belong.cmgr.Remove(c.ID())
		c.cancel()
		c.wg.Wait()
	}()
}

// ID return session id
func (c *WsConnection) ID() uint64 {
	return c.id
}

// RemoteAddr return the remote address
func (c *WsConnection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// LocalAddr return the local address
func (c *WsConnection) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *WsConnection) Write(buffer []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.isClosed {
		return fmt.Errorf("write closed, session: %d ", c.id)
	}
	select {
	case c.output <- buffer:
	default:
		return fmt.Errorf("write full buffer, session: %d", c.id)
	}
	return nil
}

func (c *WsConnection) readLoop() {
	defer func() {
		c.wg.Done()
		c.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	for {
		select {
		case <-c.ctx.Done():
			c.belong.options.Logger.Debugf("handleLoop quit clientId %d", c.id)
			return
		default:
			_, data, err := c.conn.ReadMessage()
			if err != nil {
				c.belong.options.Logger.Errorf("readMessage err: %d %v", c.id, err)
				return
			}
			if c.belong.options.Disp == nil {
				return
			}
			msg := &wmessage{
				conn: c,
				data: data,
			}
			c.belong.options.Disp.Put(msg)
		}
	}
}

func (c *WsConnection) writeLoop() {
	defer func() {
		c.wg.Done()
		c.Close()
	}()
	for {
		select {
		case <-c.ctx.Done():
			c.belong.options.Logger.Debugf("writeLoop quit clientId %d", c.id)
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
