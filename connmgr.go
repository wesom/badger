package badger

import (
	"sync"
)

type connMgr struct {
	mu    sync.RWMutex
	conns map[uint64]*Connection
}

func newConnMgr() *connMgr {
	return &connMgr{
		conns: make(map[uint64]*Connection),
	}
}

func (cmgr *connMgr) add(conn *Connection) {
	cmgr.mu.Lock()
	defer cmgr.mu.Unlock()
	cmgr.conns[conn.ConnID()] = conn
}

func (cmgr *connMgr) remove(conn *Connection) {
	cmgr.mu.Lock()
	defer cmgr.mu.Unlock()
	delete(cmgr.conns, conn.ConnID())
}

func (cmgr *connMgr) close() {
	cmgr.mu.RLock()
	defer cmgr.mu.RUnlock()
	for _, conn := range cmgr.conns {
		conn.Close()
	}
}
