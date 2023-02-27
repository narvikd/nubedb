package consensus

import (
	"errors"
	"fmt"
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
	"nubedb/pkg/filterwriter"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Node struct defines the properties of a node
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

// Chans struct defines the channels used for the observers
type Chans struct {
	nodeChanges        chan raft.Observation
	leaderChanges      chan raft.Observation
	failedHBChanges    chan raft.Observation
	requestVoteRequest chan raft.Observation
}

// IsUnBlockingInProgress returns a boolean value indicating whether the node is already trying to unblock itself
func (n *Node) IsUnBlockingInProgress() bool {
	n.RLock()
	defer n.RUnlock()
	return n.unBlockingInProgress
}

// SetUnBlockingInProgress sets the unBlockingInProgress property of the Node struct
// which indicates whether the node is trying to unblock itself
func (n *Node) SetUnBlockingInProgress(b bool) {
	n.Lock()
	defer n.Unlock()
	n.unBlockingInProgress = b
}

// New initializes and returns a new Node
func New(cfg config.Config) (*Node, error) {
	n, errNode := newNode(cfg)
	if errNode != nil {
		return nil, errNode
	}

	errRaft := n.setRaft()
	if errRaft != nil {
		return nil, errRaft
	}

	return n, nil
}

// newNode initializes and returns a new Node with the given id and address
func newNode(cfg config.Config) (*Node, error) {
	dir := path.Join("data", cfg.CurrentNode.ID)
	storageDir := path.Join(dir, "localdb")

	f, errDB := fsm.New(storageDir, cfg.CurrentNode.FSMPerformanceMode)
	if errDB != nil {
		return nil, errDB
	}

	n := &Node{
		FSM:              f,
		ID:               cfg.CurrentNode.ID,
		ConsensusAddress: cfg.CurrentNode.ConsensusAddress,
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

// setRaft initializes and starts a new consensus instance using the node's configuration.
func (n *Node) setRaft() error {
	const (
		timeout            = 10 * time.Second
		maxConnectionsPool = 10
		retainedSnapshots  = 3

		// SnapshotThreshold controls how many outstanding logs there must be before
		// we perform a snapshot. This is to prevent excessive snapshotting by
		// replaying a small set of logs instead. The value passed here is the initial
		// setting used. This can be tuned during operation using ReloadConfig.
		// Note: This defaults to a very high value and can cause spam to disk if it's set low.
		// TODO: Find appropriate value
		snapshotThreshold = 2
	)

	// Resolve the TCP address for use in Raft's consensus.
	tcpAddr, errAddr := net.ResolveTCPAddr("tcp", n.ConsensusAddress)
	if errAddr != nil {
		return errorskit.Wrap(errAddr, "couldn't resolve addr")
	}

	// Create the transport
	transport, errTransport := raft.NewTCPTransport(n.ConsensusAddress, tcpAddr, maxConnectionsPool, timeout, os.Stderr)
	if errTransport != nil {
		return errorskit.Wrap(errTransport, "couldn't create transport")
	}

	// Create the log DB
	dbStore, errRaftStore := raftboltdb.NewBoltStore(n.consensusDBPath)
	if errRaftStore != nil {
		return errorskit.Wrap(errRaftStore, "couldn't create consensus db")
	}

	// Create the snapshot store
	snaps, errSnapStore := raft.NewFileSnapshotStore(n.snapshotsDir, retainedSnapshots, os.Stderr)
	if errSnapStore != nil {
		return errorskit.Wrap(errSnapStore, "couldn't create consensus snapshot storage")
	}

	// Sett the rest of the configuration for the consensus.
	nodeID := raft.ServerID(n.ID)
	cfg := raft.DefaultConfig()
	cfg.LocalID = nodeID
	cfg.SnapshotInterval = timeout
	cfg.SnapshotThreshold = snapshotThreshold
	n.setConsensusLogger(cfg)

	// Create a new Raft instance and set it.
	r, errRaft := raft.NewRaft(cfg, n.FSM, dbStore, dbStore, snaps, transport)
	if errRaft != nil {
		return errorskit.Wrap(errRaft, "couldn't create new consensus")
	}
	n.Consensus = r

	// Start the consensus process for this node.
	errStartConsensus := n.startConsensus(string(nodeID))
	if errStartConsensus != nil {
		return errStartConsensus
	}

	errClusterReadiness := n.waitForClusterReadiness()
	if errClusterReadiness != nil {
		return errClusterReadiness
	}
	// Register the observers
	n.registerObservers()
	return nil
}

func (n *Node) setConsensusLogger(cfg *raft.Config) {
	// Setups a filter-writer which suppresses raft's errors that should have been debug errors instead
	const errCannotSnapshotNow = "snapshot now, wait until the configuration entry at"
	filters := []string{raft.ErrNothingNewToSnapshot.Error(), errCannotSnapshotNow}
	fw := filterwriter.New(os.Stderr, filters)

	l := hclog.New(&hclog.LoggerOptions{
		Name:   "consensus",
		Level:  hclog.LevelFromString("DEBUG"),
		Output: fw,
	})
	n.logger = l
	cfg.LogOutput = fw
	cfg.Logger = l
}

// startConsensus boots up the consensus process for the node, by adding it to an existing or new cluster.
func (n *Node) startConsensus(currentNodeID string) error {
	// Define the bootstrapping leader ID, this is useful in case the consensus hasn't been started yet.
	const bootstrappingLeader = "bootstrap-node"

	// Check if the consensus has already been bootstrapped.
	consensusCfg := n.Consensus.GetConfiguration().Configuration()
	if len(consensusCfg.Servers) >= 2 {
		n.logger.Info("consensus already bootstrapped")
		return nil
	}

	// Define the list of bootstrapping servers.
	bootstrappingServers := newConsensusServerList(bootstrappingLeader)

	// Search for an existing leader and use it to overwrite the bootstrapping list.
	// This is used in case bootstrappingLeader is down, or if it isn't the leader.
	leaderID, errSearchLeader := discover.SearchLeader(currentNodeID)
	if errSearchLeader == nil {
		bootstrappingServers = newConsensusServerList(leaderID)
	}

	// The consensus is going to try to bootstrap.
	// If it gives an error it's because it's already bootstrapped, or this node is not a part of it: "not a voter".
	// Any errors that aren't "not a voter", can safely be ignored as described in raft's docs.
	// not a voter shouldn't be ignored just because we need to join the existing consensus.
	future := n.Consensus.BootstrapCluster(raft.Configuration{Servers: bootstrappingServers})
	if future.Error() != nil {
		if !strings.Contains(future.Error().Error(), "not a voter") {
			return nil
		}
	}
	// Consensus not bootstrapped but there isn't any voter error (this means this node is 'bootstrappingLeader')
	if future.Error() == nil {
		return nil
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

func newConsensusServerList(nodeID string) []raft.Server {
	return []raft.Server{
		{
			ID:      raft.ServerID(nodeID),
			Address: raft.ServerAddress(config.MakeConsensusAddr(nodeID)),
		},
	}
}

func (n *Node) waitForClusterReadiness() error {
	currentTry := 0
	const (
		maxRetryCount = 7
		sleepTime     = 1 * time.Minute
	)
	for {
		currentTry++
		if currentTry > maxRetryCount {
			return errors.New("quorum retry max reached")
		}

		if n.IsQuorumPossible(true) {
			n.logger.Info("quorum possible.")
			break
		}

		_, errLeader := discover.SearchLeader(n.ID)
		if errLeader != nil {
			msg := fmt.Sprintf("it isn't possible to reach Quorum due to lack of nodes. "+
				"Tried to search for a leader to join an existent consensus. "+
				"Leader not found: %v",
				errLeader.Error(),
			)
			n.logger.Error(msg)
			n.logger.Error("it isn't possible to reach Quorum due to lack of nodes and leader not available. Retrying...")
			time.Sleep(sleepTime)
			continue
		}

		n.logger.Warn("existent leader found, reinstalling....")
		n.ReinstallNode()
	}
	return nil
}
