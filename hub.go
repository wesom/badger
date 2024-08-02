package badger

import (
	"sync"
)

type hub struct {
	mu    sync.RWMutex
	conns map[*Connection]struct{}
}

func newHub() *hub {
	return &hub{
		conns: make(map[*Connection]struct{}),
	}
}

func (h *hub) add(conn *Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.conns[conn] = struct{}{}
}

func (h *hub) del(conn *Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.conns, conn)
}

func (h *hub) len() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.conns)
}

func (h *hub) each(f func(*Connection)) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for conn := range h.conns {
		f(conn)
	}
}
