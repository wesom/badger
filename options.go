package badger

import (
	"context"

	"github.com/wesom/badger/log"
)

// Option for service
type Option func(*Options)

// Options for service
type Options struct {
	Name    string
	Version string
	Logger  log.Logger
	Context context.Context
	Signal  bool
}

// Default Option
var (
	DefaultName    = "badger-service"
	DefaultVersion = "latest"
)

func newOptions(opts ...Option) Options {
	opt := Options{
		Name:    DefaultName,
		Version: DefaultVersion,
		Logger:  log.DefaultLogger,
		Context: context.Background(),
		Signal:  true,
	}

	for _, o := range opts {
		o(&opt)
	}

	return opt
}

// WithName sets the name
func WithName(n string) Option {
	return func(o *Options) {
		o.Name = n
	}
}

// WithVersion sets the version
func WithVersion(v string) Option {
	return func(o *Options) {
		o.Version = v
	}
}

// WithLogger sets a context for service
func WithLogger(l log.Logger) Option {
	return func(o *Options) {
		o.Logger = l
	}
}

// WithContext sets a context for service
func WithContext(ctx context.Context) Option {
	return func(o *Options) {
		o.Context = ctx
	}
}

// WithSignal enable service handle signal
func WithSignal(b bool) Option {
	return func(o *Options) {
		o.Signal = b
	}
}
