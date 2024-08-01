package badger

import (
	"net/http"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type onConnectFunc func(*Connection)
type onTextMessageFunc func(*Connection, []byte)
type onBinaryMessageFunc func(*Connection, []byte)
type onErrorFunc func(*Connection, error)
type onDisconnectFunc func(*Connection)

type WsGateWay struct {
	upgrader        *websocket.Upgrader
	opts            *Options
	nextid          uint64
	cmgr            *connMgr
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
		opts:            options,
		cmgr:            newConnMgr(),
		onConnect:       func(*Connection) {},
		onTextMessage:   func(*Connection, []byte) {},
		onBinaryMessage: func(*Connection, []byte) {},
		onError:         func(*Connection, error) {},
		onDisconnect:    func(*Connection) {},
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

	conn := newConnection(wsconn, newid, s)

	s.cmgr.add(conn)
	s.onConnect(conn)

	go conn.writeLoop()
	conn.readLoop()

	s.cmgr.remove(conn)
	s.onDisconnect(conn)
}

// Close
func (s *WsGateWay) Close() error {
	s.cmgr.close()
	return nil
}

func (s *WsGateWay) OnConnect(f func(*Connection)) {
	s.onConnect = f
}

func (s *WsGateWay) OnTextMessage(f func(*Connection, []byte)) {
	s.onTextMessage = f
}

func (s *WsGateWay) OnBinaryMessage(f func(*Connection, []byte)) {
	s.onBinaryMessage = f
}

func (s *WsGateWay) OnError(f func(*Connection, error)) {
	s.onError = f
}

func (s *WsGateWay) OnDisconnect(f func(*Connection)) {
	s.onDisconnect = f
}
