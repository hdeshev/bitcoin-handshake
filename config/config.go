package config

import (
	"os"
)

type Config struct {
	BTCNodeAddress string
}

func New() *Config {
	return &Config{
		BTCNodeAddress: getEnv("BTC_NODE_ADDRESS", "localhost:18444"),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
