package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	App struct {
		Name    string `env:"APP_NAME"    env-default:"user-service"`
		Port    string `env:"SERVER_PORT" env-required:"true"`
		Address string `env:"SERVER_ADDRESS" env-default:""`
	}

	Metrics struct {
		Port string `env:"METRICS_PORT" env-required:"true"`
	}

	Telemetry struct {
		JaegerGrpcAddress string `env:"JUEGER_GRPC_ADDRESS" env-required:"true"`
	}

	Log struct {
		LogLevel  string `env:"LOG_LEVEL" env-default:"info"`
		LogToFile bool   `env:"ENABLE_LOGFILE" env-default:"false"`
		Dir       string `env:"LOG_FILE_DIR" env-default:"./logs"`
		LogPath   string `env:"-"`
	}

	Seed  bool `env:"SEED" env-default:"false"`
	Debug bool `env:"DEBUG" env-default:"false"`
}

func (c *Config) afterLoad() {
	if c.Log.LogToFile {
		c.Log.LogPath = filepath.Join(c.Log.Dir, "logs.log")
	}
}

func NewConfig() (*Config, error) {
	var (
		envPath = flag.String("env", ".env", "path to .env file")
		seed    = flag.Bool("seed", false, "seed data")
		debug   = flag.Bool("debug", false, "debug mode")
	)
	flag.Parse()

	_ = godotenv.Load(*envPath)

	cfg := &Config{}

	if err := cleanenv.ReadEnv(cfg); err != nil {
		return nil, fmt.Errorf("failed to read config from env: %w", err)
	}

	if *seed {
		cfg.Seed = true
	}
	if *debug {
		cfg.Debug = true
	}

	cfg.afterLoad()

	if cfg.Log.LogToFile {
		if err := checkExistOrCreateLogDir(cfg.Log.Dir); err != nil {
			return nil, fmt.Errorf("log directory error: %w", err)
		}
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
