package dispatch

import (
	"errors"
	"sync"
	"sync/atomic"
)

// Dispatcher dispatch message to diffrent handlers
type Dispatcher struct {
	options  Options
	queues   []chan Message
	wg       sync.WaitGroup
	exitChan chan int
	exitFlag int32
	router   *Router
}

// Default Option
var (
	DefaultPartitions = 8
	DefaultQueueCap   = 4096
)

// NewDispatcher return a obj instance
func NewDispatcher(opts ...Option) Dispatch {
	options := Options{
		QueueCap:   DefaultQueueCap,
		Partitions: DefaultPartitions,
	}

	for _, o := range opts {
		o(&options)
	}

	queues := make([]chan Message, options.Partitions)
	for i := 0; i < len(queues); i++ {
		queues[i] = make(chan Message, options.QueueCap)
	}

	d := &Dispatcher{
		options:  options,
		queues:   queues,
		exitChan: make(chan int),
		router:   newRouter(),
	}
	return d
}

func partition(key uint64, size int) int {
	return int(key % uint64(size))
}

// Handle register a handler to router
func (d *Dispatcher) Handle(protoname string, handler Handler) {
	d.router.register(protoname, handler)
}

// Delivery a task
func (d *Dispatcher) Delivery(msg Message) error {
	indexPartition := partition(msg.Key(), len(d.queues))

	if atomic.LoadInt32(&d.exitFlag) == 1 {
		return errors.New("exiting")
	}

	d.queues[indexPartition] <- msg

	return nil
}

// Start Workers
func (d *Dispatcher) Start() error {
	d.options.Logger.Infof("dipatcher start")

	for i := 0; i < len(d.queues); i++ {
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

	for i := 0; i < len(d.queues); i++ {
		close(d.queues[i])
	}

	d.options.Logger.Infof("dipatcher stopped")

	return nil
}

func (d *Dispatcher) doHandle(msg Message) {
	handler := d.router.fetch(msg.Name())
	if handler != nil {
		handler.Handle(msg)
	}
}

func (d *Dispatcher) handlePump(i int) {
	defer d.wg.Done()

	queue := d.queues[i]

quitLoop:
	for {
		select {
		case msg := <-queue:
			d.doHandle(msg)
			d.options.Logger.Debugf("queue[%d] running key %d", i, msg.Key())
		case <-d.exitChan:
			break quitLoop
		}
	}

	d.options.Logger.Debugf("quit handlePump [%d], abandon message: %d", i, len(queue))
}
