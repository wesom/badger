package dispatch

import (
	"github.com/wesom/badger/log"
)

// Message Wrapper message data
type Message interface {
	Key() uint64

	Name() string
}

// A Handler process a message request
type Handler interface {
	Handle(msg Message)
}

// Dispatch privide method abstraction
type Dispatch interface {
	// Start the server
	Start() error
	// Stop the server
	Stop() error
	// Delivery a request
	Delivery(msg Message) error
}

// Options for gate
type Options struct {
	Logger     log.Logger
	QueueCap   int
	Partitions int
}

// Option sets values in Options
type Option func(o *Options)

// WithLogger set logger
func WithLogger(l log.Logger) Option {
	return func(o *Options) {
		o.Logger = l
	}
}

// WithQueueCap set queue capacity
func WithQueueCap(c int) Option {
	return func(o *Options) {
		o.QueueCap = c
	}
}

// WithPartitions set queue partitions
func WithPartitions(p int) Option {
	return func(o *Options) {
		o.Partitions = p
	}
}
