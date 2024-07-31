package badger

import (
	"net"

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

type onConnectFunc func(connID string, remoteAddr net.Addr) error
type onTextMessageFunc func(connID string, data []byte)
type onBinaryMessageFunc func(connID string, data []byte)
type onErrorFunc func(connID string, e error)
type onDisconnectFunc func(connID string)
