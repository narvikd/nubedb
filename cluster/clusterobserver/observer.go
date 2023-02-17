package clusterobserver

import (
	"github.com/hashicorp/raft"
	"github.com/narvikd/errorskit"
	"github.com/narvikd/filekit"
	"log"
	"nubedb/cluster"
	"nubedb/discover"
	"nubedb/internal/app"
	"nubedb/internal/config"
	"sync"
	"time"
)

func Launch(a *app.App) {
	var wg sync.WaitGroup
	log.Println("observer registered, sleeping...")
	log.Println("observer awake, launching...")

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			handleUnblockCandidate(a)
			time.Sleep(10 * time.Second)
		}
	}()

	wg.Wait()
}

func handleUnblockCandidate(a *app.App) {
	const timeout = 10 * time.Second

	hotCfg := a.Node.Consensus.GetConfiguration().Configuration()
	consensusCfg := hotCfg.Clone()

	// If the server count is superior to 2, it means that the candidate was part of a cluster configuration,
	// and the server isn't coincidentally being bootstrapped.
	isNodeConsensusBlocked := a.Node.Consensus.State() == raft.Candidate && len(consensusCfg.Servers) >= 2
	if isNodeConsensusBlocked {
		time.Sleep(timeout)
		// If a minute has passed, and I'm still blocked, there's a real problem.
		if isNodeConsensusBlocked {
			unblockCandidate(a)
		}
	}
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

	errConsensusAdd := cluster.ConsensusJoin(a.Config.CurrentNode.ID, a.Config.CurrentNode.ConsensusAddress, leaderGrpcAddress)
	if errConsensusAdd != nil {
		errorskit.FatalWrap(errConsensusAdd, errPanic+"couldn't add node to consensus")
	}

	log.Fatalln("Node successfully reset. Restarting...")
}
