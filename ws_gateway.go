package badger

import (
	"net/http"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

type handleConnectFunc func(*Connection)
type handleMessageFunc func(*Connection, []byte)
type handleErrorFunc func(*Connection, error)

type WsGateWay struct {
	upgrader        *websocket.Upgrader
	opts            *Options
	closed          atomic.Bool
	hub             *hub
	onConnect       handleConnectFunc
	onDisconnect    handleConnectFunc
	onMessage       handleMessageFunc
	onBinaryMessage handleMessageFunc
	onError         handleErrorFunc
}

func NewWsGateWay(opts ...Option) *WsGateWay {
	options := &Options{
		OutputBufferSize: 128,
	}
	for _, o := range opts {
		o(options)
	}
	s := &WsGateWay{
		opts:            options,
		hub:             newHub(),
		onConnect:       func(*Connection) {},
		onDisconnect:    func(*Connection) {},
		onMessage:       func(*Connection, []byte) {},
		onBinaryMessage: func(*Connection, []byte) {},
		onError:         func(*Connection, error) {},
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

	s.closed.Store(false)

	return s
}

func (s *WsGateWay) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if s.IsClosed() {
		http.Error(w, "Server is closed", http.StatusNotAcceptable)
		return
	}

	wsconn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not upgrade to WebSocket", http.StatusBadRequest)
		return
	}

	conn := newConnection(wsconn, s)

	s.hub.add(conn)
	s.onConnect(conn)

	go conn.writeLoop()
	conn.readLoop()

	s.hub.del(conn)
	conn.iclose()
	s.onDisconnect(conn)
}

func (s *WsGateWay) Len() int {
	return s.hub.len()
}

func (s *WsGateWay) IsClosed() bool {
	return s.closed.Load()
}

func (s *WsGateWay) Close() {
	if s.IsClosed() {
		return
	}
	s.closed.Store(true)

	s.hub.each(func(c *Connection) {
		c.Close()
	})

}

func (s *WsGateWay) OnConnect(f func(*Connection)) {
	s.onConnect = f
}

func (s *WsGateWay) OnDisconnect(f func(*Connection)) {
	s.onDisconnect = f
}

func (s *WsGateWay) OnMessage(f func(*Connection, []byte)) {
	s.onMessage = f
}

func (s *WsGateWay) OnBinaryMessage(f func(*Connection, []byte)) {
	s.onBinaryMessage = f
}

func (s *WsGateWay) OnError(f func(*Connection, error)) {
	s.onError = f
}
