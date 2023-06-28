package loadgen

import (
	"context"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"loadgen/internal/common"
)

func runGeneratorNominal(ctx context.Context, cfg Config, f namesFetcher) (loadTestResult, error) {
	res := loadTestResult{}

	const nameBatchLen = 1000
	tasks := make(chan []common.Name, cfg.WorkersCount)
	for i := 0; i < cfg.WorkersCount; i++ {
		names := make([]common.Name, nameBatchLen)
		f.GetNames(names)
		tasks <- names
	}

	wg := &sync.WaitGroup{}
	wg.Add(cfg.WorkersCount)

	workerCtx, cancelWorkerCtx := context.WithTimeout(ctx, cfg.Duration)
	defer cancelWorkerCtx()

	start := time.Now()
	for i := 0; i < cfg.WorkersCount; i++ {
		go func() {
			defer wg.Done()

			names := <-tasks
			for {
				select {
				case <-workerCtx.Done():
					return
				default:
				}
				email := generateRandomEmail(names)
				code, err := sendRequest(email)
				if err != nil {
					log.Printf("failed to send request: %v", err)
					continue
				}
				if code != http.StatusOK && code != http.StatusNotFound {
					log.Printf("an unexpected status code received: %d", code)
					continue
				}
				_ = atomic.AddInt64(&res.Ops, 1)
			}
		}()
	}

	wg.Wait()
	res.Duration = time.Since(start)

	return res, nil
}
