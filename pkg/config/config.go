package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"go-simpler.org/env"
)

var Config AppConfig

type AppConfig struct {
	JwtSecret     string `env:"JWT_SECRET"`
	ServerAddress string `env:"SERVER_ADDRESS"`
	DB            DBConfig
}

type DBConfig struct {
	Port     int    `env:"DB_PORT"`
	Host     string `env:"DB_HOST"`
	Password string `env:"DB_PASSWORD"`
	User     string `env:"DB_USER"`
	Database string `env:"DB_DATABASE"`
}

func LoadConfig() (*AppConfig, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading env variables: %v", err)
	}

	err = env.Load(&Config, nil)
	if err != nil {
		return nil, fmt.Errorf("error building config object: %v", err)
	}

	return &Config, nil
}
