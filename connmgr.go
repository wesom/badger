package badger

import (
	"sync"
)

type ConnMgr struct {
	mu    sync.RWMutex
	conns map[Conn]struct{}
}

func NewConnMgr() *ConnMgr {
	return &ConnMgr{
		conns: make(map[Conn]struct{}),
	}
}

func (cmgr *ConnMgr) Count() int {
	cmgr.mu.RLock()
	defer cmgr.mu.RUnlock()
	return len(cmgr.conns)
}

func (cmgr *ConnMgr) Add(conn Conn) {
	cmgr.mu.Lock()
	defer cmgr.mu.Unlock()
	cmgr.conns[conn] = struct{}{}
}

func (cmgr *ConnMgr) Remove(conn Conn) {
	cmgr.mu.Lock()
	defer cmgr.mu.Unlock()
	delete(cmgr.conns, conn)
}

func (cmgr *ConnMgr) Close() {
	cmgr.mu.Lock()
	defer cmgr.mu.Unlock()
	for conn := range cmgr.conns {
		conn.Close()
	}
}
