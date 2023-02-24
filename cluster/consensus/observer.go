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

// registerObservers registers all observer channels for consensus.
func (n *Node) registerObservers() {
	n.registerNodeChangesChan()
	n.registerLeaderChangesChan()
	n.registerFailedHBChangesChan()
	n.registerRequestVoteRequestChan()
}

// registerNewObserver creates and registers a new observer with consensus which filters for T.
//
// Usage: registerNewObserver[type](consensus, chan)
//
// The [T any] defines a compile-time type argument named "T", it has to implement the "any" constraint
//
// (which is compatible with everything), inside the function it can be used as if was a type.
//
// When passing newObserver[raft.Something] it passes the type raft.Something as an argument.
func registerNewObserver[T any](consensus *raft.Raft, channel chan raft.Observation) {
	observer := raft.NewObserver(channel, false, func(o *raft.Observation) bool {
		_, ok := o.Data.(T)
		return ok
	})
	consensus.RegisterObserver(observer)
}

// registerNodeChangesChan registers the node changes observer channel.
func (n *Node) registerNodeChangesChan() {
	n.chans.nodeChanges = make(chan raft.Observation, 4)
	// Creates and register an observer that filters for raft state observations and sends them to the channel.
	registerNewObserver[raft.RaftState](n.Consensus, n.chans.nodeChanges)
	// Creates a goroutine to receive and handle the observations.
	go func() {
		// Blocks until something enters the channel
		for o := range n.chans.nodeChanges {
			n.logger.Info("Node Changed to role: " + o.Data.(raft.RaftState).String())
			n.checkIfNodeNeedsUnblock()
		}
	}()
}

// registerLeaderChangesChan registers the leader changes observer channel.
func (n *Node) registerLeaderChangesChan() {
	n.chans.leaderChanges = make(chan raft.Observation, 4)
	// Creates and registers an observer that filters for leader observations and sends them to the channel.
	registerNewObserver[raft.LeaderObservation](n.Consensus, n.chans.leaderChanges)
	// Creates a goroutine to receive and handle the observations.
	go func() {
		// Blocks until something enters the channel
		for o := range n.chans.leaderChanges {
			obs := o.Data.(raft.LeaderObservation)
			leaderID := string(obs.LeaderID)
			if leaderID != "" {
				n.logger.Info("New Leader: " + leaderID)
			} else {
				n.logger.Info("No Leader available in the Cluster")
				n.checkIfNodeNeedsUnblock()
			}
		}
	}()
}

// registerLeaderChangesChan registers the failed heartbeat observer channel.
func (n *Node) registerFailedHBChangesChan() {
	n.chans.failedHBChanges = make(chan raft.Observation, 4)
	// Creates and registers an observer that filters for failed heartbeat observations and sends them to the channel.
	registerNewObserver[raft.FailedHeartbeatObservation](n.Consensus, n.chans.failedHBChanges)
	// Call methods to handle incoming observations.
	go n.removeNodesOnHBStrategy()
}

// registerLeaderChangesChan registers the vote requests observer channel.
func (n *Node) registerRequestVoteRequestChan() {
	n.chans.requestVoteRequest = make(chan raft.Observation, 4)
	// Creates and registers an observer that filters for changes in election-votes observations and sends them to the channel.
	registerNewObserver[raft.RequestVoteRequest](n.Consensus, n.chans.requestVoteRequest)
	// Creates a goroutine to receive and handle the observations.
	go func() {
		// Blocks until something enters the channel
		for o := range n.chans.requestVoteRequest {
			data := o.Data.(raft.RequestVoteRequest)
			idRequester := string(data.ID)
			if !n.isNodeInConsensusServers(idRequester) {
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

func (n *Node) removeNodesOnHBStrategy() {
	// Blocks until something enters the channel
	for o := range n.chans.failedHBChanges {
		obs := o.Data.(raft.FailedHeartbeatObservation)
		warnMsg := fmt.Sprintf("REMOVING NODE '%v' from the Leader due to being offline...", obs.PeerID)
		n.logger.Warn(warnMsg)
		n.Consensus.RemoveServer(obs.PeerID, 0, 0)
		n.logger.Warn("NODE SUCCESSFULLY REMOVED FROM STATE CONSENSUS")
	}
}

// TODO: Maybe this will block because it isn't in a go routine?
func (n *Node) checkIfNodeNeedsUnblock() {
	const timeout = 1 * time.Minute
	_, leaderID := n.Consensus.LeaderWithID()
	if leaderID != "" {
		return
	}
	time.Sleep(timeout)
	_, leaderID = n.Consensus.LeaderWithID()
	if leaderID != "" {
		return
	}
	n.ReinstallNode()
}

// ReinstallNode gracefully reinstalls the node if it is stuck for too long
func (n *Node) ReinstallNode() {
	const errPanic = "COULDN'T GRACEFULLY REINSTALL NODE. "
	if n.IsUnBlockingInProgress() {
		n.logger.Warn("UNBLOCKING ALREADY IN PROGRESS")
		return
	}

	// Sets the flag that an unblocking process is in progress
	n.SetUnBlockingInProgress(true)
	defer n.SetUnBlockingInProgress(false)

	log.Println("node got stuck for too long... Node reinstall in progress...")

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
}

func (n *Node) isNodeInConsensusServers(id string) bool {
	srvs := n.Consensus.GetConfiguration().Configuration().Servers
	if len(srvs) <= 0 {
		return false
	}
	for _, server := range srvs {
		if server.ID == raft.ServerID(id) {
			return true
		}
	}
	return false
}
