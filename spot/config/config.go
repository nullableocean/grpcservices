package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

var (
	defaultEnvPath = "../.env"
	defaultLogPath = "./logs/logs.txt"

	defaultAddress = "127.0.0.1"
)

var (
	envPortKey    = "SPOT_SERVICE_PORT"
	envAddressKey = "SPOT_SERVICE_ADDRESS"
)

type Config struct {
	Port    string
	Address string
	LogPath string
}

func NewConfig() (*Config, error) {
	err := godotenv.Load(defaultEnvPath)
	if err != nil {
		return nil, err
	}

	port := os.Getenv(envPortKey)
	if port == "" {
		return nil, errors.New("empty port in .env file")
	}
	address := os.Getenv(envAddressKey)
	if address == "" {
		address = defaultAddress
	}

	return &Config{
		Port:    port,
		LogPath: defaultLogPath,
	}, nil
}
