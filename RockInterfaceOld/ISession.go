package RockInterface

import (
	"net"
)

type ISession interface {
	Start()
	Stop()
	GetConnection() net.Conn
	GetSessionId() uint32
	Send([]byte) error
	SendBuff([]byte) error
	RemoteAddr() net.Addr
	LostConnection()
	GetProperty(string) (interface{}, error)
	SetProperty(string, interface{})
	RemoveProperty(string)
}
