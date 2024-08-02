package main

import (
	"github.com/gin-gonic/gin"
	// "github.com/wesom/badger"
)

func main() {
	// gate := badger.NewWsGateWay()
	// defer gate.Close()

	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.File("index.html")
	})
	// r.GET("/ws", func(c *gin.Context) {
	// 	gate.ServeHTTP(c.Writer, c.Request, nil)
	// })

	// gate.OnTextMessage(func(c *badger.Connection, data []byte) {
	// 	c.WriteTextMessage(data)
	// })

	r.Run("localhost:8080")
}
