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
	defaultAddress = "127.0.0.1"

	defaultEnvPath = ".env"

	defaultLogsDir = "./logs"
	logFilename    = "logs.log"
)

var (
	envPortKey    = "SERVER_PORT"
	envAddressKey = "SERVER_ADDRESS"

	envLogsDirKey     = "LOGS_DIR"
	envMetricsPortKey = "METRICS_PORT"
)

var (
	envPathFlag = flag.String("env", defaultEnvPath, "path to env")
	logsDirFlag = flag.String("logdir", defaultLogsDir, "dir for logs")
	seedFlag    = flag.Bool("seed", false, "path to logs")
)

type Metrics struct {
	Port string
}

type Server struct {
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
	Server  *Server
	Metrics *Metrics
	Log     *Log
	Seed    *Seed
}

func NewConfig() (*Config, error) {
	flag.Parse()

	envPath := *envPathFlag
	godotenv.Load(envPath)

	config := &Config{}

	err := loadServerCnf(config)
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

	config.Seed = &Seed{
		Need: *seedFlag,
	}

	return config, nil
}

func loadServerCnf(config *Config) error {
	port := os.Getenv(envPortKey)
	if port == "" {
		return errors.New("empty port in .env file")
	}
	address := os.Getenv(envAddressKey)
	if address == "" {
		address = defaultAddress
	}

	config.Server = &Server{
		Port:    port,
		Address: address,
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
