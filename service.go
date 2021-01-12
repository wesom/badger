package badger

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/wesom/badger/logging"
)

// Default Option
var (
	DefaultName    = "badger-service"
	DefaultVersion = "latest"
)

type service struct {
	opts Options

	once sync.Once
}

func newService(opts ...Option) Service {
	options := newOptions(opts...)
	service := new(service)
	service.opts = options

	return service
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
	return nil
}

func (s *service) Stop() error {
	return nil
}

func (s *service) Run() error {
	if err := s.Start(); err != nil {
		return err
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGPIPE)

	for {
		select {
		case sig := <-ch:
			switch sig {
			// ignore SIGPIPE
			case syscall.SIGPIPE:
			default:
				logging.Logger().Infof("Receive [signal] %s", sig)
				goto stop
			}
		}
	}

stop:
	return s.Stop()
}
