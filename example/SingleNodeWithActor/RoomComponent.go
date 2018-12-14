package main

import (
	"errors"
	"github.com/zllangct/RockGO/actor"
	"github.com/zllangct/RockGO/cluster"
	"github.com/zllangct/RockGO/component"
	"github.com/zllangct/RockGO/configComponent"
	"reflect"
	"sync"
)

type RoomComponent struct {
	Component.Base
	locker        sync.RWMutex
	nodeComponent *Cluster.NodeComponent
	players       map[int]*Player
	RoomID        int
	actor         Actor.IActor
}

func (this *RoomComponent) GetRequire() map[*Component.Object][]reflect.Type {
	requires := make(map[*Component.Object][]reflect.Type)
	requires[this.Parent.Root()] = []reflect.Type{
		reflect.TypeOf(&Config.ConfigComponent{}),
	}
	return requires
}

func (this *RoomComponent) Awake() error{
	err := this.Parent.Root().Find(&this.nodeComponent)
	if err != nil {
		return err
	}
	err= this.nodeComponent.Register(this)
	if err!=nil {
		return err
	}
	return nil
}

var ErrLoginPlayerNotExist =errors.New("this player doesnt exist")

func (this *RoomComponent)Enter(player *Player) error {
	this.players[player.UID]=player
	sender,err:=this.Actor()
	if err!=nil {
		return err
	}
	err=player.addr.Tell(sender, &Actor.ActorMessage{
		Tittle:"HelloEnter",
		Data:[]interface{}{"hello,welcome to enter this room."},
	})
	if err!=nil{
		return err
	}
	return nil
}

func (this *RoomComponent)Actor() (Actor.IActor,error) {
	if this.actor==nil{
		err:= this.Parent.Find(this.actor)
		if err!=nil {
			return nil,err
		}
	}
	return this.actor,nil
}