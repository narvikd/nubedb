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
	n.registerRequestVoteRequestChan()
}

func (n *Node) registerNodeChangesChan() {
	n.chans.nodeChanges = make(chan raft.Observation, 4)
	observer := raft.NewObserver(n.chans.nodeChanges, false, func(o *raft.Observation) bool {
		_, ok := o.Data.(raft.RaftState)
		return ok
	})
	n.Consensus.RegisterObserver(observer)

	// Call methods
	n.onNodeChange()
}

func (n *Node) registerLeaderChangesChan() {
	n.chans.leaderChanges = make(chan raft.Observation, 4)
	observer := raft.NewObserver(n.chans.leaderChanges, false, func(o *raft.Observation) bool {
		_, ok := o.Data.(raft.LeaderObservation)
		return ok
	})
	n.Consensus.RegisterObserver(observer)

	// Call methods
	n.onNewLeader()
}

func (n *Node) registerFailedHBChangesChan() {
	n.chans.failedHBChanges = make(chan raft.Observation, 4)
	observer := raft.NewObserver(n.chans.failedHBChanges, false, func(o *raft.Observation) bool {
		_, ok := o.Data.(raft.FailedHeartbeatObservation)
		return ok
	})
	n.Consensus.RegisterObserver(observer)

	// Call methods
	n.removeNodesOnHBStrategy()
}

func (n *Node) registerRequestVoteRequestChan() {
	n.chans.requestVoteRequest = make(chan raft.Observation, 4)
	observer := raft.NewObserver(n.chans.requestVoteRequest, false, func(o *raft.Observation) bool {
		_, ok := o.Data.(raft.RequestVoteRequest)
		return ok
	})
	n.Consensus.RegisterObserver(observer)

	// Call methods
	n.onRequestVoteRequest()
}

func (n *Node) onRequestVoteRequest() {
	go func() {
		for o := range n.chans.requestVoteRequest {
			data := o.Data.(raft.RequestVoteRequest)
			idRequester := string(data.ID)
			srvs := n.Consensus.GetConfiguration().Configuration().Servers
			if len(srvs) > 0 && !inNodeInConfiguration(srvs, idRequester) {
				msgWarn := fmt.Sprintf(
					"Foreign node (%s) attempted to request a vote, but node is not in the configuration. Requesting foreign node's reinstall",
					idRequester,
				)
				n.logger.Warn(msgWarn)
				err := cluster.RequestNodeReinstall(config.MakeGrpcAddress(idRequester))
				if err != nil {
					msgErr := errorskit.Wrap(err, "couldn't reinstall foreign node")
					n.logger.Error(msgErr.Error())
				}
			}
		}
	}()
}

func (n *Node) onNodeChange() {
	go func() {
		for o := range n.chans.nodeChanges {
			n.logger.Info("Node Changed to role: " + o.Data.(raft.RaftState).String())
		}
	}()
}

func (n *Node) onNewLeader() {
	go func() {
		for o := range n.chans.leaderChanges {
			obs := o.Data.(raft.LeaderObservation)
			leaderID := string(obs.LeaderID)
			if leaderID != "" {
				n.logger.Info("New Leader: " + leaderID)
			} else {
				n.logger.Info("No Leader available in the Cluster")
			}
		}
	}()
}

func (n *Node) removeNodesOnHBStrategy() {
	go func() {
		const timeout = 20
		for o := range n.chans.failedHBChanges {
			obs := o.Data.(raft.FailedHeartbeatObservation)
			dur := time.Since(obs.LastContact)
			duration := dur.Seconds()

			if duration >= timeout {
				warnMsg := fmt.Sprintf(
					"REMOVING NODE '%v' from the Leader due to not having a connection for %v secs...",
					obs.PeerID, duration,
				)
				n.logger.Warn(warnMsg)
				n.Consensus.RemoveServer(obs.PeerID, 0, 0)
				n.logger.Warn("NODE SUCCESSFULLY REMOVED FROM STATE CONSENSUS")
			}

		}
	}()
}

func (n *Node) ReinstallNode() {
	go func() {
		const errPanic = "COULDN'T GRACEFULLY REINSTALL NODE. "
		if n.IsUnBlockingInProgress() {
			n.logger.Warn("UNBLOCKING ALREADY IN PROGRESS")
			return
		}

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

		errDeleteDirs := filekit.DeleteDirs(n.MainDir)
		if errDeleteDirs != nil {
			errorskit.FatalWrap(errDeleteDirs, errPanic+"couldn't delete dirs")
		}

		log.Fatalln("Node successfully reset. Restarting...")
	}()
}

// TODO: Refactor
func inNodeInConfiguration(servers []raft.Server, id string) bool {
	for _, server := range servers {
		if server.ID == raft.ServerID(id) {
			return true
		}
	}
	return false
}
