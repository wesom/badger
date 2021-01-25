package dispatch

import (
	"testing"
	"time"
)

type SimpleRequest struct {
	ID int
}

func (s *SimpleRequest) Key() int {
	return s.ID
}

func (s *SimpleRequest) Name() string {
	return "simple"
}

func TestPump(t *testing.T) {
	d := NewDispatcher()
	d.Start()

	for i := 0; i < 1000; i++ {
		simpleRequest := &SimpleRequest{
			ID: i,
		}
		d.Delivery(simpleRequest)
	}

	time.Sleep(time.Second * 10)

	d.Stop()
}
