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
		for o := range n.nodeChangesChan {
			n.consensusLogger.Info("Node Changed to role: " + o.Data.(raft.RaftState).String())
		}
	}()
}

func (n *Node) logNewLeader() {
	go func() {
		for o := range n.leaderChangesChan {
			obs := o.Data.(raft.LeaderObservation)
			leaderID := string(obs.LeaderID)
			if leaderID != "" {
				n.consensusLogger.Info("New Leader: " + leaderID)
			} else {
				n.consensusLogger.Info("No Leader available in the Cluster")
			}
		}
	}()
}

func (n *Node) logNewHBChange() {
	go func() {
		for o := range n.failedHBChangesChan {
			obs := o.Data.(raft.FailedHeartbeatObservation)
			duration := time.Since(obs.LastContact)
			msg := fmt.Sprintf("HB FAILED FOR NODE '%v' for %s seconds ",
				obs.PeerID, duration.Round(time.Second).String(),
			)
			n.consensusLogger.Info("HB FAILED: " + msg)
		}
	}()
}
