package gate

// Gate is a gate server abstraction
type Gate interface {
	// Start the server
	Start() error
	// Stop the server
	Stop() error
}

// Options for gate
type Options struct {
	Address string
}

// Option sets values in Options
type Option func(o *Options)

// WithAddress to listen
func WithAddress(addr string) Option {
	return func(o *Options) {
		o.Address = addr
	}
}
