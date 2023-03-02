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
	if mode != config.DiscoverDefault && mode != config.DiscoverNubeWatch {
		return errors.New("discover mode not recognized")
	}
	mode = m
	return nil
}

// SearchAliveNodes will skip currentNodeID.
func SearchAliveNodes(consensus *raft.Raft, currentNodeID string) []raft.Server {
	switch mode {
	case config.DiscoverDefault:
		return mdnsdiscover.SearchAliveNodes(consensus, currentNodeID)
	default:
		return []raft.Server{}
	}
}

// SearchLeader will return an error if a leader is not found,
// since it skips the current node and this could be a leader.
//
// Because it skips the current node, it will still return an error.
//
// This is done this way to ensure this function is never called to do gRPC operations in itself.
func SearchLeader(currentNode string) (string, error) {
	switch mode {
	case config.DiscoverDefault:
		return mdnsdiscover.SearchLeader(currentNode)
	default:
		return "", nil
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
