package badger

import (
	"sync"
)

type ConnMgr struct {
	mu    sync.RWMutex
	conns map[uint64]*Connection
}

func NewConnMgr() *ConnMgr {
	return &ConnMgr{
		conns: make(map[uint64]*Connection),
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

func (cmgr *ConnMgr) Find(connID uint64) *Connection {
	cmgr.mu.Lock()
	defer cmgr.mu.Unlock()
	v, ok := cmgr.conns[connID]
	if ok {
		return v
	}
	return nil
}

func (cmgr *ConnMgr) Close() {
	cmgr.mu.Lock()
	defer cmgr.mu.Unlock()
	for _, conn := range cmgr.conns {
		conn.Close()
	}
}
