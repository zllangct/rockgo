package Cluster

import (
	"fmt"
	"github.com/zllangct/rockgo/ecs"
	"github.com/zllangct/rockgo/logger"
)

/*
	Component组
	ComponentGroup 一般按照分布式思想，同一功能节点，分为一组。
	比如，网关组、大厅组、逻辑房间、位置服务等
*/
type ComponentGroup struct {
	Name    string
	content []ecs.IComponent
}

func (this *ComponentGroup) attachGroupTo(target *ecs.Object) {
	o := ecs.NewObject(this.Name)
	err := target.AddObject(o)
	if err != nil {
		logger.Error(err)
	}
	for _, component := range this.content {
		o.AddComponent(component)
		logger.Info(fmt.Sprintf("Attach ecs [ %s.%s ] to [ %s ]", this.Name, component.Type().String(), o.Name()))
	}
}

/*
	所有可用Component组
*/
type ComponentGroups struct {
	group map[string]*ComponentGroup //key:group name , value:ecs group
}

func (this *ComponentGroups) AllGroups() map[string]*ComponentGroup {
	if this.group == nil {
		this.group = make(map[string]*ComponentGroup)
	}
	return this.group
}

func (this *ComponentGroups) AllGroupsName() []string {
	if this.group == nil {
		this.group = make(map[string]*ComponentGroup)
	}
	arr := make([]string, 0)
	for role, _ := range this.group {
		arr = append(arr, role)
	}
	return arr
}

func (this *ComponentGroups) AddGroup(groupName string, group []ecs.IComponent) {
	if this.group == nil {
		this.group = make(map[string]*ComponentGroup)
	}
	this.group[groupName] = &ComponentGroup{
		Name:    groupName,
		content: group,
	}
}

func (this *ComponentGroups) AttachGroupsTo(groupName []string, target *ecs.Object) error {
	child, master, other := false, false, false
	for _, name := range groupName {
		switch name {
		case "master":
			master = true
		case "child":
			child = true
		default:
			other = true
		}
	}
	//为空时，默认为master
	if !other && !master && !child {
		groupName = append(groupName, "master")
	}
	//有其他角色是，需要child
	if other && !child {
		groupName = append(groupName, "child")
	}
	//有master，没有其他的时候，不需要child
	if !other && master && child {
		for i, v := range groupName {
			if v == "child" {
				groupName = append(groupName[:i], groupName[i+1:]...)
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
