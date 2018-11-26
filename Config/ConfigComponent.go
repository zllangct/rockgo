package Config

import (
	"encoding/json"
	"github.com/zllangct/RockGO/Component"
	"io/ioutil"
)

type ConfigComponent struct {
	Component.Base
	CommonConfigPath string
	CustomConfigPath string
	CusterConfigPath string
	CommonConfig     map[string]string
	CustomConfig     map[string]string
	CusterConfig     map[string]string
}

func (this *ConfigComponent)IsUnique() bool {
	return true
}

func (this *ConfigComponent) Awake() {
	this.CommonConfigPath = "./config/CommonConfig.json"
	this.CustomConfigPath = "./config/CustomConfig.json"
	this.CusterConfigPath = "./config/CusterConfig.json"
}

func (this *ConfigComponent) LoadConfig(path string, cfg map[string]string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, cfg)
	if err != nil {
		return err
	}
	return nil
}

func (this *ConfigComponent) ReloadConfig() {
	this.LoadConfig(this.CommonConfigPath, this.CommonConfig)
	this.LoadConfig(this.CustomConfigPath, this.CustomConfig)
	this.LoadConfig(this.CusterConfigPath, this.CusterConfig)
}
