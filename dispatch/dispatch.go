package dispatch

// Message Wrapper message data
type Message interface {
	Key() uint64

	Name() string
}

// A Handler process a message request
type Handler interface {
	Handle(msg Message)
}

// Dispatch privide method abstraction
type Dispatch interface {
	// Start the server
	Start() error
	// Stop the server
	Stop() error
	// Delivery a request
	Delivery(msg Message) error
}
