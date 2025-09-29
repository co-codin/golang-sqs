package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Env string

const (
	Env_Test Env = "test"
	Env_Dev  Env = "dev"
)

type Config struct {
	DatabaseName     string `env:"DB_NAME"`
	DatabaseHost     string `env:"DB_HOST"`
	DatabasePort     string `env:"DB_PORT"`
	DatabaseUser     string `env:"DB_USER"`
	DatabasePassword string `env:"DB_PASSWORD"`
	DatabasePortTest string `env:"DB_PORT_TEST"`
	Env              Env    `env:"ENV" envDefault:"dev"`
}

func (c *Config) DatabaseUrl() string {
	port := c.DatabasePort
	if c.Env == Env_Test {
		port = c.DatabasePortTest
	}
	return fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		c.DatabaseUser,
		c.DatabasePassword,
		c.DatabaseHost,
		port,
		c.DatabaseName,
	)
}

func New() (*Config, error) {
	cfg, err := env.ParseAs[Config]()

	if err != nil {
		return &cfg, fmt.Errorf("failed to load config: %w", err)
	}

	return &cfg, nil
}
