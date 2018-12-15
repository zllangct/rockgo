package Component

import (
	"fmt"
	"github.com/zllangct/RockGO/logger"
)

/*
	Component组
	ComponentGroup 一般按照分布式思想，同一功能节点，分为一组。
	比如，网关组、大厅组、逻辑房间、位置服务等
*/
type ComponentGroup struct {
	Name    string
	content []IComponent
}

func (this *ComponentGroup) attachGroupTo(target *Object) {
	o := NewObject(this.Name)
	err:=target.AddObject(o)
	if err!=nil {
		logger.Error(err)
	}
	for _, component := range this.content {
		o.AddComponent(component)
		logger.Info(fmt.Sprintf("Attach component [ %s.%s ] to [ %s ]",this.Name,component.Type().String(),target.name))
	}
}

/*
	所有可用Component组
*/
type ComponentGroups struct {
	group map[string]*ComponentGroup //key:group name , value:component group
}

func (this *ComponentGroups) AddGroup(groupName string, group []IComponent) {
	if this.group ==nil{
		this.group = make(map[string]*ComponentGroup)
	}
	this.group[groupName] = &ComponentGroup{
		Name:    groupName,
		content: group,
	}
}

func (this *ComponentGroups) AttachGroupsTo(groupName []string, target *Object) error {
	child,master,other := false,false,false

	if len(groupName)==0 || (len(groupName)==1 && groupName[0]=="single") {
		if len(this.group)== 2{
			delete(this.group, "child")
		}
		for _, value := range this.group {
			value.attachGroupTo(target)
		}
		return nil
	}
	for _, name := range groupName {
		switch name {
		case"master"  :
			master=true
		case"child"  :
			child=true
		default:
			other=true
		}
	}
	//为空时，默认为master
	if !other && !master && !child{
		groupName= append(groupName, "master")
	}
	//master和其他角色是，需要child
	if other && master && !child{
		groupName= append(groupName, "child")
	}
	//有master，没有其他的时候，不需要child
	if !other && master && child{
		for i, v := range groupName {
			if v == "child" {
				groupName = append(groupName[:i],groupName[i+1:]...)
				break
			}
		}
	}

	for _, name := range groupName {
		if g, ok := this.group[name]; ok {
			g.attachGroupTo(target)
		} else {
			logger.Error(fmt.Sprintf("the group < %s > is not exist", name))
		}
	}
	return nil
}
