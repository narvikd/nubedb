package clusterobserver

import (
	"log"
	"nubedb/internal/app"
	"sync"
	"time"
)

func Launch(a *app.App) {
	var wg sync.WaitGroup
	log.Println("observer registered, sleeping...")
	log.Println("observer awake, launching...")

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
