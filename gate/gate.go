package gate

// Gate is a gateway abstraction
type Gate interface {
	// Start the server
	Start() error
	// Stop the server
	Stop() error
}

// NewGate return a gate
func NewGate(opts ...Option) Gate {
	return newWSGate(opts...)
}
