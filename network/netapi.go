package network

import (
	"errors"
	"fmt"
	"github.com/zllangct/RockGO/logger"
	"github.com/zllangct/RockGO/utils"
	"reflect"
)


type MessageProtocol interface {
	Marshal( interface{})([]byte,error)
	Unmarshal( []byte, interface{})error
}

type NetAPI interface {
	Init(interface{},map[reflect.Type]uint32,MessageProtocol)
	Route(*Session, uint32, []byte)
	MessageEncode(interface{})(uint32,[]byte,error)
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

func (this *Base)MessageEncode(message interface{}) (uint32,[]byte,error) {
	b,err:= this.protoc.Marshal(message)
	if err!=nil {
		return 0, nil, err
	}
	t:=reflect.TypeOf(message)
	if id,ok:=this.id2mt[t];ok {
		return id, b, nil
	}else{
		return 0, nil,errors.New(fmt.Sprintf("this message type: %s not be registered",t.Name()))
	}
}

func (this *Base)Init(self interface{},id2mt map[reflect.Type]uint32,protocol MessageProtocol)  {
	this.route = map[uint32]*methodType{}
	this.id2mt = id2mt
	this.protoc = protocol

	this.Register(self)
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

func (this *Base)Register(api interface{})  {
	if this.id2mt ==nil {
		panic(ErrApiNotInit)
	}

	typ:=reflect.TypeOf(api)
	logger.Info(fmt.Sprintf("====== start to register API group:%s ======",typ.Name()))

	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mtype := method.Type
		mname := method.Name
		// Method must be exported.
		if method.PkgPath != "" {
			continue
		}
		numin:=mtype.NumIn()
		if numin != 2{
			continue
		}

		argsType := mtype.In(2)
		if !utils.IsExportedOrBuiltinType(argsType) {
			continue
		}

		if index,ok:=this.id2mt[argsType];ok {
			if _,exist:= this.route[index];exist{
				panic(ErrApiRepeated)
			}else{
				this.route[index] = &methodType{
					method:method.Func,
					ArgsType:argsType,
				}
			}
			logger.Info(fmt.Sprintf("Add api: %s, handler: %s.%s(*network.Session,*%s)",argsType.Name(),typ.Name(),mname,argsType.Name()))
		}
	}
	logger.Info(fmt.Sprintf("====== register API group:%s end ======",typ.Name()))
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