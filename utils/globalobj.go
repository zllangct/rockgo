package utils

import (
	"encoding/json"
	"github.com/zllangct/RockGO/RockInterface"
	"github.com/zllangct/RockGO/logger"
	"github.com/zllangct/RockGO/timer"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type GlobalObj struct {
	TcpServers              map[string]RockInterface.Iserver
	TcpServer               RockInterface.Iserver
	OnConnectioned          func(fconn RockInterface.ISession)
	OnClosed                func(fconn RockInterface.ISession)
	OnClusterConnectioned   func(fconn RockInterface.ISession) //集群rpc root节点回调
	OnClusterClosed         func(fconn RockInterface.ISession)
	OnClusterCConnectioned  func(fconn RockInterface.Iclient) //集群rpc 子节点回调
	OnClusterCClosed        func(fconn RockInterface.Iclient)
	OnChildNodeConnected    func(name string, fconn RockInterface.IWriter) //子节点链接成功
	OnChildNodeDisconnected func(name string, fconn RockInterface.IWriter)
	OnServerStop            func() //服务器停服回调
	Protoc                  RockInterface.IServerProtocol
	RpcSProtoc              RockInterface.IServerProtocol
	RpcCProtoc              RockInterface.IClientProtocol
	Host                    string
	DebugPort               int      //telnet port 用于单机模式
	WriteList               []string //telnet ip list
	TcpPort                 int
	MaxConn                 int
	IntraMaxConn            int //内部服务器最大连接数
	//log
	LogPath      string
	LogName      string
	MaxLogNum    int32
	MaxFileSize  int64
	LogFileUnit  logger.UNIT
	LogLevel     logger.LEVEL
	SetToConsole bool
	LogFileType  int32
	PoolSize     int32
	ConnSize     int32
	ConnTimeOut  int32
	MatchQueue   int32 //匹配队列容量
	//RoomPreGenNum     int32 //匹配房间最大缓存量
	//RoomPreGen        bool  //是否使用房间预创建
	MultiConnMode     bool
	MaxWorkerLen      int32
	MaxSendChanLen    int32
	FrameSpeed        uint8
	Name              string
	MaxPacketSize     uint32
	FrequencyControl  string                            //  100/h, 100/m, 100/s
	CmdInterpreter    RockInterface.ICommandInterpreter //xingo debug tool Interpreter
	ProcessSignalChan chan os.Signal
	safeTimerScheduel *timer.SafeTimerScheduel
	RoomMaxPlayer     int32
	/*  database */
	MysqlUser            string
	MysqlPasswd          string
	MysqlDBName          string
	MysqlIP              string
	RedisIP              string
	RedisPasswd          string
	ExploitsOnePageCount int64
	ExploitsMaxPageCount int64
	RedisDB1             int
	ClientVersion        map[string]string
}

func (this *GlobalObj) GetFrequency() (int, string) {
	fc := strings.Split(this.FrequencyControl, "/")
	if len(fc) != 2 {
		return 0, ""
	} else {
		fc0_int, err := strconv.Atoi(fc[0])
		if err == nil {
			return fc0_int, fc[1]
		} else {
			logger.Error("FrequencyControl params error: ", this.FrequencyControl)
			return 0, ""
		}
	}
}

func (this *GlobalObj) IsThreadSafeMode() bool {
	if this.PoolSize == 1 {
		return true
	} else {
		return false
	}
}

func (this *GlobalObj) GetSafeTimer() *timer.SafeTimerScheduel {
	return this.safeTimerScheduel
}

func (this *GlobalObj) Reload() {
	//读取用户自定义配置
	data, err := ioutil.ReadFile("Conf/server.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(data, this)
	if err != nil {
		panic(err)
	} else {
		//init safetimer
		if GlobalObject.safeTimerScheduel == nil && GlobalObject.IsThreadSafeMode() {
			GlobalObject.safeTimerScheduel = timer.NewSafeTimerScheduel()
		}
	}
}

var GlobalObject *GlobalObj

func init() {
	GlobalObject = &GlobalObj{
		TcpServers:              make(map[string]RockInterface.Iserver),
		Host:                    "0.0.0.0",
		TcpPort:                 8109,
		MaxConn:                 12000,
		IntraMaxConn:            100,
		LogPath:                 "./log",
		LogName:                 "server.log",
		MaxLogNum:               10,
		MaxFileSize:             100,
		LogFileUnit:             logger.KB,
		LogLevel:                logger.ERROR,
		SetToConsole:            true,
		LogFileType:             1,
		PoolSize:                10,
		MultiConnMode:           false,
		ConnSize:                0,
		ConnTimeOut:             30,
		MaxWorkerLen:            1024 * 2,
		MaxSendChanLen:          1024,
		FrameSpeed:              30,
		MatchQueue:              100,
		OnConnectioned:          func(fconn RockInterface.ISession) {},
		OnClosed:                func(fconn RockInterface.ISession) {},
		OnClusterConnectioned:   func(fconn RockInterface.ISession) {},
		OnClusterClosed:         func(fconn RockInterface.ISession) {},
		OnClusterCConnectioned:  func(fconn RockInterface.Iclient) {},
		OnClusterCClosed:        func(fconn RockInterface.Iclient) {},
		OnChildNodeConnected:    func(name string, fconn RockInterface.IWriter) {}, //子节点链接成功
		OnChildNodeDisconnected: func(name string, fconn RockInterface.IWriter) {},
		ProcessSignalChan:       make(chan os.Signal, 1),
		RoomMaxPlayer:           20000,
		ExploitsOnePageCount:    20,
		ExploitsMaxPageCount:    10,
		RedisDB1:                2,
		ClientVersion:           map[string]string{},
	}
	GlobalObject.Reload()
}
