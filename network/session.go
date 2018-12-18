package network

import (
	"sync"
)

type Conn interface {
	WriteMessage(messageType uint32, data []byte) error
	Addr() string
	Close() error
}

type Session struct {
	locker sync.RWMutex
	ID     		string
	properties  map[string]interface{}
	conn        Conn
}

func (this *Session)RemoteAddr() string {
	return this.conn.Addr()
}

func (this *Session)Close() error {
	this.locker.Lock()
	defer this.locker.Unlock()
	this.properties = nil
	return this.conn.Close()
}

func (this *Session)SetProperty(key string,value interface{})  {
	this.locker.Lock()
	this.properties[key]=value
	this.locker.Unlock()
}

func (this *Session)GetProperty(key string) (interface{},bool) {
	this.locker.RLock()
	defer this.locker.RUnlock()
	p,ok:= this.properties[key]
	return p,ok
}

func (this *Session)Emit(messageType uint32,message []byte) error {
	return this.conn.WriteMessage(messageType,message)
}


