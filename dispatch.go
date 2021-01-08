package badger

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/wesom/badger/logging"
)

// Dispatcher dispatch message to diffrent handlers
type Dispatcher struct {
	Size       int
	Partitions int
	Queues     []chan Request
	wg         sync.WaitGroup
	exitChan   chan int
	exitFlag   int32
}

// NewDispatcher return a obj instance
func NewDispatcher(parts, size int) *Dispatcher {
	if parts < 1 {
		panic("partitions must gather than zero")
	}
	d := &Dispatcher{
		Size:       size,
		Partitions: parts,
		Queues:     make([]chan Request, parts),
		exitChan:   make(chan int),
	}
	return d
}

func partition(key int, size int) int {
	return key % size
}

// Put task
func (d *Dispatcher) Put(req Request) error {
	indexPartition := partition(req.Key(), d.Partitions)

	if atomic.LoadInt32(&d.exitFlag) == 1 {
		return errors.New("exiting")
	}

	d.Queues[indexPartition] <- req

	return nil
}

// Start Workers
func (d *Dispatcher) Start() {
	for i := 0; i < d.Partitions; i++ {
		d.Queues[i] = make(chan Request, d.Size)
		d.wg.Add(1)
		go d.handlePump(i)
	}
}

// Close Workers
func (d *Dispatcher) Close() error {
	if !atomic.CompareAndSwapInt32(&d.exitFlag, 0, 1) {
		return errors.New("close exitFlag")
	}

	close(d.exitChan)

	d.wg.Wait()

	for i := 0; i < d.Partitions; i++ {
		close(d.Queues[i])
	}
	return nil
}

func (d *Dispatcher) handlePump(i int) {
	defer d.wg.Done()

	queue := d.Queues[i]
	for {
		select {
		case req := <-queue:
			logging.Logger().Debugf("queue[%d] running key %d", i, req.Key())
		case <-d.exitChan:
			goto exit
		}
	}

exit:
	logging.Logger().Infof("quit handlePump [%d], abandon request: %d", i, len(queue))
}
