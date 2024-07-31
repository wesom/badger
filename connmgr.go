package badger

import (
	"sync"
)

type connMgr struct {
	mu    sync.RWMutex
	conns map[string]*connection
}

func newConnMgr(initSize int) *connMgr {
	return &connMgr{
		conns: make(map[string]*connection, initSize),
	}
}

func (cmgr *connMgr) add(conn *connection) {
	cmgr.mu.Lock()
	defer cmgr.mu.Unlock()
	cmgr.conns[conn.connID()] = conn
}

func (cmgr *connMgr) remove(conn *connection) {
	cmgr.mu.Lock()
	defer cmgr.mu.Unlock()
	delete(cmgr.conns, conn.connID())
}

func (cmgr *connMgr) apply(connID string, f func(*connection)) {
	cmgr.mu.RLock()
	c, ok := cmgr.conns[connID]
	cmgr.mu.RUnlock()
	// c is thread safe
	if ok {
		f(c)
	}
}

func (cmgr *connMgr) close() {
	cmgr.mu.RLock()
	defer cmgr.mu.RUnlock()
	for _, conn := range cmgr.conns {
		conn.writeClose()
	}
}
