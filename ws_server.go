package badger

import (
	"net/http"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var (
	DefaultMaxConn      = 1024
	DefaultLogger       = zap.NewNop()
	DefaultEventHandler = &EventServer{}

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
	opts   *Options
	nextid uint64
	cmgr   *ConnMgr
}

func NewWsServer(opts ...Option) *WsServer {
	options := &Options{
		MaxConn: DefaultMaxConn,
		Logger:  DefaultLogger,
		Handler: DefaultEventHandler,
	}
	for _, o := range opts {
		o(options)
	}
	s := &WsServer{
		opts: options,
		cmgr: NewConnMgr(),
	}

	return s
}

func (s *WsServer) Handle(w http.ResponseWriter, r *http.Request) {
	wsconn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.opts.Logger.Error("ws upgrade err", zap.Error(err))
		return
	}
	if s.cmgr.Count() > s.opts.MaxConn {
		return
	}

	newid := atomic.AddUint64(&s.nextid, 1)
	conn := NewConnection(&WsConnection{conn: wsconn}, newid, s.opts.Logger, s.opts.Handler)

	s.cmgr.Add(conn)
	s.opts.Handler.OnConnect(conn)
	conn.Start()
	s.cmgr.Remove(conn)
	s.opts.Handler.OnDisconnect(conn)
}

// Close
func (s *WsServer) Close() error {
	s.cmgr.Close()
	return nil
}
