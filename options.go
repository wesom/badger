package badger

// Option for service
type Option func(*Options)

// Options for service
type Options struct {
	Name    string
	Version string
}

func newOptions(opts ...Option) Options {
	opt := Options{
		Name:    DefaultName,
		Version: DefaultVersion,
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
