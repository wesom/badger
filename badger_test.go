package badger

import (
	"net/http/httptest"
	"strings"
	"testing"
	"testing/quick"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func NewDialer(url string) (*websocket.Conn, error) {
	dialer := &websocket.Dialer{}
	conn, _, err := dialer.Dial(url, nil)
	return conn, err
}

func makeWsProto(s string) string {
	return "ws" + strings.TrimPrefix(s, "http")
}

func TestEcho(t *testing.T) {
	gw := NewWsGateWay()
	gw.OnMessage(func(c *Connection, data []byte) {
		c.Write(data)
	})
	server := httptest.NewServer(gw)
	defer server.Close()

	f := func(msg string) bool {
		conn, err := NewDialer(makeWsProto(server.URL))
		assert.Nil(t, err)
		defer conn.Close()

		conn.WriteMessage(websocket.TextMessage, []byte(msg))

		_, ret, err := conn.ReadMessage()

		assert.Nil(t, err)

		assert.Equal(t, msg, string(ret))

		return true
	}

	err := quick.Check(f, nil)

	assert.Nil(t, err)

}
