package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"loadgen/internal/datagen"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	c, err := GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	ctx, cancelCtx := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancelCtx()

	g, err := datagen.NewGenerator(ctx, &c.Generator)
	if err != nil {
		return fmt.Errorf("failed to initialize a new generator: %w", err)
	}
	if err := g.GenerateData(ctx, c.EmployeesCount); err != nil {
		return fmt.Errorf("failed to generate employees data: %w", err)
	}
	return nil
}
