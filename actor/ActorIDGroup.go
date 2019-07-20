package Actor

import (
	"errors"
	"math/rand"
	"sync"
)

type ActorIDGroup struct {
	locker sync.RWMutex
	Actors []ActorID
}

func (this ActorIDGroup) isRepeated(target ActorID) bool {
	//外层注意加锁

	for _, value := range this.Actors {
		if value.Equal(target) {
			return true
		}
	}
	return false
}

func (this *ActorIDGroup) Add(id ActorID) {
	this.locker.Lock()
	defer this.locker.Unlock()

	if !this.isRepeated(id) {
		this.Actors = append(this.Actors, id)
	}
}

func (this *ActorIDGroup) Sub(id ActorID) {
	this.locker.Lock()
	defer this.locker.Unlock()

	for i, value := range this.Actors {
		if value.Equal(id) {
			this.Actors = append(this.Actors[:i], this.Actors[i+1:]...)
			return
		}
	}
}

func (this *ActorIDGroup) Has(id ActorID) bool {
	this.locker.RLock()
	defer this.locker.RUnlock()

	for _, value := range this.Actors {
		if value.Equal(id) {
			return true
		}
	}
	return false
}

func (this *ActorIDGroup) Get() []ActorID {
	this.locker.RLock()
	defer this.locker.RUnlock()

	as := make([]ActorID, 0, len(this.Actors))
	copy(as, this.Actors)
	return as
}

func (this *ActorIDGroup) RndOne() (ActorID, error) {
	this.locker.RLock()
	defer this.locker.RUnlock()
	l := len(this.Actors)
	if l == 0 {
		return nil, errors.New("this actor id group is empty")
	}
	r := rand.Intn(l)
	return this.Actors[r], nil
}
