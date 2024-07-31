package badger

type WsEvent struct {
	connID string
	data   interface{}
	err    error
}

type WsEventType int

const (
	WsConnect WsEventType = iota
	WsDisconnect
	WsTextMessage
	WsBinaryMessage
	WsError
)
