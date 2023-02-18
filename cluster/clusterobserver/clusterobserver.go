package clusterobserver

import (
	"nubedb/internal/app"
	"sync"
	"time"
)

func Launch(a *app.App) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			handleUnblockCandidate(a)
			time.Sleep(10 * time.Second)
		}
	}()
	wg.Wait()
}
