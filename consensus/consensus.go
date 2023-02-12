package consensus

import (
	"github.com/hashicorp/raft"
	"github.com/hashicorp/raft-boltdb"
	"github.com/narvikd/errorskit"
	"github.com/narvikd/filekit"
	"net"
	"nubedb/consensus/fsm"
	"os"
	"path"
	"sync"
	"time"
)

type Node struct {
	db           *sync.Map
	ID           string `json:"id" validate:"required"`
	Address      string `json:"address" validate:"required"`
	dir          string
	storageDir   string
	snapshotsDir string
}

func New(db *sync.Map, id string, address string) (*raft.Raft, error) {
	n, errNode := newNode(db, id, address)
	if errNode != nil {
		return nil, errNode
	}

	r, errRaft := n.setRaft()
	if errRaft != nil {
		return nil, errRaft
	}

	return r, nil
}

func newNode(db *sync.Map, id string, address string) (*Node, error) {
	dir := path.Join("data", id)

	n := &Node{
		db:           db,
		ID:           id,
		Address:      address,
		dir:          dir,
		storageDir:   path.Join(dir, "db_store"),
		snapshotsDir: dir, // This isn't a typo, it will create a snapshots dir inside the dir automatically
	}

	errDir := filekit.CreateDirs(n.dir, false)
	if errDir != nil {
		return nil, errDir
	}

	return n, nil
}

func (n *Node) setRaft() (*raft.Raft, error) {
	const timeout = 10 * time.Second
	serverID := raft.ServerID(n.ID)

	tcpAddr, errAddr := net.ResolveTCPAddr("tcp", n.Address)
	if errAddr != nil {
		return nil, errorskit.Wrap(errAddr, "couldn't resolve addr")
	}

	transport, errTransport := raft.NewTCPTransport(n.Address, tcpAddr, 10, timeout, os.Stderr)
	if errTransport != nil {
		return nil, errorskit.Wrap(errTransport, "couldn't create transport")
	}

	dbStore, errRaftStore := raftboltdb.NewBoltStore(n.storageDir)
	if errRaftStore != nil {
		return nil, errorskit.Wrap(errRaftStore, "couldn't create raft db storage")
	}

	snaps, errSnapStore := raft.NewFileSnapshotStore(n.snapshotsDir, 2, os.Stderr)
	if errSnapStore != nil {
		return nil, errorskit.Wrap(errSnapStore, "couldn't create raft snapshot storage")
	}

	cfg := raft.DefaultConfig()
	cfg.LocalID = serverID
	cfg.SnapshotInterval = timeout
	cfg.SnapshotThreshold = 2

	r, errRaft := raft.NewRaft(cfg, fsm.New(n.db), dbStore, dbStore, snaps, transport)
	if errRaft != nil {
		return nil, errorskit.Wrap(errRaft, "couldn't create new raft")
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

	return r, nil
}
