// Package mdnsdiscover is responsible for handling the discovery of nubedb nodes.
package mdnsdiscover

import (
	"errors"
	"github.com/hashicorp/raft"
	"github.com/narvikd/errorskit"
	"github.com/narvikd/mdns"
	"log"
	"net"
	"nubedb/api/proto"
	"nubedb/api/proto/protoclient"
	"nubedb/internal/config"
	"nubedb/pkg/ipkit"
	"nubedb/pkg/resolver"
	"strings"
	"sync"
	"time"
)

const (
	// The service name identifier used for the discovery.
	serviceName       = "_nubedb._tcp"
	ErrLeaderNotFound = "couldn't find a leader"
)

// ServeAndBlock creates a new discovery service with the given node ID and port, blocks indefinitely.
func ServeAndBlock(nodeID string, port int) {
	const errGen = "Discover serve and block: "
	info := []string{"nubedb Discover"}

	ip, errGetIP := getIP(nodeID)
	if errGetIP != nil {
		errorskit.FatalWrap(errGetIP, errGen)
	}

	// Create a new mDNS service for the node.
	service, errService := mdns.NewMDNSService(nodeID, serviceName, "", "", port, []net.IP{ip}, info)
	if errService != nil {
		errorskit.FatalWrap(errService, errGen+"discover service")
	}

	// Create a new mDNS server for the service.
	server, errServer := mdns.NewServer(&mdns.Config{Zone: service})
	if errServer != nil {
		errorskit.FatalWrap(errService, errGen+"discover server")
	}

	// Shut down the server when the function returns. (Which shouldn't)
	defer func(server *mdns.Server) {
		_ = server.Shutdown()
	}(server)

	// Block indefinitely.
	select {}
}

func getIP(nodeID string) (net.IP, error) {
	hosts, errLookup := net.LookupHost(nodeID)
	if errLookup != nil {
		return nil, errorskit.Wrap(errLookup, "couldn't lookup host")
	}

	return net.ParseIP(hosts[0]), nil
}

// SearchLeader will return an error if a leader is not found,
// since it skips the current node and this could be a leader.
//
// Because it skips the current node, it will still return an error.
//
// This is done this way to ensure this function is never called to do gRPC operations in itself.
func SearchLeader(currentNode string) (string, error) {
	nodes, errNodes := searchNodes(currentNode)
	if errNodes != nil {
		return "", errNodes
	}

	for _, node := range nodes {
		grpcAddr := ipkit.NewAddr(node, config.GrpcPort)
		leader, err := isLeader(grpcAddr)
		if err != nil {
			errorskit.LogWrap(err, "couldn't contact node while searching for leaders")
			continue
		}
		if leader {
			return node, nil
		}
	}

	return "", errors.New(ErrLeaderNotFound)
}

// searchNodes returns a list of all discovered nodes, excluding the one passed as a parameter.
func searchNodes(currentNode string) ([]string, error) {
	// map to store the discovered nodes.
	hosts := make(map[string]bool)
	var lastError error

	// Try to discover nodes 3 times to add any missing nodes in the first scan.
	for i := 0; i < 3; i++ {
		hostsQuery, err := query()
		if err != nil {
			log.Println(err)
			lastError = err
			continue
		}

		for _, host := range hostsQuery {
			// In some linux versions it reports "$name." (name and a dot)
			host = strings.ReplaceAll(host, ".", "")
			hosts[host] = true
		}

		// Wait for 100 milliseconds before trying again to not spam/have some space between requests.
		// TODO: Try to refactor this
		time.Sleep(100 * time.Millisecond)
	}

	// Convert the map to a slice of strings and exclude the current node.
	result := make([]string, 0, len(hosts))
	for host := range hosts {
		if currentNode == host {
			continue
		}
		result = append(result, host)
	}

	return result, lastError
}

// query sends an mDNS query to discover nubedb nodes and returns a list of their hosts.
func query() ([]string, error) {
	var mu sync.Mutex
	var hosts []string
	entriesCh := make(chan *mdns.ServiceEntry, 4)
	go func() {
		for entry := range entriesCh {
			mu.Lock()
			hosts = append(hosts, entry.Host)
			mu.Unlock()
		}
	}()

	params := mdns.DefaultParams(serviceName)
	params.DisableIPv6 = true
	params.Entries = entriesCh

	defer close(entriesCh)
	err := mdns.Query(params)
	if err != nil {
		return nil, errorskit.Wrap(err, "discover search")
	}

	mu.Lock()
	defer mu.Unlock()
	return hosts, nil
}

// isLeader takes a GRPC address and returns if the node reports back as a Leader
func isLeader(addr string) (bool, error) {
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
		if resolver.IsHostAlive(srvID, timeout) {
			alive = append(alive, srv)
		}
	}
	return alive
}
