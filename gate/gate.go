package gate

// Gate is a gate server abstraction
type Gate interface {
	// Start the server
	Start() error
	// Stop the server
	Stop() error
}

// UIDFunc is a callback to generator id
type UIDFunc func() uint64

// Options for gate
type Options struct {
	UIDFn    UIDFunc
	Address  string
	MaxConns int
}

// Option sets values in Options
type Option func(o *Options)

// WithUIDFunc to generate id
func WithUIDFunc(fn UIDFunc) Option {
	return func(o *Options) {
		o.UIDFn = fn
	}
}

// WithAddress to listen
func WithAddress(addr string) Option {
	return func(o *Options) {
		o.Address = addr
	}
}

// WithMaxConns to limit connections
func WithMaxConns(max int) Option {
	return func(o *Options) {
		o.MaxConns = max
	}
}
