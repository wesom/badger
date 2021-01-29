package websocket

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/wesom/badger/gate"
)

type connMgr struct {
	mu    sync.RWMutex
	conns map[uint64]*Connection
}

func newConnMgr() *connMgr {
	return &connMgr{
		conns: make(map[uint64]*Connection),
	}
}

func (cmgr *connMgr) Count() int {
	cmgr.mu.RLock()
	defer cmgr.mu.RUnlock()
	return len(cmgr.conns)
}

func (cmgr *connMgr) Add(conn *Connection) {
	cmgr.mu.Lock()
	defer cmgr.mu.Unlock()
	cmgr.conns[conn.SessionID] = conn
}

func (cmgr *connMgr) Remove(id uint64) {
	cmgr.mu.Lock()
	defer cmgr.mu.Unlock()
	delete(cmgr.conns, id)
}

func (cmgr *connMgr) Load(id uint64) (*Connection, error) {
	cmgr.mu.RLock()
	defer cmgr.mu.RUnlock()
	conn, ok := cmgr.conns[id]
	if ok {
		return conn, nil
	}
	return nil, errors.New("connection do not exist")
}

func (cmgr *connMgr) CloseAll() {
	conns := make(map[uint64]*Connection)
	cmgr.mu.Lock()
	for id, conn := range cmgr.conns {
		conns[id] = conn
	}
	cmgr.mu.Unlock()

	for _, nconn := range conns {
		nconn.Close()
	}
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
	cmgr    *connMgr
	httpsrv *http.Server
}

var (
	// DefaultAddress to ws
	DefaultAddress = ":8000"
	// DefaultMaxConns to limit max connections
	DefaultMaxConns = 1024
)

// NewGate retrun a gate implement with websocket
func NewGate(opts ...gate.Option) gate.Gate {
	options := gate.Options{
		Address:  DefaultAddress,
		MaxConns: DefaultMaxConns,
	}

	for _, o := range opts {
		o(&options)
	}

	s := &wsServer{
		options: options,
		cmgr:    newConnMgr(),
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
		s.options.Logger.Errorf("ws upgrade err %v", err)
		return
	}
	if s.ConnCount() > s.options.MaxConns {
		return
	}
	conn := NewConnection(wsconn, s)
	conn.Start()
}

func (s *wsServer) ConnCount() int {
	return s.cmgr.Count()
}

func (s *wsServer) Start() error {
	ln, err := net.Listen("tcp", s.options.Address)
	if err != nil {
		return err
	}

	s.options.Logger.Infof("wsserver listening on %s", ln.Addr().String())

	go func() {
		if err := s.httpsrv.Serve(ln); err != http.ErrServerClosed {
			s.options.Logger.Errorf("wsserver serve failed: %v", err)
		}
	}()

	return nil
}

func (s *wsServer) Stop() error {
	s.cmgr.CloseAll()

	if err := s.httpsrv.Shutdown(context.TODO()); err != nil {
		s.options.Logger.Errorf("wsserver shutdown failed: %v", err)
		return err
	}

	s.options.Logger.Infof("wsserver stopped")

	return nil
}
