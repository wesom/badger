package badger

import (
	"sync"
)

type ConnMgr struct {
	mu    sync.RWMutex
	conns map[uint64]*Connection
}

func NewConnMgr(maxConn int) *ConnMgr {
	return &ConnMgr{
		conns: make(map[uint64]*Connection, maxConn),
	}
}

func (cmgr *ConnMgr) Add(conn *Connection) {
	cmgr.mu.Lock()
	defer cmgr.mu.Unlock()
	cmgr.conns[conn.ConnID()] = conn
}

func (cmgr *ConnMgr) Remove(conn *Connection) {
	cmgr.mu.Lock()
	defer cmgr.mu.Unlock()
	delete(cmgr.conns, conn.ConnID())
}

func (cmgr *ConnMgr) Apply(connID uint64, f func(*Connection)) {
	cmgr.mu.RLock()
	c, ok := cmgr.conns[connID]
	cmgr.mu.RUnlock()
	// c is thread safe
	if ok {
		f(c)
	}
}

func (cmgr *ConnMgr) Close() {
	cmgr.mu.RLock()
	defer cmgr.mu.RUnlock()
	for _, conn := range cmgr.conns {
		conn.Close()
	}
}
