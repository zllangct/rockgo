package cluster

/*
	regest rpc
*/
import (
	"fmt"
	"github.com/zllangct/RockGO/logger"
	"github.com/zllangct/RockGO/utils"
	"math/rand"
	"reflect"
	"time"
	"sync"
)

type RpcMsgHandle struct {
	PoolSize  int32
	TaskQueue []chan *RpcRequest
	Apis      map[string]reflect.Value
}

func NewRpcMsgHandle() *RpcMsgHandle {
	return &RpcMsgHandle{
		PoolSize:  utils.GlobalObject.PoolSize,
		TaskQueue: make([]chan *RpcRequest, utils.GlobalObject.PoolSize),
		Apis:      make(map[string]reflect.Value),
	}
}

/*
处理rpc消息
*/
func (this *RpcMsgHandle) DoMsg(request *RpcRequest) {
	if request.Rpcdata.MsgType == RESPONSE && request.Rpcdata.Key != "" {
		//放回异步结果
		AResultGlobalObj.FillAsyncResult(request.Rpcdata.Key, request.Rpcdata)
		return
	} else {
		//rpc 请求
		if f, ok := this.Apis[request.Rpcdata.Target]; ok {
			//存在
			st := time.Now()
			if request.Rpcdata.MsgType == REQUEST_FORRESULT {
				ret := f.Call([]reflect.Value{reflect.ValueOf(request)})
				if len(ret)==0 {
					return
				}
				packdata, err := utils.GlobalObject.RpcCProtoc.GetDataPack().Pack(0, &RpcData{
					MsgType: RESPONSE,
					Result:  ret[0].Interface().(map[string]interface{}),
					Key:     request.Rpcdata.Key,
				})
				if err == nil {
					request.Fconn.Send(packdata)
				} else {
					logger.Error(err)
				}
			} else if request.Rpcdata.MsgType == REQUEST_NORESULT {
				f.Call([]reflect.Value{reflect.ValueOf(request)})
			}

			logger.Debug(fmt.Sprintf("rpc %s cost total time: %f ms", request.Rpcdata.Target, time.Now().Sub(st).Seconds()*1000))
		} else {
			logger.Error(fmt.Sprintf("not found rpc:  %s", request.Rpcdata.Target))
		}
	}
}

func (this *RpcMsgHandle) DeliverToMsgQueue(pkg interface{}) {
	request := pkg.(*RpcRequest)
	//add to worker ConnPool
	index := rand.Int31n(utils.GlobalObject.PoolSize)
	taskQueue := this.TaskQueue[index]
	logger.Debug(fmt.Sprintf("add to rpc ConnPool : %d", index))
	taskQueue <- request
}

func (this *RpcMsgHandle) DoMsgFromGoRoutine(pkg interface{}) {
	request := pkg.(*RpcRequest)
	go this.DoMsg(request)
}

func (this *RpcMsgHandle) AddRouter(router interface{}) {
	value := reflect.ValueOf(router)
	tp := value.Type()
	for i := 0; i < value.NumMethod(); i += 1 {
		name := tp.Method(i).Name

		if _, ok := this.Apis[name]; ok {
			//存在
			panic("repeated rpc " + name)
		}
		this.Apis[name] = value.Method(i)
		logger.Info("add rpc " + name)
	}
}

func (this *RpcMsgHandle) StartWorkerLoop(poolSize int) {
	if utils.GlobalObject.IsThreadSafeMode(){
		this.TaskQueue[0] =  make(chan *RpcRequest, utils.GlobalObject.MaxWorkerLen)
		go func(){
			for{
				select {
				case rpcRequest := <- this.TaskQueue[0]:
					this.DoMsg(rpcRequest)
				case delayCall := <- utils.GlobalObject.GetSafeTimer().GetTriggerChannel():
					delayCall.Call()
				}
			}
		}()
	}else{
		var waitgroup sync.WaitGroup
		for i := 0; i < poolSize; i += 1 {
			c := make(chan *RpcRequest, utils.GlobalObject.MaxWorkerLen)
			this.TaskQueue[i] = c
			waitgroup.Add(1)
			go func(index int, taskQueue chan *RpcRequest) {
				//logger.Info(fmt.Sprintf("init rpc thread ConnPool %d.", index))
				waitgroup.Done()
				for {
					request := <-taskQueue
					this.DoMsg(request)
				}

			}(i, c)
			waitgroup.Wait()
		}
	}
}
//func callerInfo(skip int) string {
//	pc, _, line, ok := runtime.Caller(skip)
//	if !ok {
//		return "error  during runtime.Callers"
//	}
//
//	funcInfo := runtime.FuncForPC(pc)
//	if funcInfo == nil {
//		return "error during runtime.FuncForPC"
//	}
//
//	funcName := funcInfo.Name()
//	if strings.HasPrefix(funcName, ".") {
//		funcName = funcName[strings.Index(funcName, "."):]
//	}
//	funcNames := filepath.Base(funcName)
//	info := fmt.Sprintf("%s,line:%d", funcNames, line)
//	return info
//}
