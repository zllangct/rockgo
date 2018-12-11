package main

import (
	"errors"
	"github.com/zllangct/RockGO/cluster"
	"github.com/zllangct/RockGO/component"
	"github.com/zllangct/RockGO/configComponent"
	"github.com/zllangct/RockGO/logger"
	"reflect"
	"sync"
	"time"
)

type LoginComponent struct {
	Component.Base
	locker        sync.RWMutex
	nodeComponent *Cluster.NodeComponent
	players       sync.Map // [account,*PlayerInfo]
}

func (this *LoginComponent) GetRequire() map[*Component.Object][]reflect.Type {
	requires := make(map[*Component.Object][]reflect.Type)
	requires[this.Parent.Root()] = []reflect.Type{
		reflect.TypeOf(&Config.ConfigComponent{}),
	}
	return requires
}

func (this *LoginComponent) Awake() {
	err := this.Parent.Root().Find(&this.nodeComponent)
	if err != nil {
		logger.Fatal("get node component failed")
		panic(err)
		return
	}
	//模拟已存在的用户
	this.players.Store("zllang1",&PlayerInfo{Account:"zllang1",Password:"123456",Name:"zhaolei1",Age:11,Coin:100,LastLoginTime:time.Now()})
	this.players.Store("zllang2",&PlayerInfo{Account:"zllang2",Password:"123456",Name:"zhaolei2",Age:12,Coin:200,LastLoginTime:time.Now()})
	this.players.Store("zllang3",&PlayerInfo{Account:"zllang3",Password:"123456",Name:"zhaolei3",Age:13,Coin:300,LastLoginTime:time.Now()})
	this.players.Store("zllang4",&PlayerInfo{Account:"zllang4",Password:"123456",Name:"zhaolei4",Age:14,Coin:400,LastLoginTime:time.Now()})
	this.players.Store("zllang5",&PlayerInfo{Account:"zllang5",Password:"123456",Name:"zhaolei5",Age:15,Coin:500,LastLoginTime:time.Now()})
}

var ErrLoginPlayerNotExist =errors.New("this player doesnt exist")

func (this *LoginComponent)Login(account string)(*PlayerInfo, error) {
	if p,ok:= this.players.Load(account);ok{
		return p.(*PlayerInfo),nil
	}else{
		return nil,ErrLoginPlayerNotExist
	}
}