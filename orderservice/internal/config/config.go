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

	Events struct {
		Retries   int `env:"MAX_EVENT_RETRY" env-default:"3"`
		ProcLimit int `env:"MAX_PROCESSING_EVENTS" env-default:"4"`
	}

	Stockmarket struct {
		Endpoint string `env:"STOCKMARKET_GRPC_ENDPOINT" env-default:""`
	}

	Spot struct {
		Endpoint string `env:"SPOT_GRPC_ENDPOINT" env-required:"true"`
	}

	User struct {
		Endpoint string `env:"USER_GRPC_ENDPOINT" env-required:"true"`
	}

	Metrics struct {
		Port string `env:"METRICS_PORT" env-required:"true"`
	}

	Telemetry struct {
		JaegerGrpcAddress string `env:"JUEGER_GRPC_ADDRESS" env-required:"true"`
	}

	Kafka struct {
		Endpoint           string `env:"KAFKA_ENDPOINT" env-required:"true"`
		MarketsUpdateTopic string `env:"KAFKA_MARKETS_UPDATES_TOPIC" env-required:"true"`
		OrderUpdatesTopic  string `env:"KAFKA_ORDER_UPDATES_TOPIC" env-required:"true"`
		OrderCreatedTopic  string `env:"KAFKA_ORDER_CREATED_TOPIC" env-required:"true"`
		DLQTopic           string `env:"KAFKA_DLQ_TOPIC" env-required:"true"`
		GroupID            string `env:"KAFKA_GROUP" env-required:"true"`
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
		LogLevel  string `env:"LOG_LEVEL" env-default:"info"`
		LogToFile bool   `env:"ENABLE_LOGFILE" env-default:"false"`
		Dir       string `env:"LOG_FILE_DIR" env-default:"./logs"`
		LogPath   string `env:"-"`
	}

	Debug bool `env:"DEBUG" env-default:"false"`
}

func (c *Config) afterLoad() {
	c.Redis.Address = c.Redis.Host + ":" + c.Redis.Port

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
