package badger

import "net"

type Conn interface {
	ConnID() uint64

	Write(buf []byte)

	Close()

	// addr
	LocalAddr() net.Addr

	RemoteAddr() net.Addr

	// property
	SetProperty(k string, v interface{})

	GetProperty(k string) (interface{}, error)

	RemoveProperty(key string)

	Int64(k string) int64
}

type EventHandler interface {
	OnConnect(c Conn)

	OnMessage(c Conn, data []byte)

	OnDisconnect(c Conn)
}

type EventServer struct{}

func (es *EventServer) OnConnect(Conn) {
}

func (es *EventServer) OnMessage(Conn, []byte) {
}

func (es *EventServer) OnDisconnect(Conn) {
}
