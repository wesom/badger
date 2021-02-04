package dispatch

import (
	"testing"
	"time"
)

type SimpleRequest struct {
	ID uint64
}

func (s *SimpleRequest) Key() uint64 {
	return s.ID
}

func (s *SimpleRequest) Conn() Conn {
	return nil
}

func (s *SimpleRequest) Name() string {
	return "simple"
}

func TestPump(t *testing.T) {
	d := NewDispatcher()
	d.Start()

	for i := 0; i < 1000; i++ {
		simpleRequest := &SimpleRequest{
			ID: uint64(i),
		}
		d.Put(simpleRequest)
	}

	time.Sleep(time.Second * 10)

	d.Stop()
}
