package Cluster_test

import (
	"github.com/zllangct/RockGO/cluster"
	"testing"
)

func TestNewServerNode(t *testing.T) {
	node:=Cluster.DefaultNode()
	node.Serve()
}