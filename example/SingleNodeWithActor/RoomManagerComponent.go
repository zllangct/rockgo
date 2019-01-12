package main

import (
	"github.com/zllangct/RockGO/actor"
	"github.com/zllangct/RockGO/component"
	"sync"
)

type RoomManagerComponent struct {
	Component.Base
	Actor.ActorBase
	locker sync.RWMutex
	rooms map[int]Actor.IActor
	increasing int   //实际运用不这样,此处便宜行事
}

func (this *RoomManagerComponent)Awake(ctx *Component.Context){
	//初始化actor
	this.ActorInit(this.Parent())
	//初始化房间
	this.rooms = make(map[int]Actor.IActor)
	//注册actor消息
	this.AddHandler("newRoom",this.NewRoom,true)
}

func (this *RoomManagerComponent)NewRoom(message *Actor.ActorMessageInfo)  {
	c:=&Actor.ActorComponent{}
	r:=&RoomComponent{}
	_,err:=this.Parent().AddNewbjectWithComponents([]Component.IComponent{c,r})
	if err!=nil {
		message.CallError(err)
	}

	this.locker.Lock()
	this.increasing++
	r.RoomID=this.increasing
	this.rooms[r.RoomID]=c
	this.locker.Unlock()

	message.Reply("",r.RoomID)
}
