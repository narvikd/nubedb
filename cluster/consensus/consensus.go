package consensus

import (
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
	"sync"
	"time"
)

type Node struct {
	sync.RWMutex
	Consensus            *raft.Raft
	FSM                  *fsm.DatabaseFSM
	ID                   string `json:"id" validate:"required"`
	ConsensusAddress     string `json:"address"`
	MainDir              string
	storageDir           string
	snapshotsDir         string
	consensusDBPath      string
	logger               hclog.Logger
	chans                *Chans
	unBlockingInProgress bool
}

type Chans struct {
	nodeChanges        chan raft.Observation
	leaderChanges      chan raft.Observation
	failedHBChanges    chan raft.Observation
	requestVoteRequest chan raft.Observation
}

func (n *Node) IsUnBlockingInProgress() bool {
	n.RLock()
	defer n.RUnlock()
	return n.unBlockingInProgress
}

func (n *Node) SetUnBlockingInProgress(b bool) {
	n.Lock()
	defer n.Unlock()
	n.unBlockingInProgress = b
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
		FSM:              f,
		ID:               id,
		ConsensusAddress: address,
		MainDir:          dir,
		storageDir:       storageDir,
		snapshotsDir:     dir, // This isn't a typo, it will create a snapshots dir inside the dir automatically
		consensusDBPath:  filepath.Join(dir, "consensus.db"),
		chans:            new(Chans),
	}

	errDir := filekit.CreateDirs(n.MainDir, false)
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

	tcpAddr, errAddr := net.ResolveTCPAddr("tcp", n.ConsensusAddress)
	if errAddr != nil {
		return errorskit.Wrap(errAddr, "couldn't resolve addr")
	}

	transport, errTransport := raft.NewTCPTransport(n.ConsensusAddress, tcpAddr, maxConnectionsPool, timeout, os.Stderr)
	if errTransport != nil {
		return errorskit.Wrap(errTransport, "couldn't create transport")
	}

	dbStore, errRaftStore := raftboltdb.NewBoltStore(n.consensusDBPath)
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

	errStartConsensus := n.startConsensus(string(nodeID))
	if errStartConsensus != nil {
		return errStartConsensus
	}

	n.registerObservers()
	return nil
}

func (n *Node) startConsensus(currentNodeID string) error {
	const bootstrappingLeader = "bootstrap-node"
	consensusCfg := n.Consensus.GetConfiguration().Configuration()
	if len(consensusCfg.Servers) >= 2 {
		return nil // Consensus already bootstrapped
	}

	bootstrappingServers := []raft.Server{
		{
			ID:      raft.ServerID(bootstrappingLeader),
			Address: raft.ServerAddress(config.MakeConsensusAddr(bootstrappingLeader)),
		},
	}

	leaderID, errSearchLeader := discover.SearchLeader(currentNodeID)
	if errSearchLeader == nil {
		bootstrappingServers = []raft.Server{
			{
				ID:      raft.ServerID(leaderID),
				Address: raft.ServerAddress(config.MakeConsensusAddr(leaderID)),
			},
		}
	}

	// The consensus is going to try to bootstrap. If it gives an error it's because it's already bootstrapped, or
	// We're not a part of it: "not a voter".
	// Any errors that aren't "not a voter", can safely be ignored as described in raft's docs.
	// not a voter shouldn't be ignored just because we need to join the existing consensus.
	future := n.Consensus.BootstrapCluster(raft.Configuration{Servers: bootstrappingServers})
	if future.Error() != nil {
		if !strings.Contains(future.Error().Error(), "not a voter") {
			return nil
		}
	}

	errJoin := joinNodeToExistingConsensus(currentNodeID)
	if errJoin != nil {
		errLower := strings.ToLower(errJoin.Error())
		if strings.Contains(errLower, "was already part of the network") {
			return nil
		}
		return errorskit.Wrap(errJoin, "while bootstrapping")
	}

	return nil
}

func joinNodeToExistingConsensus(nodeID string) error {
	leaderID, errSearchLeader := discover.SearchLeader(nodeID)
	if errSearchLeader != nil {
		return errSearchLeader
	}
	return cluster.ConsensusJoin(nodeID, config.MakeConsensusAddr(nodeID), config.MakeGrpcAddress(leaderID))
}
