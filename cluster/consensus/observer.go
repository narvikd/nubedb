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
	n.nodeChangesChan = make(chan raft.Observation, 4)
	observer := raft.NewObserver(n.nodeChangesChan, false, func(o *raft.Observation) bool {
		_, ok := o.Data.(raft.RaftState)
		return ok
	})
	n.Consensus.RegisterObserver(observer)

	// Call methods
	n.onNodeChange()
}

func (n *Node) registerLeaderChangesChan() {
	n.leaderChangesChan = make(chan raft.Observation, 4)
	observer := raft.NewObserver(n.leaderChangesChan, false, func(o *raft.Observation) bool {
		_, ok := o.Data.(raft.LeaderObservation)
		return ok
	})
	n.Consensus.RegisterObserver(observer)

	// Call methods
	n.onNewLeader()
}

func (n *Node) registerFailedHBChangesChan() {
	n.failedHBChangesChan = make(chan raft.Observation, 4)
	observer := raft.NewObserver(n.failedHBChangesChan, false, func(o *raft.Observation) bool {
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
			n.checkIfNodeNeedsUnblock()
		}
	}()
}

func (n *Node) onNewLeader() {
	go func() {
		for o := range n.leaderChangesChan {
			obs := o.Data.(raft.LeaderObservation)
			leaderID := string(obs.LeaderID)
			if leaderID != "" {
				n.consensusLogger.Info("New Leader: " + leaderID)
			} else {
				n.consensusLogger.Info("No Leader available in the Cluster")
				n.checkIfNodeNeedsUnblock()
			}
		}
	}()
}

func (n *Node) removeNodesOnHBStrategy() {
	go func() {
		const timeout = 20
		for o := range n.failedHBChangesChan {
			obs := o.Data.(raft.FailedHeartbeatObservation)
			dur := time.Since(obs.LastContact)
			duration := dur.Seconds()

			if duration >= timeout {
				warnMsg := fmt.Sprintf(
					"REMOVING NODE '%v' from the Leader due to not having a connection for %v secs...",
					obs.PeerID, duration,
				)
				n.consensusLogger.Warn(warnMsg)
				n.Consensus.RemoveServer(obs.PeerID, 0, 0)
				n.consensusLogger.Warn("NODE SUCCESSFULLY REMOVED FROM STATE CONSENSUS")
			}

		}
	}()
}

func (n *Node) checkIfNodeNeedsUnblock() {
	const timeout = 20 * time.Second

	if n.IsUnBlockingInProgress() {
		n.consensusLogger.Warn("UNBLOCKING ALREADY IN PROGRESS")
	}

	_, leaderID := n.Consensus.LeaderWithID()
	if leaderID != "" {
		return
	}

	time.Sleep(timeout)

	_, leaderID = n.Consensus.LeaderWithID()
	if leaderID != "" {
		return
	}

	n.unblockNode()
}

func (n *Node) unblockNode() {
	const errPanic = "COULDN'T GRACEFULLY UNBLOCK NODE. "

	log.Println("node got stuck for too long... Node reinstall in progress...")

	n.SetUnBlockingInProgress(true)
	defer n.SetUnBlockingInProgress(false)

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
