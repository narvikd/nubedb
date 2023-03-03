package config

import (
	"fmt"
	"github.com/narvikd/errorskit"
	"log"
	"nubedb/pkg/resolver"
	"os"
	"strconv"
	"time"
)

var (
	BootstrappingLeader        = "bootstrap-node"
	NodeID                     = "bootstrap-node"
	ApiPort                    = 3001
	ConsensusPort              = 3002
	GrpcPort                   = 3003
	DiscoverPort               = 3004
	DiscoverSequentialMaxNodes = 15
)

const (
	DiscoverDefault    = "default"
	DiscoverSequential = "sequential"
	errNaN             = "is not a number"
)

type Config struct {
	CurrentNode Node
	Cluster     Cluster
}

type Node struct {
	ID            string
	ApiPort       int
	ConsensusPort int
	GrpcPort      int
}

type Cluster struct {
	FSMPerformanceMode bool
	DiscoverStrategy   string
}

func New() (Config, error) {
	setPorts()
	nodeHost, err := newNodeID()
	if err != nil {
		return Config{}, err
	}

	NodeID = nodeHost

	nodeCfg := Node{
		ID:            nodeHost,
		ApiPort:       ApiPort,
		ConsensusPort: ConsensusPort,
		GrpcPort:      GrpcPort,
	}

	return Config{CurrentNode: nodeCfg, Cluster: newClusterCfg()}, nil
}

func setPorts() {
	apiPort := os.Getenv("API_PORT")
	consensusPort := os.Getenv("CONSENSUS_PORT")
	grpcPort := os.Getenv("GRPC_PORT")
	discoverPort := os.Getenv("DISCOVER_PORT")

	if apiPort != "" {
		n, err := strconv.Atoi(apiPort)
		if err != nil {
			log.Fatalln("API_PORT " + errNaN)
		}
		ApiPort = n
	}

	if consensusPort != "" {
		n, err := strconv.Atoi(consensusPort)
		if err != nil {
			log.Fatalln("CONSENSUS_PORT " + errNaN)
		}
		ConsensusPort = n
	}

	if grpcPort != "" {
		n, err := strconv.Atoi(grpcPort)
		if err != nil {
			log.Fatalln("GRPC_PORT " + errNaN)
		}
		GrpcPort = n
	}

	if discoverPort != "" {
		n, err := strconv.Atoi(discoverPort)
		if err != nil {
			log.Fatalln("DISCOVER_PORT " + errNaN)
		}
		DiscoverPort = n
	}
}

func newNodeID() (string, error) {
	const resolverTimeout = 300 * time.Millisecond
	bootstrappingLeader := os.Getenv("BOOTSTRAPPING_LEADER")
	overrideID := os.Getenv("ID")

	if bootstrappingLeader != "" {
		BootstrappingLeader = bootstrappingLeader
	}

	id, errHostname := os.Hostname()
	if errHostname != nil {
		return "", errorskit.Wrap(errHostname, "couldn't get hostname on Config generation")
	}

	if overrideID != "" {
		id = overrideID
	}

	if !resolver.IsHostAlive(id, resolverTimeout) {
		return "", fmt.Errorf("no host found for: %s", id)
	}
	return id, nil
}

func newClusterCfg() Cluster {
	discoverStrategy := os.Getenv("DISCOVER_STRATEGY")
	discoverSequentialMaxNodes := os.Getenv("DISCOVER_SEQUENTIAL_MAX_NODES")
	fsmPerformance := os.Getenv("FSM_PERFORMANCE")

	clusterCfg := Cluster{
		FSMPerformanceMode: fsmPerformance == "true",
		DiscoverStrategy:   DiscoverDefault,
	}

	if discoverStrategy != "" {
		clusterCfg.DiscoverStrategy = discoverStrategy
	}

	if discoverSequentialMaxNodes != "" {
		n, err := strconv.Atoi(discoverSequentialMaxNodes)
		if err != nil {
			log.Fatalln("API_PORT " + errNaN)
		}
		DiscoverSequentialMaxNodes = n
	}

	return clusterCfg
}
