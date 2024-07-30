package badger

import (
	"net"
	"sync"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type imessage struct {
	mt   int
	data []byte
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
	closedFlag bool
}

// NewConnection return a new connection
func NewConnection(conn *websocket.Conn, id uint64, s *WsGateWay) *Connection {
	c := &Connection{
		s:          s,
		id:         id,
		wsconn:     conn,
		output:     make(chan imessage, 128),
		outputDone: make(chan struct{}),
		closedFlag: false,
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
	return c.closedFlag
}

func (c *Connection) close() {
	c.mu.Lock()
	isclosed := c.closedFlag
	c.closedFlag = true
	c.mu.Unlock()

	if !isclosed {
		c.wsconn.Close()
		c.outputDone <- struct{}{}
	}
}

func (c *Connection) writeMessage(m imessage) {
	if c.closed() {
		c.s.onError(c.ConnID(), ErrConnClosed)
		return
	}
	select {
	case c.output <- m:
	default:
		c.s.onError(c.ConnID(), ErrBufferFull)
	}
}

func (c *Connection) Close() {
	c.writeMessage(imessage{mt: websocket.CloseMessage, data: []byte{}})
}

func (c *Connection) WriteText(text []byte) {
	c.writeMessage(imessage{mt: websocket.TextMessage, data: text})
}

func (c *Connection) WriteBinary(buffer []byte) {
	c.writeMessage(imessage{mt: websocket.BinaryMessage, data: buffer})
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
			c.s.opts.Logger.Error("readLoop catch panic", zap.Any("err", err))
		}
		c.close()
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
			c.s.onTextMessage(c.ConnID(), msg)
		case websocket.BinaryMessage:
			c.s.onBinaryMessage(c.ConnID(), msg)
		}
	}
}

func (c *Connection) writeLoop() {
	defer func() {
		if err := recover(); err != nil {
			c.s.opts.Logger.Error("writeLoop catch panic", zap.Any("err", err))
		}
		c.close()
		c.wg.Done()
		c.s.opts.Logger.Info("writeLoop quit", zap.Uint64("connID", c.id))
	}()

	for {
		select {
		case msg := <-c.output:
			if msg.mt == websocket.CloseMessage {
				c.s.opts.Logger.Info("writeLoop active close", zap.Uint64("connID", c.id))
				return
			}
			if err := c.wsconn.WriteMessage(msg.mt, msg.data); err != nil {
				c.s.opts.Logger.Error("write message error", zap.Uint64("connID", c.id), zap.Error(err))
				return
			}
		case <-c.outputDone:
			c.s.opts.Logger.Info("receive exit signal from readloop", zap.Uint64("connID", c.id))
			return
		}
	}

}
