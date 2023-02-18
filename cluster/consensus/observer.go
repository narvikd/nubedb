package consensus

import (
	"fmt"
	"github.com/hashicorp/raft"
	"time"
)

func (n *Node) registerObservers() {
	n.registerNodeChangesChan()
	n.registerLeaderChangesChan()
	n.registerFailedHBChangesChan()
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

func (n *Node) registerFailedHBChangesChan() {
	n.failedHBChangesChan = make(chan raft.Observation)
	observer := raft.NewObserver(n.failedHBChangesChan, true, func(o *raft.Observation) bool {
		_, ok := o.Data.(raft.FailedHeartbeatObservation)
		return ok
	})
	n.Consensus.RegisterObserver(observer)

	// Call methods
	n.logNewHBChange()
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

func (n *Node) logNewHBChange() {
	go func() {
		for obs := range n.failedHBChangesChan {
			dataType := obs.Data.(raft.FailedHeartbeatObservation)
			duration := time.Since(dataType.LastContact)
			msg := fmt.Sprintf("HB FAILED FOR NODE '%v' for %s seconds ",
				dataType.PeerID, duration.Round(time.Second).String(),
			)
			n.consensusLogger.Info("HB FAILED: " + msg)
		}
	}()
}
