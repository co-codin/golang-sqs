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
	ApiServerPort        string `env:"APISERVER_PORT"`
	ApiServerHost        string `env:"APISERVER_HOST"`
	DatabaseName         string `env:"DB_NAME"`
	DatabaseHost         string `env:"DB_HOST"`
	DatabasePort         string `env:"DB_PORT"`
	DatabaseUser         string `env:"DB_USER"`
	DatabasePassword     string `env:"DB_PASSWORD"`
	DatabasePortTest     string `env:"DB_PORT_TEST"`
	Env                  Env    `env:"ENV" envDefault:"dev"`
	JwtSecret            string `env:"JWT_SECRET"`
	ProjectRoot          string `env:"PROJECT_ROOT"`
	AwsAccessKeyID string `env:"AWS_ACCESS_KEY_ID"`
	AwsAccessSecretKey string `env:"AWS_SECRET_ACCESS_KEY"`
	S3LocalstackEndpoint string `env:"S3_LOCALSTACK_ENDPOINT"`
	LocalstackEndpoint   string `env:"LOCALSTACK_ENDPOINT"`
	S3Bucket             string `env:"S3_BUCKET"`
	SqsQueue             string `env:"SQS_QUEUE"`
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
