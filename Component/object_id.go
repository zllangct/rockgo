package Component

import (
	"time"
	"math/rand"
	"sync"
)

var once sync.Once
var randomSeed *rand.Rand // Random number generator
var AppID int64
var v uint16
func makeTimestamp() int64 {
	return time.Now().UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}

func makeObjectId() int64 {
	once.Do(initRand)
	v++
	return AppID<<48+(makeTimestamp()<<48 ) + int64(v)
}

func initRand() {
	randomSeed = rand.New(rand.NewSource(makeTimestamp()))
}
