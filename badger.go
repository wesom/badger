package badger

// Service interface
type Service interface {
	// Init options
	Init(...Option)
	// Options returns the current options
	Options() Options
	// Run the service
	Run() error
	// String of service
	String() string
}

// NewService create a new service
func NewService(opts ...Option) Service {
	return newService(opts...)
}
