package Cluster

import (
	"errors"
	"strings"
	"sync"
)

const (
	SELECTOR_TYPE_DEFAULT  SelectorType = "Default"
	SELECTOR_TYPE_MIN_LOAD SelectorType = "MinLoad"
)

type SelectorType string

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

func (this Selector) Select(query string,detail bool,locker *sync.RWMutex) ([]*InquiryReply ,error){
	args := strings.Split(query, ":")
	length:=len(args)
	if length < 2 || length>3 || args[0]=="" || args[1] ==""{
		return nil, errors.New("query string wrong")
	}
	selector := SELECTOR_TYPE_DEFAULT
	if length==3 {
		selector = SelectorType(args[2])
	}

	err := errors.New("no available node ")
	var reply = make([]*InquiryReply,0)
	locker.RLock()
	for nodeName, nodeInfo := range master.Nodes {
		if nodeInfo.AppName == args[0] {
			for _, role := range nodeInfo.Group {
				if role == args[1] {
					if detail {
						reply = append(reply, &InquiryReply{Node: nodeName,Info:nodeInfo.Info})
					}else{
						reply = append(reply, &InquiryReply{Node: nodeName})
					}
					err = nil
					break
				}
			}
			break
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
	default:

	}
	return reply, err
}