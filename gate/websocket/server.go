package websocket

import (
	"context"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/wesom/badger/gate"
)

type connMgr struct {
	m *sync.Map
}

func NewConnMgr() *connMgr {
	return &connMgr{
		m: &sync.Map{},
	}
}

func (cmgr *connMgr) Count() int {
	var count int
	cmgr.m.Range(func(k, v interface{}) bool {
		count += 1
		return true
	})
	return count
}

func (cmgr *connMgr) Add(conn *Connection) {
	cmgr.m.Store(conn.SessionID, conn)
}

func (cmgr *connMgr) Remove(id string) {
	cmgr.m.Delete(id)
}

func (cmgr *connMgr) Load(id string) (*Connection, bool) {
	v, ok := cmgr.m.Load(id)
	if ok {
		return v.(*Connection), true
	}
	return nil, false
}

func (cmgr *connMgr) CloseAll() {
	cmgr.m.Range(func(k, v interface{}) bool {
		conn := v.(*Connection)
		conn.Close()
		cmgr.m.Delete(k)
		return true
	})
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:    1024,
	WriteBufferSize:   1024,
	EnableCompression: true,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type wsServer struct {
	options gate.Options
	conns   *connMgr
	httpsrv *http.Server
}

var (
	// DefaultAddress to ws
	DefaultAddress  = ":8000"
	DefaultMaxConns = 1024
)

// NewServer retrun a gate implement with websocket
func NewServer(opts ...gate.Option) gate.Gate {
	options := gate.Options{
		Address:  DefaultAddress,
		MaxConns: DefaultMaxConns,
	}

	for _, o := range opts {
		o(&options)
	}

	s := &wsServer{
		options: options,
		conns:   NewConnMgr(),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.wsHandler)

	s.httpsrv = &http.Server{
		Handler: mux,
	}

	return s
}

func (s *wsServer) wsHandler(w http.ResponseWriter, r *http.Request) {
	wsconn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// logger.Errorf("ws upgrade err %v", err)
		return
	}
	if s.ConnCount() > s.options.MaxConns {
		return
	}
	conn := NewConnection(wsconn, s)
	conn.Start()
	s.conns.Add(conn)
}

func (s *wsServer) ConnCount() int {
	return s.conns.Count()
}

func (s *wsServer) Start() error {
	ln, err := net.Listen("tcp", s.options.Address)
	if err != nil {
		return err
	}

	// logger.Infof("wsserver listening on %s", ln.Addr().String())

	go func() {
		if err := s.httpsrv.Serve(ln); err != http.ErrServerClosed {
			// logger.Errorf("wsserver serve failed: %v", err)
		}
	}()

	return nil
}

func (s *wsServer) Stop() error {
	if err := s.httpsrv.Shutdown(context.TODO()); err != nil {
		// logger.Errorf("wsserver shutdown failed: %v", err)
		return err
	}

	// logger.Info("wsserver stopped")

	return nil
}
