package consensus

import (
	"github.com/hashicorp/raft"
	"math"
	"nubedb/pkg/resolver"
	"strconv"
	"time"
)

func (n *Node) IsHealthy() bool {
	const genErr = "[NODE UNHEALTHY] - "
	cfg := n.Consensus.GetConfiguration().Configuration()
	consensusServers := cfg.Servers
	stats := n.Consensus.Stats()

	if len(consensusServers) <= 1 {
		n.consensusLogger.Warn(genErr + "only one server in configuration")
		return false
	}

	if n.Consensus.State() != raft.Leader && stats["last_contact"] == "never" {
		n.consensusLogger.Warn(genErr + "only one server in configuration")
		return false
	}

	leaderAddr, leaderID := n.Consensus.LeaderWithID()
	if leaderAddr == "" {
		n.consensusLogger.Warn(genErr + "local consensus reports that leader address is empty")
		return false
	}

	if leaderID == "" {
		n.consensusLogger.Warn(genErr + "local consensus reports that leader id is empty")
		return false
	}

	isQuorumPossible, errQuorum := n.isQuorumPossible()
	if errQuorum != nil {
		n.consensusLogger.Error(genErr + "couldn't check if quorum was possible")
		return false
	}

	if !isQuorumPossible {
		n.consensusLogger.Warn(genErr + "quorum is not possible due to lack of available nodes")
		return false
	}

	return true
}

func (n *Node) isQuorumPossible() (bool, error) {
	peers, _ := strconv.Atoi(n.Consensus.Stats()["num_peers"]) // Safe to ignore this error
	necessaryForQuorum := math.Ceil(float64(peers) / 2.0)
	totalNodes, err := n.getAliveNodes()
	if err != nil {
		return false, err
	}
	return float64(totalNodes) >= necessaryForQuorum, nil
}

func (n *Node) getAliveNodes() (int, error) {
	const timeout = 300 * time.Millisecond
	counter := 0
	liveCfg := n.Consensus.GetConfiguration().Configuration()
	cfg := liveCfg.Clone() // Clone CFG to not keep calling it in the for, in case the num of servers is very large
	for _, srv := range cfg.Servers {
		srvID := string(srv.ID)
		if n.ID == srvID {
			continue
		}
		if resolver.IsHostAlive(srvID, timeout) {
			counter++
		}
	}

	totalNodes := counter - 1
	return totalNodes, nil
}
