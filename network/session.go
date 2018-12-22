package network

import (
	"errors"
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
	postProcessing []func(sess *Session)
}

func (this *Session)AddPostProcessing(fn func(sess *Session))  {
	this.locker.Lock()
	defer this.locker.Unlock()

	this.postProcessing= append(this.postProcessing, fn)
}

func (this *Session)PostProcessing()  {
	this.locker.Lock()
	defer this.locker.Unlock()

	for _, fn := range this.postProcessing {
		fn(this)
	}
}

func (this *Session)RemoteAddr() string {
	this.locker.RLock()
	defer this.locker.RUnlock()

	return this.conn.Addr()
}

func (this *Session)Close() error {
	this.locker.Lock()
	defer this.locker.Unlock()

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

var ErrSessionDisconnected =errors.New("this session is broken")
func (this *Session)Emit(messageType uint32,message []byte) error {
	if this.conn==nil{
		return ErrSessionDisconnected
	}
	return this.conn.WriteMessage(messageType,message)
}


