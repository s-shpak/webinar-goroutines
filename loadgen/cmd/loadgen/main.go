package main

import (
	"context"
	"fmt"
	"loadgen/internal/loadgen"
	"log"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()
	cfg, err := GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	if err := loadgen.Generate(ctx, cfg.Generator); err != nil {
		return fmt.Errorf("load generation has failed: %w", err)
	}
	return nil
}
