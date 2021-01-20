package websocket

import (
	"context"
	"net"
	"net/http"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/wesom/badger/gate"
	"github.com/wesom/badger/logging"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:    1024,
	WriteBufferSize:   1024,
	EnableCompression: true,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func wsHandler(s *wsServer, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logging.Logger().Errorf("ws upgrade err %v", err)
		return
	}
	connection := NewConnection(s.NextID(), conn)
	connection.Start()
}

type wsServer struct {
	options gate.Options
	nextid  uint64
	httpsrv *http.Server
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
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		wsHandler(s, w, r)
	})

	s.httpsrv = &http.Server{
		Handler: mux,
	}

	return s
}

func (s *wsServer) NextID() uint64 {
	return atomic.AddUint64(&s.nextid, 1)
}

func (s *wsServer) Start() error {
	ln, err := net.Listen("tcp", s.options.Address)
	if err != nil {
		return err
	}

	logging.Logger().Infof("wsserver listening on %s", ln.Addr().String())

	go func() {
		if err := s.httpsrv.Serve(ln); err != http.ErrServerClosed {
			logging.Logger().Errorf("wsserver serve failed: %v", err)
		}
	}()

	return nil
}

func (s *wsServer) Stop() error {
	if err := s.httpsrv.Shutdown(context.TODO()); err != nil {
		logging.Logger().Errorf("wsserver shutdown failed: %v", err)
		return err
	}

	logging.Logger().Info("wsserver stopped")

	return nil
}
