package badger

import (
	"sync"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type imessage struct {
	messageType int
	data        []byte
}

// connection represents a wrapper connection
type connection struct {
	s          *WsGateWay
	id         string
	wsconn     *websocket.Conn
	output     chan imessage
	outputDone chan struct{}
	wg         sync.WaitGroup
	mu         sync.RWMutex
	stopFlag   bool
}

// NewConnection return a new connection
func newConnection(conn *websocket.Conn, id string, s *WsGateWay) *connection {
	c := &connection{
		s:          s,
		id:         id,
		wsconn:     conn,
		output:     make(chan imessage, 128),
		outputDone: make(chan struct{}),
		stopFlag:   false,
	}
	return c
}

func (c *connection) connID() string {
	return c.id
}

func (c *connection) closed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.stopFlag
}

func (c *connection) stop() {
	c.mu.Lock()
	isstoped := c.stopFlag
	c.stopFlag = true
	c.mu.Unlock()

	if !isstoped {
		c.wsconn.Close()
		c.outputDone <- struct{}{}
	}
}

func (c *connection) writeMessage(m imessage) {
	if c.closed() {
		c.s.onError(c.connID(), ErrConnClosed)
		return
	}
	select {
	case c.output <- m:
	default:
		c.s.onError(c.connID(), ErrBufferFull)
	}
}

func (c *connection) writeClose() {
	c.writeMessage(imessage{messageType: websocket.CloseMessage, data: []byte{}})
}

func (c *connection) writeText(text []byte) {
	c.writeMessage(imessage{messageType: websocket.TextMessage, data: text})
}

func (c *connection) writeBinary(buffer []byte) {
	c.writeMessage(imessage{messageType: websocket.BinaryMessage, data: buffer})
}

// Start create goroutines for reading and writing
func (c *connection) start() {
	c.wg.Add(2)
	go c.readLoop()
	go c.writeLoop()
	c.wg.Wait()
}

func (c *connection) readLoop() {
	defer func() {
		if err := recover(); err != nil {
			c.s.opts.Logger.Error("readLoop catch panic", zap.Any("err", err))
		}
		c.stop()
		c.wg.Done()
		c.s.opts.Logger.Info("readLoop quit", zap.String("connID", c.id))
	}()

	for {
		t, msg, err := c.wsconn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.s.opts.Logger.Error("read message unexpected error", zap.String("connID", c.id), zap.Error(err))
			}
			break
		}
		switch t {
		case websocket.TextMessage:
			c.s.onTextMessage(c.connID(), msg)
		case websocket.BinaryMessage:
			c.s.onBinaryMessage(c.connID(), msg)
		}
	}
}

func (c *connection) writeLoop() {
	defer func() {
		if err := recover(); err != nil {
			c.s.opts.Logger.Error("writeLoop catch panic", zap.Any("err", err))
		}
		c.stop()
		c.wg.Done()
		c.s.opts.Logger.Info("writeLoop quit", zap.String("connID", c.id))
	}()

	for {
		select {
		case msg := <-c.output:
			if msg.messageType == websocket.CloseMessage {
				c.s.opts.Logger.Info("writeLoop active close", zap.String("connID", c.id))
				return
			}
			if err := c.wsconn.WriteMessage(msg.messageType, msg.data); err != nil {
				c.s.opts.Logger.Error("write message error", zap.String("connID", c.id), zap.Error(err))
				return
			}
		case <-c.outputDone:
			c.s.opts.Logger.Info("receive exit signal from readloop", zap.String("connID", c.id))
			return
		}
	}

}
