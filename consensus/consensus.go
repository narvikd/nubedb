package consensus

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/hashicorp/raft"
	"github.com/hashicorp/raft-boltdb"
	"github.com/narvikd/errorskit"
	"github.com/narvikd/filekit"
	"net"
	"nubedb/consensus/fsm"
	"os"
	"path"
	"time"
)

type Node struct {
	Consensus    *raft.Raft
	FSM          *fsm.DatabaseFSM
	ID           string `json:"id" validate:"required"`
	Address      string `json:"address" validate:"required"`
	dir          string
	storageDir   string
	snapshotsDir string
}

func New(id string, address string) (*Node, error) {
	n, errNode := newNode(id, address)
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
	storageDir := path.Join(dir, "db_store")

	f, errDB := newFSM(storageDir)
	if errDB != nil {
		return nil, errDB
	}

	n := &Node{
		FSM:          f,
		ID:           id,
		Address:      address,
		dir:          dir,
		storageDir:   storageDir,
		snapshotsDir: dir, // This isn't a typo, it will create a snapshots dir inside the dir automatically
	}

	errDir := filekit.CreateDirs(n.dir, false)
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
	const timeout = 10 * time.Second
	serverID := raft.ServerID(n.ID)

	tcpAddr, errAddr := net.ResolveTCPAddr("tcp", n.Address)
	if errAddr != nil {
		return errorskit.Wrap(errAddr, "couldn't resolve addr")
	}

	transport, errTransport := raft.NewTCPTransport(n.Address, tcpAddr, 10, timeout, os.Stderr)
	if errTransport != nil {
		return errorskit.Wrap(errTransport, "couldn't create transport")
	}

	dbStore, errRaftStore := raftboltdb.NewBoltStore(n.storageDir)
	if errRaftStore != nil {
		return errorskit.Wrap(errRaftStore, "couldn't create raft db storage")
	}

	snaps, errSnapStore := raft.NewFileSnapshotStore(n.snapshotsDir, 2, os.Stderr)
	if errSnapStore != nil {
		return errorskit.Wrap(errSnapStore, "couldn't create raft snapshot storage")
	}

	cfg := raft.DefaultConfig()
	cfg.LocalID = serverID
	cfg.SnapshotInterval = timeout
	cfg.SnapshotThreshold = 2

	r, errRaft := raft.NewRaft(cfg, n.FSM, dbStore, dbStore, snaps, transport)
	if errRaft != nil {
		return errorskit.Wrap(errRaft, "couldn't create new raft")
	}

	raftCfg := raft.Configuration{
		Servers: []raft.Server{
			{
				ID:      serverID,
				Address: transport.LocalAddr(),
			},
		},
	}

	r.BootstrapCluster(raftCfg)

	n.Consensus = r

	return nil
}
