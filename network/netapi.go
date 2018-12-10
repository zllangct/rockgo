package network

import (
	"errors"
	"fmt"
	"github.com/zllangct/RockGO/logger"
	"reflect"
)


type MessageProtocol interface {
	Marshal(messageType interface{})([]byte,error)
	Unmarshal(data []byte,messageType interface{})error
}

type NetAPI interface {
	Init(id2mt map[reflect.Type]uint32)
	Route(sess *Session, messageID uint32,data []byte)
}

type methodType struct {
	method     reflect.Value
	ArgsType  reflect.Type
}

/*  仅初始化route id2mt 会写入，运行过程中并无竞态 */
type Base struct {
	route map[uint32]*methodType
	id2mt map[reflect.Type]uint32
	protoc MessageProtocol
}

func (this *Base)Init(id2mt map[reflect.Type]uint32,protocol MessageProtocol)  {
	this.route = map[uint32]*methodType{}
	this.id2mt = id2mt
	this.protoc = protocol
}

var ErrApiHandlerParamWrong = errors.New("this handler param wrong")
var ErrApiNotInit = errors.New("this Base is not initialized")
var ErrApiRepeated = errors.New("this Base is  repeated")

func (this *Base)On(handler interface{})  {
	if this.id2mt ==nil {
		panic(ErrApiNotInit)
	}

	mValue:=reflect.ValueOf(handler)
	mType :=reflect.TypeOf(handler)
	paramsCount:= mType.NumIn()
	if paramsCount != 2 {
		panic(ErrApiHandlerParamWrong)
	}
	argsType:=mType.In(2)
	if index,ok:=this.id2mt[argsType];ok {
		if _,exist:= this.route[index];exist{
			panic(ErrApiRepeated)
		}else{
			this.route[index] = &methodType{
				method:mValue,
				ArgsType:argsType,
			}
		}
	}
}

func (this *Base)Route(sess *Session, messageID uint32,data []byte)  {
	if mt,ok:= this.route[messageID];ok {
		v:= reflect.New(mt.ArgsType)
		err:= this.protoc.Unmarshal(data,v.Interface())
		if err!=nil{
			logger.Error(fmt.Sprintf("unmarshal message failed :%d ",messageID))
		}
		args:=[]reflect.Value{
			reflect.ValueOf(sess),
			v,
		}
		mt.method.Call(args)
		return
	}
	logger.Debug(fmt.Sprintf("this Base:%d not found",messageID))
}