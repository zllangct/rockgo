package cluster

import (
	"errors"
	"fmt"
	"github.com/zllangct/RockGO/RockInterface"
	"github.com/zllangct/RockGO/logger"
	"math/rand"
	"strings"
	"sync"
	"github.com/zllangct/RockGO/ConnPool"
)

type Child struct {
	name string
	rpc  *XingoRpc
}

func NewChild(name string, conn RockInterface.IWriter) *Child {
	return &Child{
		name: name,
		rpc:  NewXingoRpc(conn),
	}
}

func (this *Child)InitPool()  {
	this.rpc.InitPool()
}

func (this *Child) Close() {
	this.rpc.Close()
}

func (this *Child) GetName() string {
	return this.name
}

/*
添加主动conn ConnPool
 */

func (this *Child)AddWorkPool(pool ConnPool.Pool)error  {
	this.rpc.workPool=pool
	return  nil
}

/*
添加被动conn ConnPool
 */
func (this *Child)AddWorkConn(conn RockInterface.IWriter) error {
	remote, err:=conn.GetProperty("remote")
	if err != nil{
		return  err
	}
	logger.Info(fmt.Sprintf("Add one workconn from %s successed.", remote))
	this.rpc.workConn<-conn
	return nil
}

/*
归还链接到链接池
*/
func (this *Child)ConnBack(conn RockInterface.IWriter)error{
	t,err:=conn.GetProperty("type")
	if err != nil{
		return err
	}
	switch t {
	case 1:
		//主要连接,不用做处理
	case 2:
		//主动工作连接
		return this.rpc.workPool.Put(conn)
	case 3:
		//被动链接
		this.rpc.ConnBack(conn)
	}
	return  nil
}

func (this *Child) CallChildNotForResult(target string, args ...interface{}) error {
	return this.rpc.CallRpcNotForResult(target, args...)
}

func (this *Child) CallNotForResArray(target string, args []interface{}) error {
	return this.rpc.CallRpcNotForResultArray(target, args)
}

func (this *Child) CallChildForResult(target string, args ...interface{}) (*RpcData, error) {
	return this.rpc.CallRpcForResult(target, args...)
}

func (this *Child) CallForResArray(target string, args []interface{}) (*RpcData, error) {
	return this.rpc.CallRpcForResultArray(target, args)
}

type ChildMgr struct {
	childs map[string]*Child
	sync.RWMutex
}

func NewChildMgr() *ChildMgr {
	return &ChildMgr{
		childs: make(map[string]*Child, 0),
	}
}

func (this *ChildMgr) AddChild(name string, conn RockInterface.IWriter) {
	this.Lock()
	defer this.Unlock()

	this.childs[name] = NewChild(name, conn)
	logger.Debug(fmt.Sprintf("child %s connected.", name))
}

func (this *ChildMgr) RemoveChild(name string) {
	this.Lock()
	defer this.Unlock()

	delete(this.childs, name)
	logger.Debug(fmt.Sprintf("child %s lostconnection.", name))
}

func (this *ChildMgr) GetChild(name string) (*Child, error) {
	this.RLock()
	defer this.RUnlock()

	child, ok := this.childs[name]
	if ok {
		return child, nil
	} else {
		return nil, errors.New(fmt.Sprintf("no child named %s", name))
	}
}

func (this *ChildMgr) GetChildsByPrefix(namePrefix string) []*Child {
	this.RLock()
	defer this.RUnlock()

	childs := make([]*Child, 0)
	for k, v := range this.childs {
		if strings.HasPrefix(k, namePrefix) {
			childs = append(childs, v)
		}
	}
	return childs
}

func (this *ChildMgr) GetChilds() []*Child {
	this.RLock()
	defer this.RUnlock()

	childs := make([]*Child, 0)
	for _, v := range this.childs {
		childs = append(childs, v)
	}
	return childs
}

func (this *ChildMgr) GetRandomChild(namesuffix string) *Child {
	childs := make([]*Child, 0)
	if namesuffix != "" {
		//一类
		childs = this.GetChildsByPrefix(namesuffix)
	} else {
		//所有
		childs = this.GetChilds()
	}
	if len(childs) > 0 {
		pos := rand.Intn(len(childs))
		return childs[pos]
	}
	return nil
}
