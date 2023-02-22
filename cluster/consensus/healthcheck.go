package consensus

import (
	"github.com/hashicorp/raft"
	"github.com/narvikd/errorskit"
	"math"
	"nubedb/cluster"
	"strconv"
)

// IsHealthy checks if the node and the cluster are in a healthy state by validating various factors.
func (n *Node) IsHealthy() bool {
	const prefixErr = "[NODE UNHEALTHY] - "

	consensusServers := n.Consensus.GetConfiguration().Configuration().Servers
	stats := n.Consensus.Stats()

	if len(consensusServers) <= 1 {
		n.logger.Warn(prefixErr + "only one server in configuration")
		return false
	}

	if n.Consensus.State() != raft.Leader && stats["last_contact"] == "never" {
		n.logger.Warn(prefixErr + "only one server in configuration")
		return false
	}

	leaderAddr, leaderID := n.Consensus.LeaderWithID()
	if leaderAddr == "" {
		n.logger.Warn(prefixErr + "local consensus reports that leader address is empty")
		return false
	}

	if leaderID == "" {
		n.logger.Warn(prefixErr + "local consensus reports that leader id is empty")
		return false
	}

	isQuorumPossible, errQuorum := n.isQuorumPossible()
	if errQuorum != nil {
		errWrap := errorskit.Wrap(errQuorum, prefixErr)
		n.logger.Error(errWrap.Error())
		return false
	}

	if !isQuorumPossible {
		n.logger.Warn(prefixErr + "quorum is not possible due to lack of available nodes")
		return false
	}

	return true
}

// isQuorumPossible checks if a quorum is possible based on the number of available nodes.
func (n *Node) isQuorumPossible() (bool, error) {
	peers, _ := strconv.Atoi(n.Consensus.Stats()["num_peers"]) // Safe to ignore this error
	necessaryForQuorum := math.Ceil(float64(peers) / 2.0)

	totalNodes, err := n.getNumAliveNodes()
	if err != nil {
		return false, err
	}

	return float64(totalNodes) >= necessaryForQuorum, nil
}

// getNumAliveNodes gets the number of alive nodes in the cluster
func (n *Node) getNumAliveNodes() (int, error) {
	nodes, err := cluster.GetAliveNodes(n.Consensus, n.ID)
	if err != nil {
		return 0, err
	}
	return len(nodes), nil
}
