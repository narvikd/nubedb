package discover

import (
	"errors"
	"github.com/hashicorp/raft"
	"nubedb/discover/discovercommons"
	"nubedb/discover/mdnsdiscover"
	"nubedb/discover/sequentialdiscover"
	"nubedb/internal/config"
	"nubedb/pkg/ipkit"
)

var mode = config.DiscoverDefault

func SetMode(m string) error {
	if mode != config.DiscoverDefault && mode != config.DiscoverSequential {
		return errors.New("discover mode not recognized")
	}
	mode = m
	return nil
}

// SearchAliveNodes will skip currentNodeID.
func SearchAliveNodes(consensus *raft.Raft, currentNodeID string) []raft.Server {
	return discovercommons.SearchAliveNodes(consensus, currentNodeID)
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
	case config.DiscoverSequential:
		return sequentialdiscover.SearchLeader(currentNode)
	default:
		return "", nil
	}
}

func NewApiAddr(nodeHost string) string {
	switch mode {
	case config.DiscoverDefault:
		return ipkit.NewAddr(nodeHost, config.ApiPort)
	case config.DiscoverSequential:
		return sequentialdiscover.NewApiAddr(nodeHost)
	default:
		return ""
	}
}

func NewConsensusAddr(nodeHost string) string {
	switch mode {
	case config.DiscoverDefault:
		return ipkit.NewAddr(nodeHost, config.ConsensusPort)
	case config.DiscoverSequential:
		return sequentialdiscover.NewConsensusAddr(nodeHost)
	default:
		return ""
	}
}

func NewGrpcAddress(nodeHost string) string {
	switch mode {
	case config.DiscoverDefault:
		return ipkit.NewAddr(nodeHost, config.GrpcPort)
	case config.DiscoverSequential:
		return sequentialdiscover.NewGrpcAddress(nodeHost)
	default:
		return ""
	}
}
