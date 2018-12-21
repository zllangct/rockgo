package Cluster

import (
	"errors"
	"sync"
)

const (
	SELECTOR_TYPE_GROUP  SelectorType = "Group"
	SELECTOR_TYPE_DEFAULT  SelectorType = "Default"
	SELECTOR_TYPE_MIN_LOAD SelectorType = "MinLoad"
)

type SelectorType = string

type SourceGroup []*InquiryReply

//最小负载：cpu * 80% + mem * 20%
func (this SourceGroup)SelectMinLoad() int {
	var min float32 =1
	var index int = -1
	for i, info := range this {
		var cpu,mem float32 = 1,1
		if v,ok:=info.Info["cpu"];ok{
			cpu=v
		}
		if v,ok:=info.Info["mem"];ok{
			mem=v
		}
		sum := cpu * 0.8 + mem * 0.2
		if sum <= min {
			min = sum
			index=i
		}
	}
	return index
}

type Selector map[string]*NodeInfo
var ErrNoAvailableNode = errors.New("query string wrong")
// 0 选择模式 1 AppName 2 role
func (this Selector) Select(query []string,detail bool,locker *sync.RWMutex) ([]*InquiryReply ,error){
	length:=len(query)
	if length !=3 || query[0]=="" {
		return nil,ErrNoAvailableNode
	}
	selector := SELECTOR_TYPE_DEFAULT

	err := errors.New("no available node ")
	var reply = make([]*InquiryReply,0)
	locker.RLock()
	for nodeName, nodeInfo := range this {
		if nodeInfo.AppName == query[1] {
			for _, role := range nodeInfo.Role {
				if  role == query[2] {
					if detail {
						reply = append(reply, &InquiryReply{Node: nodeName,Info:nodeInfo.Info})
					}else{
						reply = append(reply, &InquiryReply{Node: nodeName})
					}
					err = nil
					break
				}
			}
			if err==nil {
				break
			}
		}
	}
	locker.RUnlock()

	switch selector {
	case SELECTOR_TYPE_DEFAULT, SELECTOR_TYPE_MIN_LOAD:
		var index  = -1
		index = SourceGroup(reply).SelectMinLoad()
		if index != -1{
			reply = []*InquiryReply{ reply[index] }
		}
	case SELECTOR_TYPE_GROUP:

	default:

	}
	return reply, err
}