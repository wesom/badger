package badger

type Option func(opts *Options)

type Options struct {
	OutputBufferSize int
}

func WithOutputBufferSize(size int) Option {
	return func(opts *Options) {
		opts.OutputBufferSize = size
	}
}
