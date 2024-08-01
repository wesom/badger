package badger

import (
	"go.uber.org/zap"
)

type Option func(opts *Options)

type Options struct {
	Logger   *zap.Logger
	MaxConns int
}

func WithLogger(l *zap.Logger) Option {
	return func(opts *Options) {
		opts.Logger = l
	}
}

func WithMaxConns(size int) Option {
	return func(opts *Options) {
		opts.MaxConns = size
	}
}
