package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	App struct {
		Port    string `env:"SERVER_PORT" env-default:"8085"`
		Address string `env:"SERVER_ADDRESS" env-default:""`
		Name    string `env:"APP_NAME" env-default:"order-service"`
	}

	Spot struct {
		Endpoint string `env:"SPOT_GRPC_ENDPOINT" env-required:"true"`
	}

	Metrics struct {
		Port string `env:"METRICS_PORT" env-required:"true"`
	}

	Telemetry struct {
		JaegerGrpcAddress string `env:"JUEGER_GRPC_ADDRESS" env-required:"true"`
	}

	Redis struct {
		Host     string        `env:"REDIS_HOST" env-default:"localhost"`
		Port     string        `env:"REDIS_PORT" env-default:"6379"`
		Password string        `env:"REDIS_PASSWORD"`
		Username string        `env:"REDIS_USERNAME"`
		DB       int           `env:"REDIS_DB" env-default:"0"`
		TTL      time.Duration `env:"REDIS_TTL" env-default:"300s"`
		Address  string        `env:"-"`
	}

	Log struct {
		Dir     string `env:"LOGS_DIR" env-default:"./logs"`
		LogPath string `env:"-"`
	}

	Seed  bool `env:"SEED" env-default:"false"`
	Debug bool `env:"DEBUG" env-default:"false"`
}

func (c *Config) afterLoad() {
	c.Redis.Address = c.Redis.Host + ":" + c.Redis.Port
	c.Log.LogPath = c.Log.Dir + "/logs.log"
}

func NewConfig() (*Config, error) {
	var (
		envPath = flag.String("env", ".env", "path to .env file")
		logDir  = flag.String("logdir", "", "dir for logs (overrides env and default)")
		seed    = flag.Bool("seed", false, "seed data")
		debug   = flag.Bool("debug", false, "debug mode")
	)
	flag.Parse()

	_ = godotenv.Load(*envPath)

	cfg := &Config{}

	if err := cleanenv.ReadEnv(cfg); err != nil {
		return nil, fmt.Errorf("config load error: %w", err)
	}

	if *logDir != "" {
		cfg.Log.Dir = *logDir
	}
	cfg.Seed = *seed || cfg.Seed
	cfg.Debug = *debug || cfg.Debug

	cfg.afterLoad()

	if err := checkExistOrCreateLogDir(cfg.Log.Dir); err != nil {
		return nil, fmt.Errorf("log dir error: %w", err)
	}

	return cfg, nil
}

func checkExistOrCreateLogDir(dir string) error {
	stat, err := os.Stat(dir)
	if err != nil && os.IsNotExist(err) {
		if err := os.Mkdir(dir, 0755); err != nil {
			return err
		}
	} else if !stat.IsDir() {
		return fmt.Errorf("logger config error: wait dir in path")
	}

	return nil
}
