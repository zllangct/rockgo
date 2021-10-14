package main

import (
	"errors"
	"github.com/zllangct/rockgo/actor"
	"github.com/zllangct/rockgo/ecs"
)

type RoomComponent struct {
	ecs.ComponentBase
	Actor.ActorBase
	players map[int]*Player
	RoomID  int
}

func (this *RoomComponent) Initialize() error {
	this.ActorInit(this.Parent(), Actor.ACTOR_TYPE_SYNC)
	return nil
}

func (this *RoomComponent) Awake(ctx *ecs.Context) {

}

var ErrLoginPlayerNotExist = errors.New("this player doesnt exist")

func (this *RoomComponent) Enter(player *Player) ([]interface{}, error) {
	this.players[player.UID] = player
	return []interface{}{"hello,welcome to enter this room."}, nil
}
