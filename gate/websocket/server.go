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
	conns map[uint64]*WsConnection
}

func newConnMgr() *connMgr {
	return &connMgr{
		conns: make(map[uint64]*WsConnection),
	}
}

func (cmgr *connMgr) Count() int {
	cmgr.mu.RLock()
	defer cmgr.mu.RUnlock()
	return len(cmgr.conns)
}

func (cmgr *connMgr) Add(conn *WsConnection) {
	cmgr.mu.Lock()
	defer cmgr.mu.Unlock()
	cmgr.conns[conn.ID()] = conn
}

func (cmgr *connMgr) Remove(id uint64) {
	cmgr.mu.Lock()
	defer cmgr.mu.Unlock()
	delete(cmgr.conns, id)
}

func (cmgr *connMgr) Load(id uint64) (*WsConnection, error) {
	cmgr.mu.RLock()
	defer cmgr.mu.RUnlock()
	conn, ok := cmgr.conns[id]
	if ok {
		return conn, nil
	}
	return nil, errors.New("connection do not exist")
}

func (cmgr *connMgr) Close() {
	conns := make(map[uint64]*WsConnection)
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

// WsServer represents an websocket server instance
type WsServer struct {
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

// NewWsServer retrun a gate implement with websocket
func NewWsServer(opts ...gate.Option) *WsServer {
	options := gate.Options{
		Address:  DefaultAddress,
		MaxConns: DefaultMaxConns,
	}

	for _, o := range opts {
		o(&options)
	}

	s := &WsServer{
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

func (s *WsServer) wsHandler(w http.ResponseWriter, r *http.Request) {
	wsconn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.options.Logger.Errorf("ws upgrade err %v", err)
		return
	}
	if s.Count() > s.options.MaxConns {
		return
	}
	conn := NewConnection(wsconn, s)
	conn.Start()
}

// Count return connection count
func (s *WsServer) Count() int {
	return s.cmgr.Count()
}

// Start the server
func (s *WsServer) Start() error {
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

// Stop the server
func (s *WsServer) Stop() error {
	s.cmgr.Close()

	if err := s.httpsrv.Shutdown(context.TODO()); err != nil {
		s.options.Logger.Errorf("wsserver shutdown failed: %v", err)
		return err
	}

	s.options.Logger.Infof("wsserver stopped")

	return nil
}
