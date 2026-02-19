package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

var (
	appName = "spot-instrument"

	defaultEnvPath = ".env"

	defaultLogsDir = "./logs"
	logFilename    = "logs.log"
)

var (
	envPortKey    = "SERVER_PORT"
	envAddressKey = "SERVER_ADDRESS"

	envLogsDirKey       = "LOGS_DIR"
	envMetricsPortKey   = "METRICS_PORT"
	envJaegerAddressKey = "JUEGER_GRPC_ADDRESS"
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

type App struct {
	Name    string
	Port    string
	Address string
}

type Log struct {
	LogPath string
}

type Seed struct {
	Need bool
}

type Config struct {
	App       *App
	Metrics   *Metrics
	Telemetry *Telemetry
	Log       *Log
	Seed      *Seed

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

	// metrics
	err = loadTelemetryCnf(config)
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
		return errors.New("empty port in .env file")
	}
	address := os.Getenv(envAddressKey)

	config.App = &App{
		Port:    port,
		Address: address,
		Name:    appName,
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
		return errors.New("prometheus port empty in .env")
	}

	config.Metrics = &Metrics{
		Port: metricsPort,
	}

	return nil
}

func loadTelemetryCnf(config *Config) error {
	jaegGrpc := os.Getenv(envJaegerAddressKey)
	if jaegGrpc == "" {
		return errors.New("telemetry config error: jaeger addres empty in .env")
	}

	config.Telemetry = &Telemetry{
		JaegerGrpcAddress: jaegGrpc,
	}

	return nil
}
