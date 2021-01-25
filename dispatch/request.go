package dispatch

// Request Wrapper message data
type Request interface {
	Key() int

	Name() string
}

// A Handler process a message request
type Handler interface {
	Handle(req Request)
}

// Dispatch privide method abstraction
type Dispatch interface {
	// Start the server
	Start() error
	// Stop the server
	Stop() error
	// Delivery a request
	Delivery(req Request) error
}
