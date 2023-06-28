package datagen

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"loadgen/internal/common"
)

type Generator struct {
	db *pgxpool.Pool
}

const maxConns = 10

func NewGenerator(ctx context.Context, cfg *Config) (*Generator, error) {
	if cfg == nil {
		return nil, errors.New("passed configuration is nil")
	}
	gen := &Generator{}
	connCfg, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the DSN: %w", err)
	}
	connCfg.MaxConns = maxConns
	gen.db, err = pgxpool.NewWithConfig(ctx, connCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize a pgx pool: %w", err)
	}
	return gen, nil
}

func (g *Generator) GenerateData(ctx context.Context, employeesCount int) error {
	positions, err := g.generatePositions(ctx)
	if err != nil {
		return fmt.Errorf("failed to generate employees positions: %w", err)
	}
	if err := g.generateEmployees(ctx, employeesCount, positions); err != nil {
		return fmt.Errorf("failed to generate employees: %w", err)
	}
	return nil
}

func (g *Generator) generatePositions(ctx context.Context) ([]int, error) {
	positions := []string{
		"Accountant", "Developer", "QA", "Designer", "PM",
	}

	tx, err := g.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin a transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	b := &pgx.Batch{}

	for _, p := range positions {
		_ = b.Queue(`
		INSERT INTO positions(title) VALUES($1)
		ON CONFLICT(title) DO NOTHING`, p)
	}

	res := tx.SendBatch(ctx, b)
	defer func() {
		_ = res.Close()
	}()

	for range positions {
		_, err := res.Exec()
		if err != nil {
			return nil, fmt.Errorf("failed to insert positions into DB: %w", err)
		}
	}
	_ = res.Close()

	r, err := tx.Query(ctx, `SELECT id FROM positions WHERE title = ANY($1)`, positions)
	if err != nil {
		return nil, fmt.Errorf("failed to get the positions IDs from the DB: %w", err)
	}

	defer r.Close()
	positionsIDs := make([]int, 0, len(positions))
	for r.Next() {
		var p int
		if err := r.Scan(&p); err != nil {
			return nil, fmt.Errorf("failed to scan the received ID: %w", err)
		}
		positionsIDs = append(positionsIDs, p)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit the positions generation result: %w", err)
	}
	return positionsIDs, nil
}

func (g *Generator) generateEmployees(ctx context.Context, employeesCount int, positions []int) error {
	workersCount := maxConns
	tasks := make(chan int, workersCount)
	results := make(chan workerResult, workersCount)
	f, err := common.NewNamesFetcher()
	if err != nil {
		return fmt.Errorf("failed to get a new names fetcher: %w", err)
	}
	workersCtx, cancelWorkersCtx := context.WithCancel(context.Background())
	defer cancelWorkersCtx()
	wg := &sync.WaitGroup{}
	wg.Add(workersCount)
	for i := 0; i < workersCount; i++ {
		go g.worker(workersCtx, wg, f, employeesCount, positions, tasks, results)
	}

	const batchSize = 500
	activeWorkers := 0
dispatcherLoop:
	for {
		select {
		case <-ctx.Done():
			cancelWorkersCtx()
			break dispatcherLoop
		default:
		}

		if activeWorkers == 0 && employeesCount == 0 {
			break
		}

		t := batchSize
		if t > employeesCount {
			t = employeesCount
		}
		if t != 0 {
			select {
			case tasks <- t:
				employeesCount -= t
				activeWorkers++
				continue
			default:
			}
		}

		if activeWorkers != 0 {
			res := <-results
			if res.Error != nil {
				log.Printf("worker has encountered an error: %v", res.Error)
				employeesCount += res.Task
			}
			activeWorkers--
		}
	}
	close(tasks)

	wg.Wait()
	return nil
}

type workerResult struct {
	Task  int
	Error error
}

func (g *Generator) worker(
	ctx context.Context,
	wg *sync.WaitGroup,
	f *common.NamesFetcher,
	employeesCount int,
	positions []int,
	tasks <-chan int,
	results chan<- workerResult,
) {
	defer wg.Done()
	for {
		t, ok := <-tasks
		if !ok {
			return
		}
		res := workerResult{
			Task: t,
		}
		if err := g.workerPayload(context.Background(), f, t, positions); err != nil {
			res.Error = err
		}
		select {
		case results <- res:
		case <-ctx.Done():
			return
		}
	}
}

func (g *Generator) workerPayload(
	ctx context.Context,
	f *common.NamesFetcher,
	employeesCount int,
	positions []int,
) error {
	names := make([]common.Name, employeesCount)
	f.GetNames(names)
	emps := make([]Employee, len(names))
	for i := range emps {
		salary := rand.Intn(200000)
		emps[i] = newEmployee(names[i], salary+1, positions[i%len(positions)])
	}
	if err := g.storeEmployees(ctx, emps); err != nil {
		return fmt.Errorf("failed to store the generated employees: %w", err)
	}
	return nil
}

func newEmployee(name common.Name, salary int, position int) Employee {
	return Employee{
		Name:     name,
		Salary:   salary,
		Position: position,
		Email:    fmt.Sprintf("%s.%s@gopher-corp.com", name.FirstName, name.LastName),
	}
}

func (g *Generator) storeEmployees(ctx context.Context, emps []Employee) error {
	b := &pgx.Batch{}
	for _, e := range emps {
		_ = b.Queue(
			`INSERT INTO employees (first_name, last_name, salary, position, email)
		VALUES ($1, $2, $3, $4, $5)`,
			e.FirstName, e.LastName, e.Salary, e.Position, e.Email,
		)
	}

	tx, err := g.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start a transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	res := tx.SendBatch(ctx, b)
	defer func() {
		_ = res.Close()
	}()

	for range emps {
		_, err := res.Exec()
		if err != nil {
			return fmt.Errorf("batch execution has failed: %w", err)
		}
	}
	_ = res.Close()

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit the batch insert transaction: %w", err)
	}
	return nil
}
