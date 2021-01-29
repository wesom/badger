package gate

import (
	"github.com/wesom/badger/log"

	"github.com/wesom/badger/dispatch"
)

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
	Logger   log.Logger
	UIDFn    UIDFunc
	Disp     dispatch.Dispatch
	Address  string
	MaxConns int
}

// Option sets values in Options
type Option func(o *Options)

// WithLogger set logger
func WithLogger(l log.Logger) Option {
	return func(o *Options) {
		o.Logger = l
	}
}

// WithUIDFunc to generate id
func WithUIDFunc(fn UIDFunc) Option {
	return func(o *Options) {
		o.UIDFn = fn
	}
}

// WithDisp to publish message
func WithDisp(d dispatch.Dispatch) Option {
	return func(o *Options) {
		o.Disp = d
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
