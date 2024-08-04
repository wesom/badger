package badger

import (
	"net"
	"sync"

	"github.com/gorilla/websocket"
)

type imessage struct {
	messageType int
	data        []byte
}

// Connection represents a wrapper connection
type Connection struct {
	s          *WsGateWay
	wsconn     *websocket.Conn
	output     chan imessage
	outputDone chan struct{}
	mu         sync.RWMutex
	stopFlag   bool
	property   map[string]any
}

// NewConnection return a new connection
func newConnection(conn *websocket.Conn, s *WsGateWay) *Connection {
	c := &Connection{
		s:          s,
		wsconn:     conn,
		output:     make(chan imessage, s.opts.OutputBufferSize),
		outputDone: make(chan struct{}),
		stopFlag:   false,
	}
	return c
}

func (c *Connection) LocalAddr() net.Addr {
	return c.wsconn.LocalAddr()
}

func (c *Connection) RemoteAddr() net.Addr {
	return c.wsconn.RemoteAddr()
}

func (c *Connection) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.property == nil {
		c.property = make(map[string]any)
		c.property[key] = value
	}
}

func (c *Connection) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.property != nil {
		v, exist := c.property[key]
		return v, exist
	}
	return nil, false
}

func (c *Connection) MustGet(key string) any {
	if v, exist := c.Get(key); exist {
		return v
	}
	panic("key: " + key + " don't exist")
}

func (c *Connection) UnSet(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.property != nil {
		delete(c.property, key)
	}
}

func (c *Connection) closed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.stopFlag
}

func (c *Connection) stop() {
	c.mu.Lock()
	isStoped := c.stopFlag
	c.stopFlag = true
	c.mu.Unlock()

	if !isStoped {
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

func (c *Connection) readLoop() {
	for {
		t, msg, err := c.wsconn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.s.onError(c, err)
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
		c.stop()
	}()

	for {
		select {
		case msg := <-c.output:
			if msg.messageType == websocket.CloseMessage {
				return
			}
			if err := c.wsconn.WriteMessage(msg.messageType, msg.data); err != nil {
				c.s.onError(c, err)
				return
			}
		case <-c.outputDone:
			return
		}
	}

}
