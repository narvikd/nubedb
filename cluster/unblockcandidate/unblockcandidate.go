package unblockcandidate

import (
	"github.com/hashicorp/raft"
	"github.com/narvikd/errorskit"
	"github.com/narvikd/filekit"
	"log"
	"nubedb/cluster"
	"nubedb/discover"
	"nubedb/internal/app"
	"nubedb/internal/config"
	"time"
)

func Do(a *app.App) {
	go func() {
		for {
			const timeout = 20 * time.Second
			// If the node is stuck in a candidate position for 20 seconds, it's blocked.
			// There's no scenario where the node is a candidate for more than 10 seconds,
			// and it's not part of a bootstrapped service.
			if a.Node.Consensus.State() == raft.Candidate {
				time.Sleep(timeout)
				// If a minute has passed, and I'm still blocked, there's a real problem.
				if a.Node.Consensus.State() == raft.Candidate {
					unblockCandidate(a)
				}
			}
			time.Sleep(10 * time.Second)
		}
	}()
}

func unblockCandidate(a *app.App) {
	const errPanic = "COULDN'T GRACEFULLY UNBLOCK CANDIDATE. "

	log.Println("node got stuck in candidate for too long... Node reinstall in progress...")

	leader, errSearchLeader := discover.SearchLeader(a.Node.ID)
	if errSearchLeader != nil {
		errorskit.FatalWrap(errSearchLeader, errPanic+"couldn't search for leader")
	}
	leaderGrpcAddress := config.MakeGrpcAddress(leader)

	errConsensusRemove := cluster.ConsensusRemove(a.Config.CurrentNode.ID, leaderGrpcAddress)
	if errConsensusRemove != nil {
		errorskit.FatalWrap(errConsensusRemove, errPanic+"couldn't remove from consensus")
	}

	future := a.Node.Consensus.Shutdown()
	if future.Error() != nil {
		errorskit.FatalWrap(future.Error(), errPanic+"couldn't shut down")
	}

	errDeleteDirs := filekit.DeleteDirs(a.Node.Dir)
	if errDeleteDirs != nil {
		errorskit.FatalWrap(errDeleteDirs, errPanic+"couldn't delete dirs")
	}

	log.Fatalln("Node successfully reset. Restarting...")
}
