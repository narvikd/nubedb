package consensus

import (
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/raft"
	"github.com/hashicorp/raft-boltdb/v2"
	"github.com/narvikd/errorskit"
	"github.com/narvikd/filekit"
	"net"
	"nubedb/cluster"
	"nubedb/cluster/consensus/fsm"
	"nubedb/discover"
	"nubedb/internal/config"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type Node struct {
	Consensus       *raft.Raft
	FSM             *fsm.DatabaseFSM
	ID              string `json:"id" validate:"required"`
	Address         string `json:"address"`
	Dir             string
	storageDir      string
	snapshotsDir    string
	consensusDB     string
	consensusLogger hclog.Logger
}

func New(cfg config.Config) (*Node, error) {
	n, errNode := newNode(cfg.CurrentNode.ID, cfg.CurrentNode.ConsensusAddress)
	if errNode != nil {
		return nil, errNode
	}

	errRaft := n.setRaft()
	if errRaft != nil {
		return nil, errRaft
	}

	return n, nil
}

func newNode(id string, address string) (*Node, error) {
	dir := path.Join("data", id)
	storageDir := path.Join(dir, "localdb")

	f, errDB := newFSM(storageDir)
	if errDB != nil {
		return nil, errDB
	}

	n := &Node{
		FSM:          f,
		ID:           id,
		Address:      address,
		Dir:          dir,
		storageDir:   storageDir,
		snapshotsDir: dir, // This isn't a typo, it will create a snapshots dir inside the dir automatically
		consensusDB:  filepath.Join(dir, "consensus.db"),
	}

	errDir := filekit.CreateDirs(n.Dir, false)
	if errDir != nil {
		return nil, errDir
	}

	return n, nil
}

func newFSM(dir string) (*fsm.DatabaseFSM, error) {
	db, err := badger.Open(badger.DefaultOptions(dir))
	if err != nil {
		return nil, errorskit.Wrap(err, "couldn't open badgerDB")
	}

	return fsm.New(db), nil
}

func (n *Node) setRaft() error {
	const (
		timeout            = 10 * time.Second
		maxConnectionsPool = 10
		retainedSnapshots  = 3
		// Note: This defaults to a very high value and can cause spam to disk if it's too low.
		// TODO: Find appropriate value
		snapshotThreshold = 2
	)

	tcpAddr, errAddr := net.ResolveTCPAddr("tcp", n.Address)
	if errAddr != nil {
		return errorskit.Wrap(errAddr, "couldn't resolve addr")
	}

	transport, errTransport := raft.NewTCPTransport(n.Address, tcpAddr, maxConnectionsPool, timeout, os.Stderr)
	if errTransport != nil {
		return errorskit.Wrap(errTransport, "couldn't create transport")
	}

	dbStore, errRaftStore := raftboltdb.NewBoltStore(n.consensusDB)
	if errRaftStore != nil {
		return errorskit.Wrap(errRaftStore, "couldn't create consensus db")
	}

	snaps, errSnapStore := raft.NewFileSnapshotStore(n.snapshotsDir, retainedSnapshots, os.Stderr)
	if errSnapStore != nil {
		return errorskit.Wrap(errSnapStore, "couldn't create consensus snapshot storage")
	}

	nodeID := raft.ServerID(n.ID)
	cfg := raft.DefaultConfig()
	cfg.LocalID = nodeID
	cfg.SnapshotInterval = timeout
	cfg.SnapshotThreshold = snapshotThreshold
	n.setConsensusLogger(cfg)

	r, errRaft := raft.NewRaft(cfg, n.FSM, dbStore, dbStore, snaps, transport)
	if errRaft != nil {
		return errorskit.Wrap(errRaft, "couldn't create new consensus")
	}
	n.Consensus = r

	return n.startConsensus(string(nodeID))
}

func (n *Node) startConsensus(currentNodeID string) error {
	var bootstrappingServers []raft.Server
	for i := 1; i <= 3; i++ {
		id := fmt.Sprintf("node%v", i)
		srv := raft.Server{
			ID:      raft.ServerID(id),
			Address: raft.ServerAddress(config.MakeConsensusAddr(id)),
		}
		bootstrappingServers = append(bootstrappingServers, srv)
	}

	future := n.Consensus.BootstrapCluster(raft.Configuration{Servers: bootstrappingServers})
	if future.Error() != nil {
		if !strings.Contains(future.Error().Error(), "not a voter") {
			return nil // Consensus already bootstrapped.
		}
	}
	if future.Error() == nil && isNodePresentInServers(currentNodeID, bootstrappingServers) {
		return nil // Consensus not bootstrapped, but node is a part of it, so it will bootstrap without any intervention.
	}

	// At this point, the consensus wasn't bootstrapped before.
	// A bootstrap config was created where this node isn't part of it.
	return joinNodeToExistingConsensus(currentNodeID)
}

func isNodePresentInServers(nodeID string, servers []raft.Server) bool {
	for _, server := range servers {
		if nodeID == string(server.ID) {
			return true
		}
	}
	return false
}

func joinNodeToExistingConsensus(nodeID string) error {
	leaderID, errSearchLeader := discover.SearchLeader(nodeID)
	if errSearchLeader != nil {
		return errSearchLeader
	}

	return cluster.ConsensusJoin(nodeID, config.MakeConsensusAddr(nodeID), config.MakeGrpcAddress(leaderID))
}
