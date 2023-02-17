package config

import (
	"fmt"
	"github.com/narvikd/resolver"
	"os"
	"strings"
	"time"
)

const (
	ApiPort       = 3001
	ConsensusPort = 3002
	GrpcPort      = 3003
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

	cfg := Config{CurrentNode: NewNodeCfg(currentNodeID)}
	return cfg, nil
}

func NewNodeCfg(nodeID string) NodeCfg {
	return NodeCfg{
		ID:               nodeID,
		ApiPort:          ApiPort,
		ApiAddress:       MakeApiAddr(nodeID),
		ConsensusPort:    ConsensusPort,
		ConsensusAddress: MakeConsensusAddr(nodeID),
		GrpcPort:         GrpcPort,
		GrpcAddress:      MakeGrpcAddress(nodeID),
	}
}

func MakeApiAddr(nodeID string) string {
	return makeAddr(nodeID, ApiPort)
}

func MakeConsensusAddr(nodeID string) string {
	return makeAddr(nodeID, ConsensusPort)
}

func MakeGrpcAddress(nodeID string) string {
	return makeAddr(nodeID, GrpcPort)
}

func makeAddr(host string, port int) string {
	return fmt.Sprintf("%s:%v", host, port)
}
