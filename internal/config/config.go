package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

const (
	StoragePostgres = "postgres"
	StorageMemory   = "memory"
)

type (
	Config struct {
		Server    Server
		Storage   Storage
		Database  Database
		Snowflake Snowflake
		BaseURL   string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	}

	Server struct {
		Port string `env:"PORT" envDefault:"8080"`
	}

	Storage struct {
		Type string `env:"STORAGE" envDefault:"postgres"`
	}

	Database struct {
		URL string `env:"DATABASE_URL"`
	}

	Snowflake struct {
		MachineID uint64 `env:"SNOWFLAKE_MACHINE_ID" envDefault:"0"`
	}
)

func New() (*Config, error) {
	_ = godotenv.Load(".env")

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("config.New: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config.New: %w", err)
	}

	return cfg, nil
}

func (c *Config) validate() error {
	switch c.Storage.Type {
	case StoragePostgres:
		if c.Database.URL == "" {
			return fmt.Errorf("DATABASE_URL is required when STORAGE=%s", StoragePostgres)
		}
	case StorageMemory:
	default:
		return fmt.Errorf("unknown STORAGE value %q (expected %q or %q)", c.Storage.Type, StoragePostgres, StorageMemory)
	}
	return nil
}
