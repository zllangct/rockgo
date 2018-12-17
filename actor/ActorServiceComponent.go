package Actor

import (
	"errors"
	"fmt"
	"github.com/zllangct/RockGO/cluster"
	"github.com/zllangct/RockGO/component"
	"github.com/zllangct/RockGO/configComponent"
	"github.com/zllangct/RockGO/logger"
	"reflect"
	"sync"
)

type ActorServiceComponent struct {
	Component.Base
	locker        sync.RWMutex
	nodeComponent *Cluster.NodeComponent
	service        *sync.Map // [service,[]actorID]
}

func (this *ActorServiceComponent) IsUnique() int {
	return Component.UNIQUE_TYPE_GLOBAL
}

func (this *ActorServiceComponent) GetRequire() map[*Component.Object][]reflect.Type {
	requires := make(map[*Component.Object][]reflect.Type)
	requires[this.Parent.Root()] = []reflect.Type{
		reflect.TypeOf(&Config.ConfigComponent{}),
		reflect.TypeOf(&Cluster.NodeComponent{}),
	}
	return requires
}

func (this *ActorServiceComponent) Awake() error {
	this.service = &sync.Map{}
	err := this.Parent.Root().Find(&this.nodeComponent)
	if err != nil {
		return err
	}
	err = this.nodeComponent.Register(this)
	if err != nil {
		return err
	}
	return nil
}

func (this *ActorServiceComponent)ServiceInvalidChecking() {

}

func (this *ActorServiceComponent) ServiceInquiry(role string, reply *ActorIDGroup) error {
	if g, ok := this.service.Load(role); ok {
		*reply = *(g.(*ActorIDGroup))
	} else {
		return errors.New(fmt.Sprintf("no any actor of this role :%s", role))
	}
	return nil
}

func (this *ActorServiceComponent)Sync2Peer(op int,service ActorService,except []string) error {
	nodes,err:= this.nodeComponent.GetNodeGroup("actor_service")
	if err!=nil{
		logger.Error(err)
	}
	//过滤掉排除的
	nodes_filter:=nodes.Nodes()
	for index, value := range nodes_filter {
		for _, ex := range except {
			if ex == value {
				nodes_filter= append(nodes_filter[:index],nodes_filter[index+1:]...)
				break
			}
		}
	}
	args:=SyncService{
		ActorService:service,
		Peer:[]string{},
		Op:op,
	}
	args.Peer= append(args.Peer, this.nodeComponent.Addr)

	for _, addr := range nodes_filter {
		c,err:=this.nodeComponent.GetNodeClient(addr)
		if err!=nil {
			logger.Error(err)
			continue
		}
		var reply bool
		err= c.Call("ActorServiceComponent.SyncService",args,&reply)
		if err!=nil {
			logger.Error(err)
			continue
		}
	}
	return nil
}

type ActorService struct {
	Service string
	ActorID ActorID
}

const(
	SyncServiceOpAdd =iota //服务增加
	SyncServiceOpRemove	   //服务下线
)

type SyncService struct {
	ActorService
	Peer []string
	Op int
}

func (this *ActorServiceComponent) ServiceRegister(args ActorService, reply *bool) error {
	g, ok := this.service.LoadOrStore(args.Service, &ActorIDGroup{
		Actors: []ActorID{args.ActorID},
	})
	ag:=g.(*ActorIDGroup)
	has:=true
	if ok {
		if !ag.Has(args.ActorID) {
			has=false
			ag.Add(args.ActorID)
		}
	}
	if !ok && has {
		return nil
	}
	return this.Sync2Peer(SyncServiceOpAdd,args,[]string{this.nodeComponent.Addr})
}

func (this *ActorServiceComponent) ServiceUnregister(args ActorService, reply *bool) error {
	this.service.Range(func(key, value interface{}) bool {
		value.(*ActorIDGroup).Sub(args.ActorID)
		return true
	})
	return this.Sync2Peer(SyncServiceOpRemove,args,[]string{this.nodeComponent.Addr})
}

func (this *ActorServiceComponent) SyncService(args SyncService, reply *bool) error {
	var rep bool
	if args.Op== SyncServiceOpAdd{
		_=this.ServiceRegister(args.ActorService,&rep)
	}else{
		_=this.ServiceUnregister(args.ActorService,&rep)
	}
	return this.Sync2Peer(args.Op,args.ActorService,args.Peer)
}

