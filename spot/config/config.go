package config

import (
	"errors"
	"flag"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

var (
	defaultEnvPath = ".env"
	defaultLogPath = "./logs/logs.txt"

	defaultAddress = "127.0.0.1"
)

var (
	envPortKey    = "SERVER_PORT"
	envAddressKey = "SERVER_ADDRESS"

	envLogPath     = "LOG_FILE"
	envMetricsPort = "PROMETHEUS_PORT"
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
type Config struct {
	Server  *Server
	Metrics *Metrics
	Log     *Log
}

func NewConfig() (*Config, error) {
	envPath := parseEnvFlag()
	err := godotenv.Load(envPath)
	if err != nil {
		return nil, err
	}

	config := &Config{}

	// order server
	err = loadServerCnf(config)
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

	return config, nil
}

func parseEnvFlag() string {
	envPath := flag.String("env", defaultEnvPath, "path to env")
	flag.Parse()

	return *envPath
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
	logPath := os.Getenv(envLogPath)
	if logPath == "" {
		logPath = defaultLogPath
	}

	// если директории нет - создаём
	dir, _ := filepath.Split(logPath)
	if _, e := os.Stat(dir); os.IsNotExist(e) {
		if err := os.Mkdir(dir, 0755); err != nil {
			return err
		}
	}

	config.Log = &Log{
		LogPath: logPath,
	}

	return nil
}

func loadMetricsCnf(config *Config) error {
	metricsPort := os.Getenv(envMetricsPort)
	if metricsPort == "" {
		return errors.New("prometheus port empty in .env")
	}

	config.Metrics = &Metrics{
		Port: metricsPort,
	}

	return nil
}
