package main

import (
	"github.com/zllangct/RockGO/actor"
	"github.com/zllangct/RockGO/component"
	"sync"
)

type RoomManagerComponent struct {
	Component.Base
	locker sync.RWMutex
	rooms map[int]Actor.IActor
	messageHandler map[string]func(message *Actor.ActorMessageInfo)
	increasing int   //实际运用不这样,此处便宜行事
	actor   *Actor.ActorComponent
}

func (this *RoomManagerComponent)Awake(ctx *Component.Context){
	this.rooms = make(map[int]Actor.IActor)
	this.messageHandler=map[string]func(message *Actor.ActorMessageInfo){
		"newRoom":this.NewRoom,
	}
}

func (this *RoomManagerComponent) MessageHandlers() map[string]func(message *Actor.ActorMessageInfo) {
	return this.messageHandler
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
