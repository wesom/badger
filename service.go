package badger

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/wesom/badger/dispatch"
	"github.com/wesom/badger/gate"
	"github.com/wesom/badger/gate/websocket"
	"github.com/wesom/badger/util/uid"
)

type service struct {
	opts     Options
	idgen    *uid.ID
	gate     gate.Gate
	dispatch dispatch.Dispatch
}

func newService(opts ...Option) Service {
	options := newOptions(opts...)
	svc := new(service)
	svc.opts = options

	svc.idgen = uid.New()
	svc.dispatch = dispatch.NewDispatcher(
		dispatch.WithLogger(options.Logger),
	)
	svc.gate = websocket.NewGate(
		gate.WithUIDFunc(func() uint64 {
			return svc.idgen.Next()
		}),
		gate.WithDisp(svc.dispatch),
		gate.WithLogger(options.Logger),
	)

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
	s.opts.Logger.Infof("service start")

	if err := s.dispatch.Start(); err != nil {
		return err
	}

	if err := s.gate.Start(); err != nil {
		return err
	}
	return nil
}

func (s *service) Stop() error {
	if err := s.gate.Stop(); err != nil {
		return err
	}

	if err := s.dispatch.Stop(); err != nil {
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
		s.opts.Logger.Infof("received signal %s", sig)
	case <-s.opts.Context.Done():
		s.opts.Logger.Infof("received context done")
	}

	return s.Stop()
}
