package cluster

import (
	"github.com/hashicorp/raft"
	"github.com/narvikd/errorskit"
	"github.com/narvikd/filekit"
	"log"
	"nubedb/internal/app"
	"nubedb/internal/config"
	"time"
)

func LaunchObserver(a *app.App) {
	log.Println("observer launched, sleeping...")
	time.Sleep(20 * time.Second)
	log.Println("observer awake, launching...")
	go func() {
		for {
			handleUnblockCandidate(a)
			time.Sleep(10 * time.Second)
		}
	}()
}

func handleUnblockCandidate(a *app.App) {
	const timeout = 1 * time.Minute
	if a.Node.Consensus.State() == raft.Candidate {
		time.Sleep(timeout)
		// If a minute has passed, and I'm still a candidate, there's a problem
		if a.Node.Consensus.State() == raft.Candidate {
			unblockCandidate(a)
		}
	}
}

func unblockCandidate(a *app.App) {
	const errPanic = "COULDN'T GRACEFULLY UNBLOCK CANDIDATE"

	log.Println("node got stuck in candidate for too long... Node reinstall in progress...")

	leader := config.NodeCfg{}

	for id, nodeCfg := range a.Config.Nodes {
		if id == a.Config.CurrentNode.ID {
			continue
		}

		b, errComms := isLeader(nodeCfg.GrpcAddress)
		if errComms != nil {
			errorskit.LogWrap(errComms, "couldn't contact to node while unblocking candidate")
			continue
		}

		if b {
			leader = nodeCfg
			break
		}
	}

	if leader.ID == "" {
		log.Fatalln(errPanic)
	}

	errRemove := consensusRemove(a.Config.CurrentNode.ID, leader.GrpcAddress)
	if errRemove != nil {
		errorskit.FatalWrap(errRemove, errPanic)
	}

	future := a.Node.Consensus.Shutdown()
	if future.Error() != nil {
		errorskit.FatalWrap(future.Error(), errPanic)
	}

	errDeleteDirs := filekit.DeleteDirs(a.Node.Dir)
	if errDeleteDirs != nil {
		errorskit.FatalWrap(errDeleteDirs, errPanic)
	}

	errAdd := consensusJoin(a.Config.CurrentNode.ID, a.Config.CurrentNode.ConsensusAddress, leader.GrpcAddress)
	if errAdd != nil {
		errorskit.FatalWrap(errAdd, errPanic)
	}

	log.Fatalln("Node successfully reset. Restarting...")
}
