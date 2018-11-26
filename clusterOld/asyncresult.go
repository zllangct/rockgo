package clusterOld

import (
	"errors"
	"fmt"
	"github.com/zllangct/RockGO/logger"
	"github.com/zllangct/RockGO/utils"
	_ "os"
	"sync"
	"time"
)

type AsyncResult struct {
	key    string
	result chan *RpcData
}

var AResultGlobalObj *AsyncResultMgr = NewAsyncResultMgr()

func NewAsyncResult(key string) *AsyncResult {
	return &AsyncResult{
		key:    key,
		result: make(chan *RpcData, 1),
	}
}

func (this *AsyncResult) GetKey() string {
	return this.key
}

func (this *AsyncResult) SetResult(data *RpcData) {
	this.result <- data
}

func (this *AsyncResult) GetResult(timeout time.Duration) (*RpcData, error) {
	select {
	case <-time.After(timeout):
		logger.Error(fmt.Sprintf("GetResult AsyncResult: timeout %s", this.key))
		close(this.result)
		return &RpcData{}, errors.New(fmt.Sprintf("GetResult AsyncResult: timeout %s", this.key))
	case result := <-this.result:
		return result, nil
	}
	return &RpcData{}, errors.New("GetResult AsyncResult error. reason: no")
}

type AsyncResultMgr struct {
	idGen   *utils.UUIDGenerator
	results sync.Map
}

func NewAsyncResultMgr() *AsyncResultMgr {
	return &AsyncResultMgr{
		results: sync.Map{},
		idGen:   utils.NewUUIDGenerator("async_result_"),
	}
}

func (this *AsyncResultMgr) Add() *AsyncResult {
	r := NewAsyncResult(this.idGen.Get())
	this.results.Store(r.GetKey(),r)
	return r
}

func (this *AsyncResultMgr) Remove(key string) {
	this.results.Delete(key)
}

func (this *AsyncResultMgr) GetAsyncResult(key string) (*AsyncResult, error) {
	r, ok := this.results.Load(key)
	if ok {
		return r.(*AsyncResult), nil
	} else {
		return nil, errors.New("not found AsyncResult")
	}
}

func (this *AsyncResultMgr) FillAsyncResult(key string, data *RpcData) error {
	r, err := this.GetAsyncResult(key)
	if err == nil {
		this.Remove(key)
		r.SetResult(data)
		return nil
	} else {
		return err
	}
}
