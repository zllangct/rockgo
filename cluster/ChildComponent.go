package Cluster

import (
	"github.com/zllangct/RockGO/component"
	"github.com/zllangct/RockGO/configComponent"
	"github.com/zllangct/RockGO/logger"
	"github.com/zllangct/RockGO/rpc"
	"reflect"
	"sync"
	"time"
)

type ChildComponent struct {
	Component.Base
	locker sync.RWMutex
	NodeID string
	rpcMaster      *rpc.TcpClient					//master节点
	nodeComponent  *NodeComponent
	reportCollecter []func()(string,float32)

	config *Config.ConfigComponent
}

func (this *ChildComponent) GetRequire() (map[*Component.Object][]reflect.Type) {
	requires:=make(map[*Component.Object][]reflect.Type)
	requires[this.Parent.Root()] = []reflect.Type{
		reflect.TypeOf(&Config.ConfigComponent{}),
		reflect.TypeOf(&NodeComponent{}),
	}
	return requires
}

func(this *ChildComponent)Awake(){
	err:= this.Parent.Root().Find(&this.config)
	if err != nil {
		logger.Fatal("get config component failed")
		panic(err)
		return
	}
	err= this.Parent.Root().Find(&this.nodeComponent)
	if err != nil {
		logger.Fatal("get node component failed")
		panic(err)
		return
	}

	this.ConnectToMaster()
	go this.DoReport()
}

//上报节点信息
func (this *ChildComponent)DoReport()  {
	args:=&NodeInfo{
		Address:      this.config.ClusterConfig.LocalAddress,
		Group:   this.config.ClusterConfig.Role,
		AppName: this.config.ClusterConfig.AppName,
	}
	var reply bool
	var interval = time.Duration(this.config.ClusterConfig.ReportInterval)
	for {
		this.locker.RLock()
		m:=make(map[string]float32)
		for _, collector := range this.reportCollecter {
			f,d:=collector()
			m[f]=d
		}
		args.Info = m
		this.locker.RUnlock()
		this.rpcMaster.Call("MasterService.ReportNodeInfo",args,&reply)
		time.Sleep(time.Second * interval)
	}
}

//增加上报信息
func (this *ChildComponent)AddReportInfo(field string,collectFunction func()(string,float32)) {
	this.locker.Lock()
	this.reportCollecter = append(this.reportCollecter, collectFunction)
	this.locker.Unlock()
}

//连接到master
func (this *ChildComponent) ConnectToMaster() error {
	addr:=this.config.ClusterConfig.MasterAddress
	callback :=func(event string,data ...interface{}) {
		switch event {
		case "close":
			this.nodeComponent.isOnline =false
			this.OnDropped()
		}
	}
	var err error
	for err != nil {
		this.rpcMaster, err = this.nodeComponent.ConnectToNode(addr,callback)
		if err != nil {
			return err
		}
	}
	this.locker.Lock()
	this.nodeComponent.isOnline = true
	this.locker.Unlock()
	println(time.Now().Format("2006-01-02T 15:04:05"), "  reconnect to master success")
	return nil
}

//当节点掉线
func (this *ChildComponent)OnDropped()  {
	//重新连接
	for{
		println(time.Now().Format("2006-01-02T 15:04:05"), "  reconnecting to master ......")
		err :=this.ConnectToMaster()
		if err ==nil {
			this.locker.RLock()
			if this.nodeComponent.isOnline {
				this.locker.RUnlock()
				return
			}
			this.locker.RUnlock()
		}
		time.Sleep(time.Second * 2)
	}
}