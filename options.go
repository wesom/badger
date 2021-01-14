package badger

import "context"

// Option for service
type Option func(*Options)

// Options for service
type Options struct {
	Name    string
	Version string
	Context context.Context
	Signal  bool
}

func newOptions(opts ...Option) Options {
	opt := Options{
		Name:    DefaultName,
		Version: DefaultVersion,
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
