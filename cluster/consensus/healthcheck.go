package consensus

import "github.com/hashicorp/raft"

func (n *Node) IsHealthy() bool {
	const genErr = "[NODE UNHEALTHY] - "
	fCfg := n.Consensus.GetConfiguration()
	if fCfg.Error() != nil {
		n.LogWrapErr(fCfg.Error(), "couldn't get consensus configuration")
		return false
	}
	hotCfg := fCfg.Configuration()
	cfg := hotCfg.Clone()

	consensusServers := cfg.Servers
	if len(consensusServers) <= 1 {
		n.consensusLogger.Warn(genErr + "only one server in configuration")
		return false
	}

	if n.Consensus.State() != raft.Leader && n.Consensus.LastContact().IsZero() {
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

	return true
}
