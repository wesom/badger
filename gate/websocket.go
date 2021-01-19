package gate

import (
	"context"
	"net"
	"net/http"

	"github.com/wesom/badger/logging"
)

func wsHandler(w http.ResponseWriter, req *http.Request) {

}

type wsServer struct {
	options Options
	httpsrv *http.Server
}

var (
	// DefaultAddress to ws
	DefaultAddress = ":8000"
)

func newWSGate(opts ...Option) *wsServer {
	options := Options{
		Address: DefaultAddress,
	}

	for _, o := range opts {
		o(&options)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", wsHandler)

	s := &wsServer{
		options: options,
		httpsrv: &http.Server{
			Handler: mux,
		},
	}

	return s
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
