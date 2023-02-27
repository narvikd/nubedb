package consensus

import (
	"github.com/hashicorp/raft"
	"math"
	"nubedb/discover"
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

	if !n.IsQuorumPossible(false) {
		n.logger.Warn(prefixErr + "quorum is not possible due to lack of available nodes")
		return false
	}

	return true
}

// IsQuorumPossible checks if a quorum is possible based on the number of available nodes.
//
// In the raft consensus algorithm a quorum can be reached when (onlineNodes >= (consensusNodes/2)+1)
//
// To be in the safe side, it will return as false without the added margin.
//
// preAlert changes the formula to return false before the number of nodes needed for the quorum would become insufficient.
//
// Example:
//
// preAlert to false -> For 5 nodes, 3 are required.
//
// preAlert to true -> For 5 nodes, 4 are required.
//
// preAlert to false -> For 10 nodes, 5 are required.
//
// preAlert to true -> For 10 nodes, 8 are required.
func (n *Node) IsQuorumPossible(preAlert bool) bool {
	const (
		half       = 0.5
		percentile = 0.75
	)
	// Safe to ignore this error since it will always return a number
	consensusNodes, _ := strconv.ParseFloat(n.Consensus.Stats()["num_peers"], 64)
	necessaryForQuorum := math.Round(consensusNodes * half)
	if preAlert {
		necessaryForQuorum = math.Round(consensusNodes * percentile)
	}
	onlineNodes := len(discover.SearchAliveNodes(n.Consensus, n.ID))

	return float64(onlineNodes) >= necessaryForQuorum
}
