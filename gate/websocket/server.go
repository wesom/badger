package websocket

import (
	"context"
	"net"
	"net/http"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/wesom/badger/gate"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:    1024,
	WriteBufferSize:   1024,
	EnableCompression: true,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type wsServer struct {
	options    gate.Options
	totalConns uint64
	httpsrv    *http.Server
}

var (
	// DefaultAddress to ws
	DefaultAddress = ":8000"
)

// NewServer retrun a gate implement with websocket
func NewServer(opts ...gate.Option) gate.Gate {
	options := gate.Options{
		Address: DefaultAddress,
	}

	for _, o := range opts {
		o(&options)
	}

	s := &wsServer{
		options: options,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.wsHandler)

	s.httpsrv = &http.Server{
		Handler: mux,
	}

	return s
}

func (s *wsServer) wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// logger.Errorf("ws upgrade err %v", err)
		return
	}
	connection := NewConnection(conn)
	connection.Start()
	atomic.AddUint64(&s.totalConns, 1)
}

func (s *wsServer) ConnCount() uint64 {
	return s.totalConns
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
