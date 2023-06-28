package main

import (
	"flag"
	"fmt"
	"time"

	"loadgen/internal/loadgen"
)

type Config struct {
	Generator loadgen.Config
}

func GetConfig() (Config, error) {
	c := Config{}

	flag.DurationVar(&c.Generator.Duration, "dur", time.Second*5, "load testing duration")
	flag.IntVar(&c.Generator.WorkersCount, "workers", 10, "number of workers")
	flag.Parse()

	if c.Generator.WorkersCount <= 0 {
		return c, fmt.Errorf("workers count should be greater than 0, got %d", c.Generator.WorkersCount)
	}

	return c, nil
}
