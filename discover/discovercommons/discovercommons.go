package discovercommons

import (
	"fmt"
	"github.com/hashicorp/raft"
	"github.com/narvikd/errorskit"
	"net"
	"nubedb/api/proto"
	"nubedb/api/proto/protoclient"
	"nubedb/pkg/resolver"
	"time"
)

// IsLeader takes a GRPC address and returns if the node reports back as a Leader
func IsLeader(addr string) (bool, error) {
	conn, errConn := protoclient.NewConnection(addr)
	if errConn != nil {
		return false, errConn
	}
	defer conn.Cleanup()

	res, errTalk := conn.Client.IsLeader(conn.Ctx, &proto.Empty{})
	if errTalk != nil {
		return false, errorskit.Wrap(errTalk, "failed to get an ok response from the Node via grpc")
	}

	return res.IsLeader, nil
}

// SearchAliveNodes will skip currentNodeID.
func SearchAliveNodes(consensus *raft.Raft, currentNodeID string) []raft.Server {
	const timeout = 300 * time.Millisecond
	var alive []raft.Server

	liveCfg := consensus.GetConfiguration().Configuration()
	cfg := liveCfg.Clone() // Clone CFG to not keep calling it in the for, in case the num of servers is very large
	for _, srv := range cfg.Servers {
		srvID := string(srv.ID)
		if currentNodeID == srvID {
			continue
		}

		if !resolver.IsHostAlive(srvID, timeout) {
			continue
		}

		conn, err := net.DialTimeout("tcp", string(srv.Address), timeout)
		if err != nil {
			continue
		}
		_ = conn.Close()
		alive = append(alive, srv)
		fmt.Println("ADDR ALIVE:", alive)
	}
	return alive
}
