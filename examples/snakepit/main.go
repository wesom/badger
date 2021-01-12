package main

import (
	"log"

	"github.com/wesom/badger"
)

func main() {
	service := badger.NewService()
	service.Init(
		badger.WithName("snakepit"),
	)
	log.Printf("service %s start ...", service.String())

	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}
