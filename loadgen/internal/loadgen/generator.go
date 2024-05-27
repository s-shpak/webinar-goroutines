package loadgen

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"loadgen/internal/common"
)

func Generate(ctx context.Context, cfg Config) error {
	res, err := generate(ctx, cfg)
	if err != nil {
		return err
	}
	opsPerSecond := res.Ops / int64(res.Duration.Seconds())
	log.Printf("ops per second: %d", opsPerSecond)
	return nil
}

type loadTestResult struct {
	Duration time.Duration
	Ops      int64
}

func generate(ctx context.Context, cfg Config) (loadTestResult, error) {
	f, err := common.NewNamesFetcher()
	if err != nil {
		return loadTestResult{}, fmt.Errorf("failed to get a new names fetcher: %w", err)
	}

	//return runGeneratorNominal(ctx, cfg, f)
	return runGenerator(ctx, cfg, f)
}

type namesFetcher interface {
	GetNames(dst []common.Name)
}

func runGenerator(ctx context.Context, cfg Config, f namesFetcher) (loadTestResult, error) {
	return loadTestResult{}, fmt.Errorf("NYI")
}

func generateRandomEmail(names []common.Name) string {
	idx := rand.Intn(len(names))
	n := names[idx]
	return fmt.Sprintf("%s.%s@gopher-corp.com", n.FirstName, n.LastName)
}

func sendRequest(email string) (int, error) {
	emailEscaped := url.PathEscape(email)
	u, err := url.JoinPath("http://localhost:8080/employee-by-email/", emailEscaped)
	if err != nil {
		return 0, fmt.Errorf("failed to join the URL path: %w", err)
	}
	resp, err := http.DefaultClient.Get(u)
	if err != nil {
		return 0, fmt.Errorf("http request has failed: %w", err)
	}
	defer resp.Body.Close()

	return resp.StatusCode, nil
}
