package main

import (
	"github.com/gin-gonic/gin"
	"github.com/wesom/badger"
)

func main() {
	gate := badger.NewWsGateWay()
	defer gate.Close()

	r := gin.Default()

	r.GET("/ws", func(c *gin.Context) {
		gate.ServeHTTP(c.Writer, c.Request)
	})

	gate.OnTextMessage(func(c *badger.Connection, data []byte) {
		c.WriteTextMessage(data)
	})

	gate.OnBinaryMessage(func(c *badger.Connection, data []byte) {
		c.WriteBinaryMessage(data)
	})

	r.Run("localhost:8080")
}
