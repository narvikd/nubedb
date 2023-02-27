package config

import (
	"fmt"
	"github.com/narvikd/errorskit"
	"nubedb/pkg/resolver"
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
	ID                 string
	ApiPort            int
	ApiAddress         string
	ConsensusPort      int
	ConsensusAddress   string
	GrpcPort           int
	GrpcAddress        string
	FSMPerformanceMode bool
}

type Config struct {
	CurrentNode NodeCfg
}

func New() (Config, error) {
	nodeID, err := newNodeID()
	if err != nil {
		return Config{}, err
	}

	nodeCfg := NodeCfg{
		ID:                 nodeID,
		ApiPort:            ApiPort,
		ApiAddress:         MakeApiAddr(nodeID),
		ConsensusPort:      ConsensusPort,
		ConsensusAddress:   MakeConsensusAddr(nodeID),
		GrpcPort:           GrpcPort,
		GrpcAddress:        MakeGrpcAddress(nodeID),
		FSMPerformanceMode: strings.ToLower(os.Getenv("FSM_PERFORMANCE")) == "true",
	}
	return Config{CurrentNode: nodeCfg}, nil
}

func newNodeID() (string, error) {
	const resolverTimeout = 300 * time.Millisecond
	hostname, errHostname := os.Hostname()
	if errHostname != nil {
		return "", errorskit.Wrap(errHostname, "couldn't get hostname on Config generation")
	}
	if !resolver.IsHostAlive(hostname, resolverTimeout) {
		return "", fmt.Errorf("no host found for: %s", hostname)
	}
	return hostname, nil
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
