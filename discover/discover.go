package discover

import (
	"fmt"
	"github.com/narvikd/errorskit"
	"github.com/narvikd/mdns"
	"log"
	"os"
	"sync"
	"time"
)

const serviceName = "_nubedb._tcp"

func ServeAndBlock(port int) error {
	host, errHostname := os.Hostname()
	if errHostname != nil {
		return errorskit.Wrap(errHostname, "couldn't find hostname")
	}

	info := []string{"nubedb Discover"}
	service, errService := mdns.NewMDNSService(host, serviceName, "", "", port, nil, info)
	if errService != nil {
		return errorskit.Wrap(errService, "discover service")
	}

	// Create the mDNS server, defer shutdown
	server, errServer := mdns.NewServer(&mdns.Config{Zone: service})
	if errServer != nil {
		return errorskit.Wrap(errService, "discover server")
	}

	defer server.Shutdown()
	select {} // Block forever
	return nil
}

func Search() ([]string, error) {
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
			fmt.Printf("Got new entry: %v\n", entry.Host)
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
