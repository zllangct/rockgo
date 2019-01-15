package Actor

import (
	"github.com/zllangct/RockGO/network"
	"sync"
)

const (
	LOCAL_SERVICE = ""
)

type ActorServiceCaller struct {
	locker sync.RWMutex
	proxy *ActorProxyComponent
	services map[string]*ActorService
}

func NewActorServiceCaller(proxy *ActorProxyComponent) *ActorServiceCaller {
	return &ActorServiceCaller{proxy:proxy,services:make(map[string]*ActorService)}
}

func NewActorServiceCallerFromSession(sess *network.Session ,proxy *ActorProxyComponent) *ActorServiceCaller {
	g,ok:=sess.GetProperty("ActorServiceCaller")
	if ok {
		return g.(*ActorServiceCaller)
	}
	sc:=NewActorServiceCaller(proxy)
	sess.SetProperty("ActorServiceCaller",sc)
	return sc
}

func (this *ActorServiceCaller)Call(role string,serviceName string,args ...interface{})([]interface{},error)  {
	var err error
	//优先尝试缓存客户端，避免反复查询，尽量去中心化
	service,ok:=this.services[serviceName]
	if ok {
		res,err := service.Call(args)
		if err!=nil {
			delete(this.services, serviceName)
		}else {
			return res,err
		}
	}
	//无缓存，或者通过缓存调用失败，重新查询调用
	service,err= this.proxy.GetActorService(role,serviceName)
	if err!=nil {
		return nil,err
	}
	this.services[serviceName] = service
	res,err := service.Call(args)
	if err!=nil {
		delete(this.services, serviceName)
	}
	return res,err
}