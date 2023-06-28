package main

import (
	"flag"
	"fmt"

	"loadgen/internal/datagen"
)

type Config struct {
	Generator      datagen.Config
	EmployeesCount int
}

func GetConfig() (Config, error) {
	c := Config{}

	flag.StringVar(&c.Generator.DSN, "d", "", "dsn")
	flag.IntVar(&c.EmployeesCount, "n", 1000, "employees count")
	flag.Parse()

	if c.EmployeesCount <= 0 {
		return c, fmt.Errorf("employees count should be greater than 0, got %d", c.EmployeesCount)
	}

	return c, nil
}
