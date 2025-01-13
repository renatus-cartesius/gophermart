package config

import (
	"flag"
	"os"
)

type Config struct {
	SrvAddress    string
	DBURI         string
	AccrualAddres string
}

func LoadConfig() (*Config, error) {

	config := &Config{}

	flag.StringVar(&config.SrvAddress, "a", "localhost:8080", "addres for server exposing")
	flag.StringVar(&config.DBURI, "d", "", "database connection string")
	flag.StringVar(&config.AccrualAddres, "r", "localhost:8081", "accrual address")

	if envSrvAddress := os.Getenv("RUN_ADDRESS"); envSrvAddress != "" {
		config.SrvAddress = envSrvAddress
	}
	if envDBURI := os.Getenv("DATABASE_URI"); envDBURI != "" {
		config.DBURI = envDBURI
	}
	if envAccrualAddress := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccrualAddress != "" {
		config.AccrualAddres = envAccrualAddress
	}

	return config, nil
}
