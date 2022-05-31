package badger

import (
	"net"

	"github.com/gorilla/websocket"
)

// WsConnection represents a websocket connect wrapper
type WsConnection struct {
	conn *websocket.Conn
}

func (c *WsConnection) ReadMessage() ([]byte, error) {
	_, p, err := c.conn.ReadMessage()
	return p, err
}

func (c *WsConnection) WriteMessage(data []byte) error {
	return c.conn.WriteMessage(websocket.BinaryMessage, data)
}

func (c *WsConnection) Close() error {
	if tcpconn, ok := c.conn.UnderlyingConn().(*net.TCPConn); ok {
		// avoid timewait
		tcpconn.SetLinger(0)
	}
	return c.conn.Close()
}

func (c *WsConnection) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *WsConnection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}
