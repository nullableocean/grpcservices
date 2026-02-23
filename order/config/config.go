package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

var (
	appName = "order-service"

	defaultEnvPath = ".env"

	defaultLogsDir = "./logs"
	logFilename    = "logs.log"
)

var (
	envPortKey    = "SERVER_PORT"
	envAddressKey = "SERVER_ADDRESS"

	envSpotEndpointKey = "SPOT_GRPC_ENDPOINT"

	envLogsDirKey     = "LOGS_DIR"
	envMetricsPortKey = "METRICS_PORT"

	envJaegerAddressKey = "JUEGER_GRPC_ADDRESS"
	envRedisHostKey     = "REDIS_HOST"
	envRedisPortKey     = "REDIS_PORT"
	envRedisPasswordKey = "REDIS_PASSWORD"
	envRedisUsernameKey = "REDIS_USERNAME"
	envRedisDBKey       = "REDIS_DB"
	envRedisTTLKey      = "REDIS_TTL_SECONDS"
)

var (
	envPathFlag = flag.String("env", defaultEnvPath, "path to env")
	logsDirFlag = flag.String("logdir", defaultLogsDir, "dir for logs")
	seedFlag    = flag.Bool("seed", false, "path to logs")
	debugFlag   = flag.Bool("debug", false, "debug")
)

type Telemetry struct {
	JaegerGrpcAddress string
}

type Metrics struct {
	Port string
}

type Spot struct {
	Endpoint string
}

type App struct {
	Port    string
	Address string
	Name    string
}

type Log struct {
	LogPath string
}

type Seed struct {
	Need bool
}

type Redis struct {
	Address  string
	Password string
	Username string
	DB       int
	TTL      time.Duration
}

type Config struct {
	App       *App
	Spot      *Spot
	Metrics   *Metrics
	Telemetry *Telemetry
	Redis     *Redis

	Log  *Log
	Seed *Seed

	Debug bool
}

func NewConfig() (*Config, error) {
	flag.Parse()

	envPath := *envPathFlag
	godotenv.Load(envPath)

	config := &Config{}

	err := loadAppCnf(config)
	if err != nil {
		return nil, err
	}

	// spot api server
	err = loadSpotApiCnf(config)
	if err != nil {
		return nil, err
	}

	//logger
	err = loadLoggerCnf(config)
	if err != nil {
		return nil, err
	}

	// metrics
	err = loadMetricsCnf(config)
	if err != nil {
		return nil, err
	}

	err = loadTelemtryCnf(config)
	if err != nil {
		return nil, err
	}

	err = loadRedisCnf(config)
	if err != nil {
		return nil, err
	}

	config.Seed = &Seed{
		Need: *seedFlag,
	}

	config.Debug = *debugFlag

	return config, nil
}

func loadAppCnf(config *Config) error {
	port := os.Getenv(envPortKey)
	if port == "" {
		return errors.New("empty app server port in .env file")
	}
	address := os.Getenv(envAddressKey)
	config.App = &App{
		Port:    port,
		Address: address,
		Name:    appName,
	}

	return nil
}

func loadSpotApiCnf(config *Config) error {
	spotGrpcEndpoint := os.Getenv(envSpotEndpointKey)
	if spotGrpcEndpoint == "" {
		return errors.New("empty spot endpoint address in .env file")
	}

	config.Spot = &Spot{
		Endpoint: spotGrpcEndpoint,
	}

	return nil
}

func loadLoggerCnf(config *Config) error {
	logDir := *logsDirFlag
	if logDir == "" {
		logDir = os.Getenv(envLogsDirKey)
	}

	if logDir == "" {
		logDir = defaultLogsDir
	}

	stat, err := os.Stat(logDir)
	if err != nil && os.IsNotExist(err) {
		if err := os.Mkdir(logDir, 0755); err != nil {
			return err
		}
	} else if !stat.IsDir() {
		return fmt.Errorf("logger config error: wait dir in path")
	}

	logpath := filepath.Join(logDir, logFilename)

	config.Log = &Log{
		LogPath: logpath,
	}

	return nil
}

func loadMetricsCnf(config *Config) error {
	metricsPort := os.Getenv(envMetricsPortKey)
	if metricsPort == "" {
		// может в лог и запустить без метрик?
		return errors.New("prometheus port empty in .env")
	}

	config.Metrics = &Metrics{
		Port: metricsPort,
	}

	return nil
}

func loadTelemtryCnf(config *Config) error {
	jaegGrpc := os.Getenv(envJaegerAddressKey)
	if jaegGrpc == "" {
		// может в лог и запустить без метрик?
		return errors.New("telemtry config error. jaeger grpc address empty in .env")
	}

	config.Telemetry = &Telemetry{
		JaegerGrpcAddress: jaegGrpc,
	}

	return nil
}

func loadRedisCnf(config *Config) error {
	redisHost := os.Getenv(envRedisHostKey)
	redisPort := os.Getenv(envRedisPortKey)
	redisAddr := redisHost + ":" + redisPort

	redisPassword := os.Getenv(envRedisPasswordKey)
	redisUsername := os.Getenv(envRedisUsernameKey)

	redisDB := 0
	redisDBStr := os.Getenv(envRedisDBKey)
	if redisDBStr != "" {
		_, err := strconv.Atoi(redisDBStr)
		if err != nil {
			return fmt.Errorf("invalid redis DB value: %w", err)
		}
	}

	redisTTL := 300 * time.Second // default 5 minutes
	redisTTLStr := os.Getenv(envRedisTTLKey)
	if redisTTLStr != "" {
		ttlSeconds := 300
		_, err := fmt.Sscanf(redisTTLStr, "%d", &ttlSeconds)
		if err != nil {
			return fmt.Errorf("invalid redis TTL value: %w", err)
		}
		redisTTL = time.Duration(ttlSeconds) * time.Second
	}

	config.Redis = &Redis{
		Address:  redisAddr,
		Password: redisPassword,
		Username: redisUsername,
		DB:       redisDB,
		TTL:      redisTTL,
	}

	return nil
}
