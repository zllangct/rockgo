package Cluster_test

import (
	"github.com/zllangct/RockGO/Cluster"
	"testing"
)

func TestNewServerNode(t *testing.T) {
	node:=Cluster.NewServerNode()
	node.Serve()
}