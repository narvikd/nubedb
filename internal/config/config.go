package config

import (
	"fmt"
	"github.com/narvikd/resolver"
	"os"
	"strings"
	"time"
)

type NodeCfg struct {
	ID               string
	ApiPort          int
	ApiAddress       string
	ConsensusPort    int
	ConsensusAddress string
	GrpcPort         int
	GrpcAddress      string
}

type Config struct {
	CurrentNode NodeCfg
	Nodes       map[string]NodeCfg
}

func New() (Config, error) {
	const resolverTimeout = 300 * time.Millisecond
	currentNodeID := os.Getenv("NODE")

	if !strings.Contains(currentNodeID, "node") {
		return Config{}, fmt.Errorf("NODE env variable not set or is incorrect: %s", currentNodeID)
	}

	if !resolver.IsHostAlive(currentNodeID, resolverTimeout) {
		return Config{}, fmt.Errorf("no host found for: %s", currentNodeID)
	}

	nodes := make(map[string]NodeCfg)
	for i := 1; i <= 3; i++ {
		n := fmt.Sprintf("node%v", i)
		nodes[n] = newNodeCfg(n)
	}

	currentNode, exist := nodes[currentNodeID]
	if !exist {
		return Config{}, fmt.Errorf("current node not found in nodes: %s", currentNodeID)
	}

	cfg := Config{
		CurrentNode: currentNode,
		Nodes:       nodes,
	}
	return cfg, nil
}

func newNodeCfg(nodeID string) NodeCfg {
	const (
		apiPort       = 3001
		consensusPort = 3002
		grpcPort      = 3003
	)

	return NodeCfg{
		ID:               nodeID,
		ApiPort:          apiPort,
		ApiAddress:       makeAddr(nodeID, apiPort),
		ConsensusPort:    consensusPort,
		ConsensusAddress: makeAddr(nodeID, consensusPort),
		GrpcPort:         grpcPort,
		GrpcAddress:      makeAddr(nodeID, grpcPort),
	}
}

func makeAddr(host string, port int) string {
	return fmt.Sprintf("%s:%v", host, port)
}
