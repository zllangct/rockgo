package Network

import (
	"errors"
	"fmt"
	"github.com/zllangct/RockGO/RockInterface"
	"github.com/zllangct/RockGO/logger"
	"sync"
)

type SessionMgr struct {
	connections map[uint32]RockInterface.ISession
	conMrgLock  sync.RWMutex
}

func (this *SessionMgr) Add(conn RockInterface.ISession) {
	this.conMrgLock.Lock()
	defer this.conMrgLock.Unlock()
	this.connections[conn.GetSessionId()] = conn
	logger.Debug(fmt.Sprintf("Total connection: %d", len(this.connections)))
}

func (this *SessionMgr) Remove(conn RockInterface.ISession) error {
	this.conMrgLock.Lock()
	defer this.conMrgLock.Unlock()
	_, ok := this.connections[conn.GetSessionId()]
	if ok {
		delete(this.connections, conn.GetSessionId())
		logger.Info(len(this.connections))
		return nil
	} else {
		return errors.New("not found!!")
	}

}

func (this *SessionMgr) Get(sid uint32) (RockInterface.ISession, error) {
	this.conMrgLock.Lock()
	defer this.conMrgLock.Unlock()
	v, ok := this.connections[sid]
	if ok {
		delete(this.connections, sid)
		return v, nil
	} else {
		return nil, errors.New("not found!!")
	}
}

func (this *SessionMgr) Len() int {
	this.conMrgLock.Lock()
	defer this.conMrgLock.Unlock()
	return len(this.connections)
}

func NewConnectionMgr() *SessionMgr {
	return &SessionMgr{
		connections: make(map[uint32]RockInterface.ISession),
	}
}
