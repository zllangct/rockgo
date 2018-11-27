package Config

import (
	"encoding/json"
	"errors"
	"github.com/zllangct/RockGO/Component"
	"github.com/zllangct/RockGO/logger"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
)

type ConfigComponent struct {
	Component.Base
	CommonConfigPath string
	CusterConfigPath string
	CustomConfigPath map[string]string
	CommonConfig     *CommonConfig
	CusterConfig     *CusterConfig
	CustomConfig     map[string]interface{}
}


func (this *ConfigComponent) IsUnique() bool {
	return true
}

func (this *ConfigComponent) Awake() {
	this.CommonConfigPath = "./Config/CommonConfig.json"
	this.CusterConfigPath = "./Config/CusterConfig.json"
	//初始化默认配置
	this.SetDefault()
	//读取配置文件
	this.ReloadConfig()
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
			b, err := json.MarshalIndent(cfg,"","    ")
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
	this.loadConfig(this.CommonConfigPath, this.CommonConfig)
	this.loadConfig(this.CusterConfigPath, this.CusterConfig)
	for name, path := range this.CustomConfigPath{
		this.loadConfig(path, this.CustomConfig[name])
	}
}
// ConfigComponent.CustomConfig[name] = structure
func (this *ConfigComponent) LoadCoustomConfig(name string,path string,structure interface{})(err error) {
	if name == "" || path ==""{
		return errors.New("config name or path can ont be empty")
	}
	kind:=reflect.TypeOf(structure).Kind()
	if  kind != reflect.Ptr && kind != reflect.Map{
		err=errors.New("structure must be pointer or map")
		return
	}
	this.loadConfig(path, structure)
	this.CustomConfig[name]=structure
	this.CustomConfig[name]= path
	return
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
	this.CusterConfig = &CusterConfig{
		MasterIPAddress: "127.0.0.1",
		MasterPort:      8888,
		LocalPort:       6666,
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
	LogFileMax       int             // log文件最大值
	LogConsolePrint  bool            //是否输出log到控制台
}

type CusterConfig struct {
	MasterIPAddress string
	MasterPort      int
	LocalPort       int
	APPName         []string
	Rule            []string
}
