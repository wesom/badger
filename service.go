package badger

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/wesom/badger/gate"

	"github.com/wesom/badger/logging"
)

// Default Option
var (
	DefaultName    = "badger-service"
	DefaultVersion = "latest"
)

type service struct {
	opts Options
	gate gate.Gate

	once sync.Once
}

func newService(opts ...Option) Service {
	options := newOptions(opts...)
	svc := new(service)
	svc.opts = options
	svc.gate = gate.NewGate()

	return svc
}

func (s *service) Init(opts ...Option) {
	for _, o := range opts {
		o(&s.opts)
	}
}

func (s *service) String() string {
	return s.opts.Name + ":" + s.opts.Version
}

func (s *service) Options() Options {
	return s.opts
}

func (s *service) Start() error {
	if err := s.gate.Start(); err != nil {
		return err
	}
	return nil
}

func (s *service) Stop() error {
	if err := s.gate.Stop(); err != nil {
		return err
	}
	return nil
}

func (s *service) Run() error {
	if err := s.Start(); err != nil {
		return err
	}

	ch := make(chan os.Signal, 1)
	if s.opts.Signal {
		signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL)
	}

	select {
	case sig := <-ch:
		logging.Logger().Infof("Receive signal %s", sig)
	case <-s.opts.Context.Done():
		logging.Logger().Info("Receive context done")
	}

	return s.Stop()
}
