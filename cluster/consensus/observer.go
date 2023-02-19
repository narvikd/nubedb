package consensus

import (
	"fmt"
	"github.com/hashicorp/raft"
	"github.com/narvikd/errorskit"
	"github.com/narvikd/filekit"
	"log"
	"nubedb/cluster"
	"nubedb/discover"
	"nubedb/internal/config"
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
	n.onNodeChange()
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
	n.removeNodesOnHBStrategy()
}

func (n *Node) onNodeChange() {
	go func() {
		for o := range n.nodeChangesChan {
			n.consensusLogger.Info("Node Changed to role: " + o.Data.(raft.RaftState).String())
			const timeout = 20 * time.Second
			// If the node is stuck in a candidate position for 20 seconds, it's blocked.
			// There's no scenario where the node is a candidate for more than 10 seconds,
			// and it's not part of a bootstrapped service.
			if n.Consensus.State() == raft.Candidate {
				time.Sleep(timeout)
				// If a minute has passed, and I'm still blocked, there's a real problem.
				if n.Consensus.State() == raft.Candidate {
					n.unblockCandidate()
				}
			}
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

func (n *Node) removeNodesOnHBStrategy() {
	go func() {
		const timeout = 1.0
		for o := range n.failedHBChangesChan {
			obs := o.Data.(raft.FailedHeartbeatObservation)
			duration := time.Since(obs.LastContact)
			durationMins := duration.Minutes()

			if durationMins >= timeout {
				warnMsg := fmt.Sprintf(
					"REMOVING NODE '%v' from the Leader due to not having a connection for %v minutes...",
					obs.PeerID, durationMins,
				)
				n.consensusLogger.Warn(warnMsg)
				n.Consensus.RemoveServer(obs.PeerID, 0, 0)
				n.consensusLogger.Warn("NODE SUCCESSFULLY REMOVED FROM STATE CONSENSUS")
			}

		}
	}()
}

func (n *Node) unblockCandidate() {
	const errPanic = "COULDN'T GRACEFULLY UNBLOCK CANDIDATE. "

	log.Println("node got stuck in candidate for too long... Node reinstall in progress...")

	future := n.Consensus.Shutdown()
	if future.Error() != nil {
		errorskit.FatalWrap(future.Error(), errPanic+"couldn't shut down")
	}

	leader, errSearchLeader := discover.SearchLeader(n.ID)
	if errSearchLeader != nil {
		errorskit.FatalWrap(errSearchLeader, errPanic+"couldn't search for leader")
	}
	leaderGrpcAddress := config.MakeGrpcAddress(leader)

	errConsensusRemove := cluster.ConsensusRemove(n.ID, leaderGrpcAddress)
	if errConsensusRemove != nil {
		errorskit.FatalWrap(errConsensusRemove, errPanic+"couldn't remove from consensus")
	}

	errDeleteDirs := filekit.DeleteDirs(n.Dir)
	if errDeleteDirs != nil {
		errorskit.FatalWrap(errDeleteDirs, errPanic+"couldn't delete dirs")
	}

	log.Fatalln("Node successfully reset. Restarting...")
}
