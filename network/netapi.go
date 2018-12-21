package network

import (
	"errors"
	"fmt"
	"github.com/zllangct/RockGO/component"
	"github.com/zllangct/RockGO/logger"
	"github.com/zllangct/RockGO/utils"
	"reflect"
	"runtime/debug"
)


type MessageProtocol interface {
	Marshal( interface{})([]byte,error)
	Unmarshal( []byte, interface{})error
}

type NetAPI interface {
	Init(interface{},*Component.Object, map[reflect.Type]uint32,MessageProtocol)  //初始化
	Route(*Session, uint32, []byte)	//反序列化并路由到api处理函数
	//MessageEncode(interface{})(uint32,[]byte,error) //消息序列化
	SetParent(object *Component.Object)		//设置
	Reply(session *Session,message interface{})error
}

type methodType struct {
	method     reflect.Value
	ArgsType  reflect.Type
}

/*  仅初始化route id2mt 会写入，运行过程中并无竞态 */
type ApiBase struct {
	route map[uint32]*methodType
	id2mt map[reflect.Type]uint32
	protoc MessageProtocol
	parent  *Component.Object
	resv   reflect.Value
	isInit bool
}

var ErrNotInit =errors.New("this api is not initialized")
var ErrApiHandlerParamWrong = errors.New("this handler param wrong")
var ErrApiRepeated = errors.New("this ApiBase is  repeated")

func (this *ApiBase)checkInit()  {
	if !this.isInit {
		panic(ErrNotInit)
	}
}

func (this *ApiBase)SetParent(parent *Component.Object){
	this.checkInit()
	this.parent = parent
}

func (this *ApiBase)GetParent() (*Component.Object,error)  {
	this.checkInit()
	var err error
	if this.parent==nil {
		err=errors.New("this api has not parent")
	}
	return this.parent,err
}

func (this *ApiBase)Reply(sess *Session,message interface{}) error {
	t,m,err:=this.MessageEncode(message)
	if err!=nil {
		return err
	}
	return sess.Emit(t,m)
}

func (this *ApiBase)MessageEncode(message interface{}) (uint32,[]byte,error) {
	this.checkInit()
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

func (this *ApiBase)Init(subStruct interface{},parent *Component.Object,id2mt map[reflect.Type]uint32,protocol MessageProtocol)  {
	this.route = map[uint32]*methodType{}
	this.id2mt = id2mt
	this.protoc = protocol
	this.isInit = true
	this.parent = parent
	this.Register(subStruct)
}

func (this *ApiBase)Route(sess *Session, messageID uint32,data []byte)  {
	this.checkInit()
	defer (func() {
		if r := recover(); r != nil {
			var str string
			switch r.(type) {
			case error:
				str =r.(error).Error()
			case string:
				str = r.(string)
			}
			err := errors.New(str+ string(debug.Stack()))
			logger.Error(err)
		}
	})()
	if mt,ok:= this.route[messageID];ok {
		v:= reflect.New(mt.ArgsType)
		err:= this.protoc.Unmarshal(data,v.Interface())
		if err!=nil{
			logger.Debug(fmt.Sprintf("unmarshal message failed :%s ,%s",mt.ArgsType.Elem().Name(),err))
			return
		}
		args:=[]reflect.Value{
			this.resv,
			reflect.ValueOf(sess),
			v.Elem(),
		}
		re:=mt.method.Call(args)
		errinterface:=re[0].Interface()
		if errinterface !=nil && errinterface.(error)!=nil {
			logger.Error(err)
		}
		return
	}
	logger.Debug(fmt.Sprintf("this ApiBase:%d not found",messageID))
}

func (this *ApiBase)On(handler interface{})  {
	this.checkInit()
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
var typeOfError = reflect.TypeOf((*error)(nil)).Elem()
func (this *ApiBase)Register(api interface{})  {
	this.checkInit()
	this.resv = reflect.ValueOf(api)
	typ:=reflect.TypeOf(api)
	logger.Info(fmt.Sprintf("====== start to register API group: [ %s ] ======",typ.Elem().Name()))

	st:=reflect.TypeOf(&Session{})
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mtype := method.Type
		mname := method.Name
		// Method must be exported.
		if method.PkgPath != "" {
			continue
		}
		numin:=mtype.NumIn()
		if numin != 3{
			continue
		}

		sessType := mtype.In(1)
		if sessType!=st {
			continue
		}

		argsType := mtype.In(2)
		if !utils.IsExportedOrBuiltinType(argsType) {
			continue
		}
		// Method needs one out.
		if mtype.NumOut() != 1 {
			continue
		}
		// The return type of the method must be error.
		if returnType := mtype.Out(0); returnType != typeOfError {
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
			logger.Info(fmt.Sprintf("Add api: [ %s ], handler: [ %s.%s(*network.Session,*%s) ]",argsType.Elem().Name(),typ.Elem().Name(),mname,argsType.Elem().Name()))
		}
	}
	logger.Info(fmt.Sprintf("======   register API group: [ %s ] end   ======",typ.Elem().Name()))
}

func (this *ApiBase)GetMessageType(message interface{}) (uint32,bool) {
	id,ok:=this.id2mt[reflect.TypeOf(message)]
	return id,ok
}
