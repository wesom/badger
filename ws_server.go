package badger

import (
	"net"
	"net/http"

	"github.com/google/uuid"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

func generateConnID() string {
	return uuid.New().String()
}

type WsGateWay struct {
	upgrader        *websocket.Upgrader
	opts            *Options
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
		onConnect: func(connID string, remoteAddr net.Addr) error {
			return nil
		},
		onTextMessage:   func(connID string, data []byte) {},
		onBinaryMessage: func(connID string, data []byte) {},
		onError:         func(connID string, e error) {},
		onDisconnect:    func(connID string) {},
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

	newid := generateConnID()

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

func (s *WsGateWay) OnConnect(f func(connID string, remoteAddr net.Addr) error) {
	s.onConnect = f
}

func (s *WsGateWay) OnTextMessage(f func(connID string, data []byte)) {
	s.onTextMessage = f
}

func (s *WsGateWay) OnBinaryMessage(f func(connID string, data []byte)) {
	s.onBinaryMessage = f
}

func (s *WsGateWay) OnError(f func(connID string, e error)) {
	s.onError = f
}

func (s *WsGateWay) OnDisconnect(f func(connID string)) {
	s.onDisconnect = f
}

func (s *WsGateWay) SendTextMesaage(connID string, data []byte) {
	s.cmgr.Apply(connID, func(c *Connection) {
		c.WriteText(data)
	})
}

func (s *WsGateWay) SendBinaryMesaage(connID string, data []byte) {
	s.cmgr.Apply(connID, func(c *Connection) {
		c.WriteBinary(data)
	})
}

func (s *WsGateWay) Kick(connID string) {
	s.cmgr.Apply(connID, func(c *Connection) {
		c.Close()
	})
}
