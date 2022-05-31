package badger

import "go.uber.org/zap"

type Option func(opts *Options)

type Options struct {
	MaxConn int
	Logger  *zap.Logger
	Handler EventHandler
}

func WithMaxConn(maxconn int) Option {
	return func(opts *Options) {
		opts.MaxConn = maxconn
	}
}

func WithLogger(l *zap.Logger) Option {
	return func(opts *Options) {
		opts.Logger = l
	}
}

func WithEventHandler(h EventHandler) Option {
	return func(opts *Options) {
		opts.Handler = h
	}
}
