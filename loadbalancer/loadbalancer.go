package loadbalancer

import (
	"fmt"
	"github.com/hashicorp/raft"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"nubedb/internal/app"
	"nubedb/internal/config"
	"sync"
	"time"
)

type LB struct {
	sync.RWMutex
	nodes       []*url.URL
	currentNode int
}

func (l *LB) updateNodes(n []raft.Server) {
	var list []*url.URL
	for _, v := range n {
		srv := config.MakeApiAddr(string(v.ID))
		u := &url.URL{
			Scheme: "http",
			Host:   srv,
		}
		list = append(list, u)
	}
	l.Lock()
	defer l.Unlock()
	l.nodes = list
}

func (l *LB) getNextNode() *url.URL {
	l.Lock()
	defer l.Unlock()
	server := l.nodes[l.currentNode]
	l.currentNode = (l.currentNode + 1) % len(l.nodes)
	return server
}

func Start(a *app.App, port int) {
	timeout := 10 * time.Second
	time.Sleep(timeout)

	log.Println("Load balancer Started")
	srvs := a.Node.Consensus.GetConfiguration().Configuration().Servers
	lb := &LB{}
	lb.updateNodes(srvs)

	updateNodesPeriodically(a, lb)

	httpServ := &http.Server{
		Addr:              fmt.Sprintf(":%v", port),
		Handler:           lb,
		ReadHeaderTimeout: timeout, // GoSec: G114
	}
	err := httpServ.ListenAndServe()
	if err != nil {
		log.Fatalln(err)
	}
}

func (l *LB) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	server := l.getNextNode()
	proxy := httputil.NewSingleHostReverseProxy(server)
	proxy.ServeHTTP(w, r)
}

func updateNodesPeriodically(a *app.App, lb *LB) {
	go func() {
		for {
			srvs := a.Node.Consensus.GetConfiguration().Configuration().Servers
			lb.updateNodes(srvs)
			time.Sleep(1 * time.Minute)
		}
	}()
}
