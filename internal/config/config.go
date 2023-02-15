package config

import (
	"fmt"
	"log"
	"os"
	"strings"
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

func New() Config {
	currentNodeID := os.Getenv("NODE")
	if !strings.Contains(currentNodeID, "node") {
		log.Fatalln("NODE ENV not set or is incorrect", currentNodeID)
	}

	nodes := make(map[string]NodeCfg)
	for i := 1; i <= 3; i++ {
		n := fmt.Sprintf("node%v", i)
		nodes[n] = newNodeCfg(n)
	}

	currentNode, exist := nodes[currentNodeID]
	if !exist {
		log.Fatalln("current node not found in nodes:", currentNodeID)
	}

	return Config{
		CurrentNode: currentNode,
		Nodes:       nodes,
	}
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
