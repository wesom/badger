package badger

import (
	"net"
	"net/http"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		EnableCompression: true,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

type WsServer struct {
	opts            *Options
	nextid          uint64
	cmgr            *ConnMgr
	onConnect       onConnectFunc
	onTextMessage   onTextMessageFunc
	onBinaryMessage onBinaryMessageFunc
	onDisconnect    onDisconnectFunc
	onError         onErrorFunc
}

func NewWsServer(opts ...Option) *WsServer {
	options := &Options{
		Logger: zap.NewNop(),
	}
	for _, o := range opts {
		o(options)
	}
	s := &WsServer{
		opts: options,
		cmgr: NewConnMgr(),
		onConnect: func(connID uint64, remoteAddr net.Addr) error {
			return nil
		},
		onTextMessage:   func(connID uint64, data []byte) {},
		onBinaryMessage: func(connID uint64, data []byte) {},
		onError:         func(connID uint64, e error) {},
		onDisconnect:    func(connID uint64) {},
	}

	return s
}

func (s *WsServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wsconn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.opts.Logger.Error("ws upgrade err", zap.Error(err))
		return
	}

	newid := atomic.AddUint64(&s.nextid, 1)

	err = s.onConnect(newid, wsconn.RemoteAddr())
	if err != nil {
		s.opts.Logger.Error("ws onConnect err", zap.Error(err))
		return
	}

	conn := NewConnection(wsconn, newid, s)
	s.cmgr.Add(conn)
	conn.Start()
	s.cmgr.Remove(conn)
	s.onDisconnect(conn.ConnID())
}

// Close
func (s *WsServer) Close() error {
	s.cmgr.Close()
	return nil
}

func (s *WsServer) OnConnect(f func(connID uint64, remoteAddr net.Addr) error) {
	s.onConnect = f
}

func (s *WsServer) OnTextMessage(f func(connID uint64, data []byte)) {
	s.onTextMessage = f
}

func (s *WsServer) OnBinayMessage(f func(connID uint64, data []byte)) {
	s.onBinaryMessage = f
}

func (s *WsServer) OnError(f func(connID uint64, e error)) {
	s.onError = f
}

func (s *WsServer) OnDisconnect(f func(connID uint64)) {
	s.onDisconnect = f
}

func (s *WsServer) SendTextMesaage(connID uint64, data []byte) {
	conn := s.cmgr.Find(connID)
	if conn != nil {
		conn.WriteText(data)
	}
}

func (s *WsServer) SendBinaryMesaage(connID uint64, data []byte) {
	conn := s.cmgr.Find(connID)
	if conn != nil {
		conn.WriteBinary(data)
	}
}

func (s *WsServer) Kick(connID uint64) {
	conn := s.cmgr.Find(connID)
	if conn != nil {
		conn.Close()
	}
}
