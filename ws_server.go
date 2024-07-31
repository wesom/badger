package badger

import (
	"net"
	"net/http"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type WsGateWay struct {
	upgrader        *websocket.Upgrader
	opts            *Options
	nextid          uint64
	cmgr            *ConnMgr
	onConnect       onConnectFunc
	onTextMessage   onTextMessageFunc
	onBinaryMessage onBinaryMessageFunc
	onDisconnect    onDisconnectFunc
	onError         onErrorFunc
}

func NewWsGateWay(opts ...Option) *WsGateWay {
	options := &Options{
		Logger:   zap.NewNop(),
		MaxConns: 1024,
	}
	for _, o := range opts {
		o(options)
	}
	s := &WsGateWay{
		opts: options,
		cmgr: NewConnMgr(options.MaxConns),
		onConnect: func(connID uint64, remoteAddr net.Addr) error {
			return nil
		},
		onTextMessage:   func(connID uint64, data []byte) {},
		onBinaryMessage: func(connID uint64, data []byte) {},
		onError:         func(connID uint64, e error) {},
		onDisconnect:    func(connID uint64) {},
	}

	// Set upgrader options
	s.upgrader = &websocket.Upgrader{
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		EnableCompression: true,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	return s
}

func (s *WsGateWay) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wsconn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.opts.Logger.Error("ws upgrade err", zap.Error(err))
		http.Error(w, "Could not upgrade to WebSocket", http.StatusBadRequest)
		return
	}

	newid := atomic.AddUint64(&s.nextid, 1)

	err = s.onConnect(newid, wsconn.RemoteAddr())
	if err != nil {
		s.opts.Logger.Error("ws onConnect err", zap.Error(err))
		wsconn.Close()
		return
	}

	conn := NewConnection(wsconn, newid, s)
	defer func() {
		s.cmgr.Remove(conn)
		s.onDisconnect(conn.ConnID())
	}()

	s.cmgr.Add(conn)
	conn.Start()
}

// Close
func (s *WsGateWay) Close() error {
	s.cmgr.Close()
	return nil
}

func (s *WsGateWay) OnConnect(f func(connID uint64, remoteAddr net.Addr) error) {
	s.onConnect = f
}

func (s *WsGateWay) OnTextMessage(f func(connID uint64, data []byte)) {
	s.onTextMessage = f
}

func (s *WsGateWay) OnBinaryMessage(f func(connID uint64, data []byte)) {
	s.onBinaryMessage = f
}

func (s *WsGateWay) OnError(f func(connID uint64, e error)) {
	s.onError = f
}

func (s *WsGateWay) OnDisconnect(f func(connID uint64)) {
	s.onDisconnect = f
}

func (s *WsGateWay) SendTextMesaage(connID uint64, data []byte) {
	s.cmgr.Apply(connID, func(c *Connection) {
		c.WriteText(data)
	})
}

func (s *WsGateWay) SendBinaryMesaage(connID uint64, data []byte) {
	s.cmgr.Apply(connID, func(c *Connection) {
		c.WriteBinary(data)
	})
}

func (s *WsGateWay) Kick(connID uint64) {
	s.cmgr.Apply(connID, func(c *Connection) {
		c.Close()
	})
}
