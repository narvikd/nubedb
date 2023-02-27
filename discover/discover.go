package discover

import (
	"errors"
	"github.com/hashicorp/raft"
	"nubedb/discover/mdnsdiscover"
	"nubedb/internal/config"
	"nubedb/pkg/ipkit"
)

var mode = config.DiscoverDefault

func SetMode(m string) error {
	if mode != config.DiscoverDefault && mode != config.DiscoverNubeRegistry {
		return errors.New("discover mode not recognized")
	}
	mode = m
	return nil
}

func GetAliveNodes(consensus *raft.Raft, currentNodeID string) []raft.Server {
	switch mode {
	case config.DiscoverDefault:
		return mdnsdiscover.GetAliveNodes(consensus, currentNodeID)
	default:
		return []raft.Server{}
	}
}

func NewConsensusAddr(nodeHost string) string {
	switch mode {
	case config.DiscoverDefault:
		return ipkit.NewAddr(nodeHost, config.ConsensusPort)
	default:
		return ""
	}
}

func NewGrpcAddress(nodeHost string) string {
	switch mode {
	case config.DiscoverDefault:
		return ipkit.NewAddr(nodeHost, config.GrpcPort)
	default:
		return ""
	}
}
