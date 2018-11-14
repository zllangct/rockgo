package ConnPool

import (
	"errors"
	"github.com/zllangct/RockGO/RockInterface"
)

var (
	//ErrClosed 连接池已经关闭Error
	ErrClosed = errors.New("ConnPool is closed")
)

//Pool 基本方法
type Pool interface {
	Get() (RockInterface.IWriter, error)

	Put(RockInterface.IWriter) error

	Close(RockInterface.IWriter) error

	Release()

	Len() int
}
