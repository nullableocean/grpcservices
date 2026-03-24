package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	App struct {
		Port    string `env:"SERVER_PORT" env-default:"8085"`
		Address string `env:"SERVER_ADDRESS" env-default:""`
		Name    string `env:"APP_NAME" env-default:"spot-instrument"`
	}

	Auth struct {
		JWT_SECRET string `env:"JWT_SECRET" env-required:"true"`
	}

	Postgres struct {
		Username string `env:"PG_USER" env-required:"true"`
		Password string `env:"PG_PASS" env-required:"true"`
		Host     string `env:"PG_HOST" env-required:"true"`
		Port     string `env:"PG_PORT" env-required:"true"`
		DBName   string `env:"PG_DBNAME" env-required:"true"`
		DSN      string `env:"-"`
	}

	Metrics struct {
		Port string `env:"METRICS_PORT" env-required:"true"`
	}

	Telemetry struct {
		JaegerGrpcAddress string `env:"JAEGER_GRPC_ADDRESS" env-required:"true"`
	}

	Log struct {
		LogLevel  string `env:"LOG_LEVEL" env-default:"info"`
		LogToFile bool   `env:"ENABLE_LOGFILE" env-default:"false"`
		Dir       string `env:"LOG_FILE_DIR" env-default:"./logs"`
		LogPath   string `env:"-"`
	}

	Debug bool `env:"DEBUG" env-default:"false"`
}

func (c *Config) afterLoad() {
	// dsn: postgres://username:password@localhost:5432/dbname
	c.Postgres.DSN = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.Postgres.Username,
		c.Postgres.Password,
		c.Postgres.Host,
		c.Postgres.Port,
		c.Postgres.DBName,
	)

	if c.Log.LogToFile {
		c.Log.LogPath = c.Log.Dir + "/logs.log"
	}
}

func NewConfig() (*Config, error) {
	var (
		envPath = flag.String("env", ".env", "path to .env file")
		debug   = flag.Bool("debug", false, "debug mode")
	)
	flag.Parse()

	_ = godotenv.Load(*envPath)

	cfg := &Config{}

	if err := cleanenv.ReadEnv(cfg); err != nil {
		return nil, fmt.Errorf("config load error: %w", err)
	}

	cfg.Debug = *debug || cfg.Debug

	cfg.afterLoad()

	if cfg.Log.LogToFile {
		if err := checkExistOrCreateLogDir(cfg.Log.Dir); err != nil {
			return nil, fmt.Errorf("log dir error: %w", err)
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
