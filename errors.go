package badger

import "errors"

var (
	ErrWriteClosed = errors.New("connection is closed while writing")
	ErrBufferFull  = errors.New("write message buffer is full")
	ErrConnClosed  = errors.New("connection is closed already")
)
