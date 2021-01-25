package dispatch

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/wesom/badger/log"
)

// Dispatcher dispatch message to diffrent handlers
type Dispatcher struct {
	size       int
	partitions int
	queues     []chan Request
	wg         sync.WaitGroup
	exitChan   chan int
	exitFlag   int32
	router     *Router
	logger     log.Logger
}

// Default Option
var (
	DefaultPartitions = 8
	DefaultQueueSize  = 4096
)

// NewDispatcher return a obj instance
func NewDispatcher() Dispatch {
	d := &Dispatcher{
		size:       DefaultQueueSize,
		partitions: DefaultPartitions,
		queues:     make([]chan Request, DefaultPartitions),
		exitChan:   make(chan int),
		router:     newRouter(),
		logger:     log.DefaultLogger,
	}
	return d
}

func partition(key int, size int) int {
	return key % size
}

// Handle register a handler to router
func (d *Dispatcher) Handle(protoname string, handler Handler) {
	d.router.register(protoname, handler)
}

// Delivery a task
func (d *Dispatcher) Delivery(req Request) error {
	indexPartition := partition(req.Key(), d.partitions)

	if atomic.LoadInt32(&d.exitFlag) == 1 {
		return errors.New("exiting")
	}

	d.queues[indexPartition] <- req

	return nil
}

// Start Workers
func (d *Dispatcher) Start() error {
	for i := 0; i < d.partitions; i++ {
		d.queues[i] = make(chan Request, d.size)
		d.wg.Add(1)
		go d.handlePump(i)
	}
	return nil
}

// Stop Workers
func (d *Dispatcher) Stop() error {
	if !atomic.CompareAndSwapInt32(&d.exitFlag, 0, 1) {
		return errors.New("close exitFlag")
	}

	close(d.exitChan)

	d.wg.Wait()

	for i := 0; i < d.partitions; i++ {
		close(d.queues[i])
	}
	return nil
}

func (d *Dispatcher) doHandle(req Request) {
	handler := d.router.fetch(req.Name())
	if handler != nil {
		handler.Handle(req)
	}
}

func (d *Dispatcher) handlePump(i int) {
	defer d.wg.Done()

	queue := d.queues[i]

quitLoop:
	for {
		select {
		case req := <-queue:
			d.doHandle(req)
			d.logger.Debugf("queue[%d] running key %d", i, req.Key())
		case <-d.exitChan:
			break quitLoop
		}
	}

	d.logger.Infof("quit handlePump [%d], abandon request: %d", i, len(queue))
}
