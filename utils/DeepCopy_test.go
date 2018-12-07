package utils_test

import (
	"fmt"
	"github.com/zllangct/RockGO/utils"
	"reflect"
	"testing"
)

type (
	Player struct {
		Id     int
		Level  int
		Heroes map[int]*Hero
		Equips []*Equip
	}

	Hero struct {
		Id     int
		Level  int
		Skills []*Skill
	}

	Equip struct {
		Id    int
		Level int
	}

	Skill struct {
		Id    int
		Level int
	}
)

func NewHero() *Hero {
	return &Hero{
		Id:     1,
		Level:  1,
		Skills: append([]*Skill{NewSkill()}, NewSkill(), NewSkill()),
	}
}

func NewSkill() *Skill {
	return &Skill{1, 1}
}

func NewEquip() *Equip {
	return &Equip{1, 1}
}

func NewPlayer() *Player {
	return &Player{
		Id:     1,
		Level:  1,
		Heroes:   map[int]*Hero{1: NewHero(), 2: NewHero(), 3: NewHero()},
		Equips: append([]*Equip{NewEquip()}, NewEquip(), NewEquip()),
	}
}

func (self *Hero) Print() {
	fmt.Printf("Id=%d, Level=%d\n", self.Id, self.Level)
	for _, v := range self.Skills {
		fmt.Printf("%v\n", *v)
	}
}

func (self *Player) Print() {
	fmt.Printf("Id=%d, Level=%d\n", self.Id, self.Level)
	for _, v := range self.Heroes {
		v.Print()
	}

	for _, v := range self.Equips {
		fmt.Printf("%+v\n", *v)
	}
}


func TestCopy(T *testing.T) {
	p1 := NewPlayer()
	p2 := utils.Copy(p1).(*Player)
	fmt.Println(reflect.DeepEqual(p1, p2))
}
