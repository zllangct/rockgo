package RockGO

import (
	"github.com/zllangct/RockGO/cluster"
)

func NewServer() *Cluster.ServerNode {
	return Cluster.DefaultNode()
}