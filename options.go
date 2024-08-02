package badger

import (
	"go.uber.org/zap"
)

type Option func(opts *Options)

type Options struct {
	Logger           *zap.Logger
	OutputBufferSize int
}

func WithLogger(l *zap.Logger) Option {
	return func(opts *Options) {
		opts.Logger = l
	}
}

func WithOutputBufferSize(size int) Option {
	return func(opts *Options) {
		opts.OutputBufferSize = size
	}
}
