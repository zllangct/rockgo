package Cluster

import (
	"errors"
	"github.com/zllangct/RockGO/rpc"
	"math/rand"
)

type NodeID struct {
	addr string
	nodeComponent *NodeComponent
}

func (this *NodeID) GetClient() (*rpc.TcpClient,error)  {
	if this.addr == "" {
		return nil,errors.New("this node id is empty")
	}
	return this.nodeComponent.GetNodeClient(this.addr)
}

type NodeIDGroup struct {
	nodeComponent *NodeComponent
	nodes []*InquiryReply
}

//所有节点，仅地址
func (this *NodeIDGroup) Nodes() []string {
	nodes:=make([]string,len(this.nodes))
	for _, v := range this.nodes {
		nodes= append(nodes, v.Node)
	}
	return nodes
}

//所有节点，详细信息
func (this *NodeIDGroup) NodesDetail() []*InquiryReply {
	return this.nodes
}

//随机选择一个
func (this *NodeIDGroup)RandClient() (*rpc.TcpClient,error) {
	length:=len(this.nodes)
	if length == 0 {
		return nil,errors.New("this node id group is empty")
	}
	index:=rand.Intn(length)

	return this.nodeComponent.GetNodeClient(this.nodes[index].Node)
}

//选择一个负载最低的节点
func (this *NodeIDGroup)MinLoadClient() (*rpc.TcpClient,error) {
	if len(this.nodes) == 0 {
		return nil,errors.New("this node id group is empty")
	}
	index:=SourceGroup(this.nodes).SelectMinLoad()
	return this.nodeComponent.GetNodeClient(this.nodes[index].Node)
}