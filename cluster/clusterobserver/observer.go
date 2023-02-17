package clusterobserver

import (
	"github.com/hashicorp/raft"
	"github.com/narvikd/errorskit"
	"github.com/narvikd/filekit"
	"log"
	"nubedb/cluster"
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

	leader := config.NodeCfg{}

	hotCfg := a.Node.Consensus.GetConfiguration().Configuration()
	consensusCfg := hotCfg.Clone()

	for _, srv := range consensusCfg.Servers {
		if string(srv.ID) == a.Config.CurrentNode.ID {
			continue
		}

		b, errComms := cluster.IsLeader(
			config.MakeGrpcAddress(string(srv.ID)),
		)
		if errComms != nil {
			errorskit.LogWrap(errComms, "couldn't contact to node while unblocking candidate")
			continue
		}

		if b {
			leader = config.NewNodeCfg(string(srv.ID))
			break
		}

		// Sleep between requests to not saturate the network too quickly
		time.Sleep(300 * time.Millisecond)
	}

	if leader.ID == "" {
		log.Fatalln(errPanic + "couldn't find any leader alive in the cluster. " +
			"Is the node disconnected from the network?. " +
			"ABORTING RESET: RESTARTING...")
	}

	errConsensusRemove := cluster.ConsensusRemove(a.Config.CurrentNode.ID, leader.GrpcAddress)
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

	errConsensusAdd := cluster.ConsensusJoin(a.Config.CurrentNode.ID, a.Config.CurrentNode.ConsensusAddress, leader.GrpcAddress)
	if errConsensusAdd != nil {
		errorskit.FatalWrap(errConsensusAdd, errPanic+"couldn't add node to consensus")
	}

	log.Fatalln("Node successfully reset. Restarting...")
}
