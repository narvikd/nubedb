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
	const resolverTimeout = 300 * time.Millisecond
	hostname, errHostname := os.Hostname()
	if errHostname != nil {
		return Config{}, errorskit.Wrap(errHostname, "couldn't get hostname on Config generation")
	}

	if !resolver.IsHostAlive(hostname, resolverTimeout) {
		return Config{}, fmt.Errorf("no host found for: %s", hostname)
	}

	nodeCfg := NodeCfg{
		ID:                 hostname,
		ApiPort:            ApiPort,
		ApiAddress:         MakeApiAddr(hostname),
		ConsensusPort:      ConsensusPort,
		ConsensusAddress:   MakeConsensusAddr(hostname),
		GrpcPort:           GrpcPort,
		GrpcAddress:        MakeGrpcAddress(hostname),
		FSMPerformanceMode: strings.ToLower(os.Getenv("FSM_PERFORMANCE")) == "true",
	}
	return Config{CurrentNode: nodeCfg}, nil
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
