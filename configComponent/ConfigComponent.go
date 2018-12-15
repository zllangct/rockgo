package Config

import (
	"encoding/json"
	"errors"
	"github.com/zllangct/RockGO/component"
	"github.com/zllangct/RockGO/logger"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
)

var Config *ConfigComponent

type ConfigComponent struct {
	Component.Base
	commonConfigPath  string
	clusterConfigPath string
	customConfigPath  map[string]string
	CommonConfig      *CommonConfig
	ClusterConfig     *ClusterConfig
	CustomConfig      map[string]interface{}
}

func (this *ConfigComponent) IsUnique() int {
	return Component.UNIQUE_TYPE_GLOBAL
}

func (this *ConfigComponent) Awake() error {
	this.commonConfigPath = "./config/CommonConfig.json"
	this.clusterConfigPath = "./config/ClusterConfig.json"
	//初始化默认配置
	this.SetDefault()
	//读取配置文件
	this.ReloadConfig()
	//全局共享
	Config = this
	return nil
}

func (this *ConfigComponent) loadConfig(configpath string, cfg interface{}) error {
	data, err := ioutil.ReadFile(configpath)
	if err != nil {
		//文件不存在时创建配置文件，并写入默认值
		if os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(configpath), 0666); err != nil {
				if os.IsPermission(err) {
					return err
				}
			}
			b, err := json.MarshalIndent(cfg, "", "    ")
			if err != nil {
				return err
			}
			err = ioutil.WriteFile(configpath, b, 0666)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		err = json.Unmarshal(data, cfg)
		if err != nil {
			return err
		}
	}
	return nil
}

//重新读取配置文件，包括自定义配置文件
func (this *ConfigComponent) ReloadConfig() {
	err := this.loadConfig(this.commonConfigPath, this.CommonConfig)
	if err != nil {
		panic(err)
	}
	err = this.loadConfig(this.clusterConfigPath, this.ClusterConfig)
	if err != nil {
		panic(err)
	}
	for name, path := range this.customConfigPath {
		err = this.loadConfig(path, this.CustomConfig[name])
		if err != nil {
			panic(err)
		}
	}
}

// configComponent.CustomConfig[name] = structure
func (this *ConfigComponent) LoadCustomConfig(name string, path string, structure interface{}) (err error) {
	if name == "" || path == "" {
		return errors.New("config name or path can ont be empty")
	}
	kind := reflect.TypeOf(structure).Kind()
	if kind != reflect.Ptr && kind != reflect.Map {
		err = errors.New("structure must be pointer or map")
		return
	}
	err = this.loadConfig(path, structure)
	this.CustomConfig[name] = structure
	this.CustomConfig[name] = path
	return err
}

func (this *ConfigComponent) SetDefault() {
	this.CommonConfig = &CommonConfig{
		Debug: true,
		//runtime
		RuntimeMaxWorker: runtime.NumCPU(),
		//log
		LogLevel:        logger.DEBUG,
		LogPath:         "./log",
		LogMode:         logger.ROLLFILE,
		LogFileUnit:     logger.MB,
		LogFileMax:      10,
		LogConsolePrint: true,
	}
	this.CustomConfig = nil
	this.ClusterConfig = &ClusterConfig{
		MasterAddress: "127.0.0.1:6666",
		LocalAddress:  "127.0.0.1:6666",
		AppName:       "defaultApp",
		Role:          []string{"master"},
		NodeDefine: map[string]Node{
			/*
				内置角色：master、child、location、actor_location
			*/
			//master节点
			"node_master": {LocalAddress: "0.0.0.0:6666", Role: []string{"master"}},
			//位置服务节点
			"node_location": {LocalAddress: "0.0.0.0:6603", Role: []string{"location"}},
			//actor位置服务节点
			"node_actor_location": {LocalAddress: "0.0.0.0:6604", Role: []string{"actor_location"}},
			//位置服务+actor位置服务
			"node_location_actor_location": {LocalAddress: "0.0.0.0:6604", Role: []string{"location", "actor_location"}},

			//用户自定义
			"node_gate":  {LocalAddress: "0.0.0.0:6601", Role: []string{"gate"}},
			"node_login": {LocalAddress: "0.0.0.0:6602", Role: []string{"login"}},
			"node_room":  {LocalAddress: "0.0.0.0:6605", Role: []string{"room"}},

			//dubug 或 单服
			"node_single": {LocalAddress: "0.0.0.0:6666", Role: []string{"master","gate", "login", "room"}},
		},

		ReportInterval:       3000,
		RpcTimeout:           9000,
		RpcCallTimeout:       5000,
		RpcHeartBeatInterval: 3000,
		IsLocationMode:       true,

		NetConnTimeout:   9000,
		NetListenAddress: "0.0.0.0:5555",

		IsActorModel: true,
	}
}

/*
	Default config
*/
type CommonConfig struct {
	Debug            bool            //是否为Debug模式
	RuntimeMaxWorker int             //runtime最大工作线程
	LogLevel         logger.LEVEL    //log等级
	LogPath          string          //log的存储根目录
	LogMode          logger.ROLLTYPE //log文件存储模式，分为按文件大小分割，按日期分割
	LogFileUnit      logger.UNIT     //log文件大小单位
	LogFileMax       int64           // log文件最大值
	LogConsolePrint  bool            //是否输出log到控制台

}
type Node struct {
	LocalAddress string
	Role         []string
}

type ClusterConfig struct {
	MasterAddress string   //Master 地址,例如:127.0.0.1:8888
	LocalAddress  string   //本节点IP,注意配置文件时，填写正确的局域网地址或者外网地址，不可为0.0.0.0
	AppName       string   //本节点拥有的app
	Role          []string //本节点拥有角色
	NodeDefine    map[string]Node

	ReportInterval       int  //子节点节点信息上报间隔，单位秒
	RpcTimeout           int  //tcp链接超时，单位毫秒
	RpcCallTimeout       int  //rpc调用超时
	RpcHeartBeatInterval int  //tcp心跳间隔
	IsLocationMode       bool //是否启用位置服务器

	//外网
	NetConnTimeout   int    //外网链接超时
	NetListenAddress string //网关对外服务地址

	//actor
	IsActorModel bool
}
