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
	target.AddObject(o)
	for _, component := range this.content {
		o.AddComponent(component)
	}
}

/*
	所有可用Component组
*/
type ComponentGroups struct {
	group map[string]*ComponentGroup //key:group name , value:component group
}

func (this *ComponentGroups) AddGroup(groupName string, group []IComponent) {
	this.group[groupName] = &ComponentGroup{
		Name:    groupName,
		content: group,
	}
}

func (this *ComponentGroups) AttachGroupsTo(groupName []string, target *Object) error {
	child,master := false,false
	for _, name := range groupName {
		if g, ok := this.group[name]; ok {
			g.attachGroupTo(target)
		} else {
			if name == "single" {
				for _, value := range this.group {
					value.attachGroupTo(target)
				}
			}
			logger.Error(fmt.Sprintf("the group < %s > is not exist", name))
		}
		if name == "master" {
			master = true
		}
		if name == "child"{
			child = true
		}
	}
	if !child && !master{
		if g, ok := this.group["child"]; ok {
			g.attachGroupTo(target)
		}
	}
	return nil
}
