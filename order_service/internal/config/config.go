package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"github.com/sony/gobreaker"
	"go.opentelemetry.io/otel/codes"
)

type Config struct {
	App            AppConfig
	Auth           AuthConfig
	Postgres       PostgresConfig
	Spot           SpotConfig
	Metrics        MetricsConfig
	Telemetry      TelemetryConfig
	Log            LogConfig
	GRPC           GRPCConfig
	CircuitBreaker CircuitBreakerConfig
	Retry          RetryConfig
	Kafka          KafkaConfig
	Outbox         OutboxConfig
	Env            EnvConfig
}

type AppConfig struct {
	Name            string        `env:"APP_NAME" env-default:"order-service"`
	Address         string        `env:"SERVER_ADDRESS" env-default:""`
	Port            string        `env:"SERVER_PORT" env-default:"8085"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" env-default:"15s"`
}

type EnvConfig struct {
	Debug bool `env:"DEBUG" env-default:"false"`
}

type AuthConfig struct {
	JWTSecret string `env:"JWT_SECRET" env-required:"true"`
}

type PostgresConfig struct {
	Username        string        `env:"PG_USER" env-required:"true"`
	Password        string        `env:"PG_PASS" env-required:"true"`
	Host            string        `env:"PG_HOST" env-required:"true"`
	Port            string        `env:"PG_PORT" env-required:"true"`
	DBName          string        `env:"PG_DBNAME" env-required:"true"`
	MaxConns        int32         `env:"PG_MAX_CONNS" env-default:"100"`
	MinConns        int32         `env:"PG_MIN_CONNS" env-default:"10"`
	MaxConnLifetime time.Duration `env:"PG_MAX_CONN_LIFETIME" env-default:"30m"`
	MaxConnIdleTime time.Duration `env:"PG_MAX_CONN_IDLE_TIME" env-default:"5m"`
	ConnTimeout     time.Duration `env:"PG_CONN_TIMEOUT" env-default:"5s"`
	DSN             string        `env:"-"`
}

type SpotConfig struct {
	Endpoint string `env:"SPOT_GRPC_ENDPOINT" env-required:"true"`
}

type MetricsConfig struct {
	Port string `env:"METRICS_PORT" env-required:"true"`
	Path string `env:"METRICS_PATH" env-default:"/metrics"`
}

type TelemetryConfig struct {
	ExporterGrpcAddress string  `env:"OPEN_TELEMETRY_EXPORTER_GRPC_ADDRESS" env-required:"true"`
	SampleRatio         float64 `env:"TELEMETRY_SAMPLE_RATIO" env-default:"0.1"`
}

type LogConfig struct {
	Level     string `env:"LOG_LEVEL" env-default:"info"`
	LogToFile bool   `env:"ENABLE_LOGFILE" env-default:"false"`
	Dir       string `env:"LOG_FILE_DIR" env-default:"./logs"`
	Path      string `env:"-"`
}

type GRPCConfig struct {
	ServerMaxRecvMsgSize       int           `env:"GRPC_SERVER_MAX_RECV_MSG_SIZE" env-default:"4194304"`
	ServerMaxSendMsgSize       int           `env:"GRPC_SERVER_MAX_SEND_MSG_SIZE" env-default:"4194304"`
	ServerMaxConcurrentStreams uint32        `env:"GRPC_SERVER_MAX_CONCURRENT_STREAMS" env-default:"1000"`
	ClientTimeout              time.Duration `env:"GRPC_CLIENT_TIMEOUT" env-default:"10s"`
	Keepalive                  KeepaliveConfig
}

type KeepaliveConfig struct {
	Time                time.Duration `env:"GRPC_KEEPALIVE_TIME" env-default:"30s"`
	Timeout             time.Duration `env:"GRPC_KEEPALIVE_TIMEOUT" env-default:"10s"`
	PermitWithoutStream bool          `env:"GRPC_KEEPALIVE_PERMIT_WITHOUT_STREAM" env-default:"true"`
}

type CircuitBreakerConfig struct {
	MaxRequests uint32                             `env:"BREAKER_MAX_REQUESTS" env-default:"3"`
	Interval    time.Duration                      `env:"BREAKER_INTERVAL" env-default:"10s"`
	Timeout     time.Duration                      `env:"BREAKER_TIMEOUT" env-default:"30s"`
	ReadyToTrip func(counts gobreaker.Counts) bool // не из env, задаётся в коде
}

type RetryConfig struct {
	MaxRetries uint          `env:"RETRY_MAX" env-default:"3"`
	Backoff    time.Duration `env:"RETRY_BACKOFF" env-default:"100ms"` // начальная задержка
	MaxBackoff time.Duration `env:"RETRY_MAX_BACKOFF" env-default:"2s"`
	RetryCodes []codes.Code  `env:"-"` // задаётся в коде
}

type KafkaConfig struct {
	Brokers             []string      `env:"KAFKA_BROKERS" env-required:"false"` // разделённые запятыми
	TopicUpdates        string        `env:"KAFKA_TOPIC_UPDATES" env-default:"order-updates"`
	TopicMarketsUpdates string        `env:"KAFKA_TOPIC_MARKETS_UPDATES" env-default:"markets-updates"`
	TopicCreated        string        `env:"KAFKA_TOPIC_CREATED" env-default:"order-created"`
	DLQTopic            string        `env:"KAFKA_DLQ_TOPIC" env-default:"order-dlq"`
	GroupID             string        `env:"KAFKA_GROUP_ID" env-default:"order-service"`
	MinBytes            int           `env:"KAFKA_MIN_BYTES" env-default:"10e3"`
	MaxBytes            int           `env:"KAFKA_MAX_BYTES" env-default:"10e6"`
	MaxWait             time.Duration `env:"KAFKA_MAX_WAIT" env-default:"1s"`
	DialTimeout         time.Duration `env:"KAFKA_DIAL_TIMEOUT" env-default:"10s"`
}

type OutboxConfig struct {
	PollInterval time.Duration `env:"OUTBOX_POLL_INTERVAL" env-default:"1s"`
	BatchSize    int           `env:"OUTBOX_BATCH_SIZE" env-default:"100"`
	MaxRetries   int           `env:"OUTBOX_MAX_RETRIES" env-default:"3"`
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
		c.Log.Path = c.Log.Dir + "/logs.log"
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

	cfg.Env.Debug = *debug || cfg.Env.Debug

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
