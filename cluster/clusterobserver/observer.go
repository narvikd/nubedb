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
	log.Println("observer launched, sleeping...")
	time.Sleep(10 * time.Second)
	log.Println("observer awake, launching...")

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			handleUnblockCandidate(a)
			time.Sleep(10 * time.Second)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		handleBootstrapping(a)
	}()

	wg.Wait()
}

func handleUnblockCandidate(a *app.App) {
	const timeout = 20 * time.Second
	if a.Node.Consensus.State() == raft.Candidate {
		time.Sleep(timeout)
		// If a minute has passed, and I'm still a candidate, there's a problem
		if a.Node.Consensus.State() == raft.Candidate {
			unblockCandidate(a)
		}
	}
}

func unblockCandidate(a *app.App) {
	const errPanic = "COULDN'T GRACEFULLY UNBLOCK CANDIDATE. "

	log.Println("node got stuck in candidate for too long... Node reinstall in progress...")

	leader := config.NodeCfg{}

	for id, nodeCfg := range a.Config.Nodes {
		if id == a.Config.CurrentNode.ID {
			continue
		}

		b, errComms := cluster.IsLeader(nodeCfg.GrpcAddress)
		if errComms != nil {
			errorskit.LogWrap(errComms, "couldn't contact to node while unblocking candidate")
			continue
		}

		if b {
			leader = nodeCfg
			break
		}

		// Sleep between requests to not saturate the network too quickly
		time.Sleep(300 * time.Millisecond)
	}

	if leader.ID == "" {
		log.Fatalln(errPanic + "leader id is empty")
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

func handleBootstrapping(a *app.App) {
	var leaders []config.NodeCfg

	if a.Node.Consensus.State() != raft.Leader {
		return
	}

	if a.Config.CurrentNode.ID != "node1" {
		return
	}

	log.Println("Checking if cluster needs to be bootstrapped...")

	for _, nodeCfg := range a.Config.Nodes {
		b, errComms := cluster.IsLeader(nodeCfg.GrpcAddress)
		if errComms != nil {
			errorskit.FatalWrap(errComms, "couldn't contact to node while bootstrapping")
		}

		if b {
			leaders = append(leaders, nodeCfg)
		}

		// Sleep between requests to not saturate the network too quickly
		time.Sleep(300 * time.Millisecond)
	}

	if len(leaders) != 3 {
		log.Println("No need to bootstrap the cluster...")
		return
	}

	log.Println("Bootstrapping cluster...")

	for id, nodeCfg := range a.Config.Nodes {
		if id == a.Config.CurrentNode.ID {
			continue
		}

		future := a.Node.Consensus.AddVoter(
			raft.ServerID(nodeCfg.ID), raft.ServerAddress(nodeCfg.ConsensusAddress), 0, 0,
		)
		if future.Error() != nil {
			errorskit.FatalWrap(future.Error(), "failed to add server while bootstrapping")
		}

		// Sleep between requests to not saturate the network too quickly
		time.Sleep(300 * time.Millisecond)
	}

	log.Println("Bootstrapping successful!")
}
