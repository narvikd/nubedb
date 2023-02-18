package discover

import (
	"errors"
	"fmt"
	"github.com/narvikd/errorskit"
	"github.com/narvikd/mdns"
	"log"
	"net"
	"nubedb/cluster"
	"nubedb/internal/config"
	"strings"
	"sync"
	"time"
)

const serviceName = "_nubedb._tcp"

func ServeAndBlock(nodeID string, port int) {
	const errGen = "Discover serve and block: "
	info := []string{"nubedb Discover"}

	ip, errGetIP := getIP(nodeID)
	if errGetIP != nil {
		errorskit.FatalWrap(errGetIP, errGen)
	}

	service, errService := mdns.NewMDNSService(nodeID, serviceName, "", "", port, []net.IP{ip}, info)
	if errService != nil {
		errorskit.FatalWrap(errService, errGen+"discover service")
	}

	// Create the mDNS server, defer shutdown
	server, errServer := mdns.NewServer(&mdns.Config{Zone: service})
	if errServer != nil {
		errorskit.FatalWrap(errService, errGen+"discover server")
	}

	// TODO: This maybe never shutdowns correctly
	defer func(server *mdns.Server) {
		_ = server.Shutdown()
	}(server)
	select {} // Block forever
}

func getIP(nodeID string) (net.IP, error) {
	hosts, errLookup := net.LookupHost(nodeID)
	if errLookup != nil {
		return nil, errorskit.Wrap(errLookup, "couldn't lookup host")
	}

	return net.ParseIP(hosts[0]), nil
}

// SearchNodes returns a list where currentNode is skipped
func SearchNodes(currentNode string) (map[string]net.IP, error) {
	hosts := make(map[string]net.IP)
	var lastError error

	for i := 0; i < 3; i++ {
		hostsQuery, err := query()
		if err != nil {
			log.Println(err)
			lastError = err
			continue
		}

		for h, ip := range hostsQuery {
			// In some linux versions it reports "$name." (name and a dot)
			hostSlice := strings.Split(h, ".")
			host := hostSlice[0]
			if host != currentNode {
				hosts[host] = ip
			}
		}
		time.Sleep(100 * time.Millisecond) // TODO: Try to refactor this
	}

	return hosts, lastError
}

func query() (map[string]net.IP, error) {
	hosts := make(map[string]net.IP)
	entriesCh := make(chan *mdns.ServiceEntry, 16)
	go func() {
		var mu sync.Mutex
		for entry := range entriesCh {
			mu.Lock()
			hosts[entry.Host] = entry.AddrV4
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
	return hosts, nil
}

// SearchLeader will return an error if a leader is not found, since it skips the current node.
//
// If the current node is as leader, it will still return an error
func SearchLeader(currentNode string) (string, error) {
	nodes, errNodes := SearchNodes(currentNode)
	if errNodes != nil {
		return "", errNodes
	}

	fmt.Println("SEARCH LEADER NODES", nodes)

	for host, _ := range nodes {
		leader, err := cluster.IsLeader(config.MakeGrpcAddress(host))
		if err != nil {
			errorskit.LogWrap(err, "couldn't contact node while searching for leaders")
			continue
		}
		if leader {
			return host, nil
		}
	}

	return "", errors.New("couldn't find a leader")
}
