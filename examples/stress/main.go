package main

import (
	"fmt"
	"time"

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

	// gate.OnMessage(func(c *badger.Connection, data []byte) {
	// })

	done := make(chan bool)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ticker.C:
				fmt.Printf("%s gate len: %d\n", time.Now().Format("2006-01-02 15:04:05"), gate.Len())
			case <-done:
				return
			}

		}
	}()

	r.Run(":8080")

	done <- true
}
