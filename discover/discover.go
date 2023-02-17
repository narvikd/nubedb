package discover

import (
	"errors"
	"github.com/narvikd/errorskit"
	"github.com/narvikd/mdns"
	"log"
	"nubedb/cluster"
	"nubedb/internal/config"
	"os"
	"sync"
	"time"
)

const serviceName = "_nubedb._tcp"

func ServeAndBlock(port int) {
	host, errHostname := os.Hostname()
	if errHostname != nil {
		errorskit.FatalWrap(errHostname, "couldn't find hostname")
	}

	info := []string{"nubedb Discover"}
	service, errService := mdns.NewMDNSService(host, serviceName, "", "", port, nil, info)
	if errService != nil {
		errorskit.FatalWrap(errService, "discover service")
	}

	// Create the mDNS server, defer shutdown
	server, errServer := mdns.NewServer(&mdns.Config{Zone: service})
	if errServer != nil {
		errorskit.FatalWrap(errService, "discover server")
	}

	// TODO: This maybe never shutdowns correctly
	defer func(server *mdns.Server) {
		_ = server.Shutdown()
	}(server)
	select {} // Block forever
}

// SearchNodes returns a list where currentNode is skipped
func SearchNodes(currentNode string) ([]string, error) {
	hosts := make(map[string]bool)
	var lastError error

	for i := 0; i < 3; i++ {
		hostsQuery, err := query()
		if err != nil {
			log.Println(err)
			lastError = err
			continue
		}

		for _, host := range hostsQuery {
			hosts[host] = true
		}
		time.Sleep(100 * time.Millisecond) // TODO: Try to refactor this
	}

	result := make([]string, 0, len(hosts))
	for host := range hosts {
		if currentNode == host {
			continue
		}
		result = append(result, host)
	}

	return result, lastError
}

func query() ([]string, error) {
	var hosts []string
	entriesCh := make(chan *mdns.ServiceEntry, 16)
	go func() {
		var mu sync.Mutex
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

	for _, node := range nodes {
		leader, err := cluster.IsLeader(config.MakeGrpcAddress(node))
		if err != nil {
			errorskit.LogWrap(err, "couldn't contact node while searching for leaders")
			continue
		}
		if leader {
			return node, nil
		}
	}

	return "", errors.New("couldn't find a leader")
}
