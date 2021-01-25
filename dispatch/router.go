package dispatch

// A Router mapping message to diffrent handlers
type Router struct {
	m map[string]Handler
}

func newRouter() *Router {
	return &Router{
		m: make(map[string]Handler),
	}
}

// fetch return a handler with the name
func (r *Router) fetch(name string) Handler {
	if handler, ok := r.m[name]; ok {
		return handler
	}
	return nil
}

// register a handler to router
func (r *Router) register(name string, handler Handler) {
	if _, exist := r.m[name]; exist {
		panic("router: multiple registrations for " + name)
	}
	r.m[name] = handler
}
