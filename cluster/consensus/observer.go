package consensus

import "github.com/hashicorp/raft"

func (n *Node) registerObservers() {
	n.registerNodeChangesChan()
	n.registerLeaderChangesChan()
}

func (n *Node) registerNodeChangesChan() {
	n.nodeChangesChan = make(chan raft.Observation)
	observer := raft.NewObserver(n.nodeChangesChan, true, func(o *raft.Observation) bool {
		_, ok := o.Data.(raft.RaftState)
		return ok
	})
	n.Consensus.RegisterObserver(observer)

	// Call methods
	n.logNewNodeChange()
}

func (n *Node) registerLeaderChangesChan() {
	n.leaderChangesChan = make(chan raft.Observation)
	observer := raft.NewObserver(n.leaderChangesChan, true, func(o *raft.Observation) bool {
		_, ok := o.Data.(raft.LeaderObservation)
		return ok
	})
	n.Consensus.RegisterObserver(observer)

	// Call methods
	n.logNewLeader()
}

func (n *Node) logNewNodeChange() {
	go func() {
		for obs := range n.nodeChangesChan {
			dataType := obs.Data.(raft.RaftState)
			n.consensusLogger.Info("Node Changed to role: " + dataType.String())
		}
	}()
}

func (n *Node) logNewLeader() {
	go func() {
		for obs := range n.leaderChangesChan {
			dataType := obs.Data.(raft.LeaderObservation)
			leaderID := string(dataType.LeaderID)
			if leaderID != "" {
				n.consensusLogger.Info("New Leader: " + string(dataType.LeaderID))
			} else {
				n.consensusLogger.Info("No Leader available in the Cluster")
			}
		}
	}()
}
