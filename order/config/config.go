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
	envPortKey    = "ORDER_SERVICE_PORT"
	envAddressKey = "ORDER_SERVICE_ADDRESS"

	envSpotPortKey    = "SPOT_SERVICE_PORT"
	envSpotAddressKey = "SPOT_SERVICE_ADDRESS"
)

type Config struct {
	Port               string
	Address            string
	SpotServicePort    string
	SpotServiceAddress string
	LogPath            string
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

	spotServicePort := os.Getenv(envSpotPortKey)
	if spotServicePort == "" {
		return nil, errors.New("empty spot service port in .env file")
	}

	spotAddress := os.Getenv(envSpotAddressKey)
	if spotAddress == "" {
		return nil, errors.New("empty spot address in .env file")
	}

	return &Config{
		Port:               port,
		Address:            address,
		LogPath:            defaultLogPath,
		SpotServicePort:    spotServicePort,
		SpotServiceAddress: spotAddress,
	}, nil
}
