package discover

import (
	"fmt"
	"github.com/hashicorp/mdns"
	"github.com/narvikd/errorskit"
	"log"
	"os"
	"sync"
	"time"
)

const serviceName = "_nubedb._tcp"

func Start() {
	var wg sync.WaitGroup
	time.Sleep(10 * time.Second)

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := serveAndBlock(8000)
		if err != nil {
			return
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			err := search()
			if err != nil {
				log.Println(err)
			}
			time.Sleep(3 * time.Second)
		}
	}()
	wg.Wait()
}

func search() error {
	// Make a channel for results and start listening
	entriesCh := make(chan *mdns.ServiceEntry, 16)
	go func() {
		for entry := range entriesCh {
			fmt.Printf("Got new entry: %v\n", entry.Host)
		}
	}()

	// Start the lookup
	params := mdns.DefaultParams(serviceName)
	params.DisableIPv6 = true
	params.Entries = entriesCh

	defer close(entriesCh)
	err := mdns.Query(params)
	if err != nil {
		return errorskit.Wrap(err, "discover search")
	}
	return nil
}

func serveAndBlock(port int) error {
	// Setup our service export
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
