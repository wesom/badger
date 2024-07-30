package badger

import (
	"net"

	"go.uber.org/zap"
)

type Option func(opts *Options)

type Options struct {
	Logger *zap.Logger
}

func WithLogger(l *zap.Logger) Option {
	return func(opts *Options) {
		opts.Logger = l
	}
}

type onConnectFunc func(connID uint64, remoteAddr net.Addr) error
type onTextMessageFunc func(connID uint64, data []byte)
type onBinaryMessageFunc func(connID uint64, data []byte)
type onErrorFunc func(connID uint64, e error)
type onDisconnectFunc func(connID uint64)