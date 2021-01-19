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

func TestPump(t *testing.T) {
	d := NewDispatch(10, 1000)
	d.Start()

	for i := 0; i < 1000; i++ {
		simpleRequest := &SimpleRequest{
			ID: i,
		}
		d.Put(simpleRequest)
	}

	time.Sleep(time.Second * 10)

	d.Stop()
}
