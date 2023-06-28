package loadgen

import "time"

type Config struct {
	Duration     time.Duration
	WorkersCount int
}
