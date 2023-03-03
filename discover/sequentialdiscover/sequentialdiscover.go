package sequentialdiscover

import (
	"errors"
	"fmt"
	"github.com/narvikd/errorskit"
	"log"
	"nubedb/discover/discovercommons"
	"nubedb/internal/config"
	"nubedb/pkg/ipkit"
	"strconv"
	"strings"
)

const ErrLeaderNotFound = "couldn't find a leader"

// SearchLeader will return an error if a leader is not found,
// since it skips the current node and this could be a leader.
//
// Because it skips the current node, it will still return an error.
//
// This is done this way to ensure this function is never called to do gRPC operations in itself.
func SearchLeader(currentNode string) (string, error) {
	var nodes []string
	for i := 1; i <= config.DiscoverSequentialMaxNodes; i++ {
		nodeName := fmt.Sprintf("nubedb-node%v", i)
		if i == 1 && nodeName != currentNode {
			nodes = append(nodes, config.BootstrappingLeader)
			continue
		}
		if nodeName == currentNode {
			continue
		}
		nodes = append(nodes, nodeName)
	}

	for _, node := range nodes {
		grpcAddr := NewGrpcAddress(node)
		leader, err := discovercommons.IsLeader(grpcAddr)
		if err != nil {
			errorskit.LogWrap(err, "couldn't contact node while searching for leaders")
			continue
		}
		if leader {
			return node, nil
		}
	}

	return "", errors.New(ErrLeaderNotFound)
}

func NewApiAddr(nodeHost string) string {
	const defaultPort = 10011

	if nodeHost == config.BootstrappingLeader {
		return ipkit.NewAddr(nodeHost, defaultPort)
	}

	if nodeHost == config.NodeID {
		return ipkit.NewAddr(nodeHost, config.ApiPort)
	}

	s := strings.Split(nodeHost, "nubedb-node")
	if len(s) <= 0 {
		log.Fatalln("couldn't create consensus addr")
	}

	n, err := strconv.Atoi(s[1])
	if err != nil {
		log.Fatalln("couldn't create consensus addr. Because nodeHost didn't had a number in the end")
	}

	port := defaultPort + (10 * n)

	return ipkit.NewAddr(nodeHost, port)
}

func NewConsensusAddr(nodeHost string) string {
	const defaultPort = 10012

	if nodeHost == config.BootstrappingLeader {
		return ipkit.NewAddr(nodeHost, defaultPort)
	}

	if nodeHost == config.NodeID {
		return ipkit.NewAddr(nodeHost, config.ConsensusPort)
	}

	s := strings.Split(nodeHost, "nubedb-node")
	if len(s) <= 0 {
		log.Fatalln("couldn't create consensus addr")
	}

	n, err := strconv.Atoi(s[1])
	if err != nil {
		log.Fatalln("couldn't create consensus addr. Because nodeHost didn't had a number in the end")
	}

	port := defaultPort + (10 * n)

	return ipkit.NewAddr(nodeHost, port)
}

func NewGrpcAddress(nodeHost string) string {
	const defaultPort = 10013

	if nodeHost == config.BootstrappingLeader {
		return ipkit.NewAddr(nodeHost, defaultPort)
	}

	if nodeHost == config.NodeID {
		return ipkit.NewAddr(nodeHost, config.GrpcPort)
	}

	s := strings.Split(nodeHost, "nubedb-node")
	if len(s) <= 0 {
		log.Fatalln("couldn't create grpc addr")
	}

	n, err := strconv.Atoi(s[1])
	if err != nil {
		log.Fatalln("couldn't create grpc addr. Because nodeHost didn't had a number in the end")
	}

	port := defaultPort + (10 * n)

	return ipkit.NewAddr(nodeHost, port)
}
