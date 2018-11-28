package Component

import (
	"github.com/zllangct/RockGO/3rd/errors"
	"reflect"
)

/*
	组件组
*/
type ComponentGroup  []IComponent


func (this ComponentGroup) attachGroupTo(target *Object) error {
	for _, component := range this {
		target.AddComponent(component)
	}
	return nil
}
/*
	组件组集合
	使用组件组时组建不会重复添加
*/
type ComponentGroups struct {
	group map[string]ComponentGroup //key:group name , value:component group
}

func (this *ComponentGroups) AddGroup(groupName string, group ComponentGroup) error {
	//去重复
	tempGroup:=map[reflect.Type]IComponent{}
	for _, value := range group {
		tempGroup[reflect.TypeOf(value)]=value
	}
	var g []IComponent
	for _, value := range tempGroup {
		g= append(g, value)
	}
	this.group[groupName] = g
	//加入single组
	tempGroup=map[reflect.Type]IComponent{}
	temp:= append(([]IComponent)(this.group["single"]), ([]IComponent)(group)...)
	for _, value := range temp {
		tempGroup[reflect.TypeOf(value)]=value
	}
	for _, value := range tempGroup {
		this.group["single"]= append(this.group["single"], value)
	}
	return nil
}

func (this *ComponentGroups) AttachGroupTo(groupName []string,target *Object) error {
	tempGroup:=map[reflect.Type]IComponent{}
	//group去重，不可重复添加
	for _, value := range groupName {
		if g,ok:=this.group[value];ok{
			for _, c := range g {
				tempGroup[reflect.TypeOf(g)]=c
			}
		}else{
			return errors.Fail(ErrMissingGroup{}, nil, "no this group:"+value)
		}
	}
	var g []IComponent
	for _, value := range tempGroup {
		g= append(g, value)
	}
	ComponentGroup(g).attachGroupTo(target)
	return nil
}


