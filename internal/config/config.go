package config

import (
	"log"
	"os"
	"strconv"
)

// Config represents a config struct.
type Config struct {
	WorkersNum int //WORKERS_NUM
	QueueSize  int //QUEUE_SIZE
}

// MustLoad loads config from ENVIRONMENT VARIABLES.
// Panics if meets invalid values.
func MustLoad() Config {
	cfg := Config{
		WorkersNum: 4,
		QueueSize:  64,
	}

	if v := os.Getenv("WORKERS_NUM"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			log.Fatalf("invalid WORKERS_NUM: %v", err)
		}
		cfg.WorkersNum = n
	}

	if v := os.Getenv("QUEUE_SIZE"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			log.Fatalf("invalid QUEUE_SIZE: %v", err)
		}
		cfg.QueueSize = n
	}
	return cfg
}
