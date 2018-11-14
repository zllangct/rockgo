package Config

import (
	"github.com/zllangct/RockGO/Component"
	"reflect"
	"io/ioutil"
	"encoding/json"
)

type ConfigComponent struct {
	Config
	ClusterConf
	obj          *Component.Object //父对象
	CustomConfig map[string]string
	ActorConfig  *ClusterConf
	close        chan bool         //关闭信号
}

func (this *ConfigComponent) Type() reflect.Type {
	return reflect.TypeOf(this)
}

func (this *ConfigComponent) Awake(parent *Component.Object) {
	this.obj = parent
}

func (this *ConfigComponent) LoadConfig(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, this.Config)
	if err != nil {
		return err
	}
	return nil
}

func (this *ConfigComponent) LoadCustomConfig(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, this.CustomConfig)
	if err != nil {
		return err
	}
	return nil
}

func (this *ConfigComponent) LoadActorConfig(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(data, this.ActorConfig)
	if err != nil {
		panic(err)
	}
	this.Master.Path = path
	return  nil
}

