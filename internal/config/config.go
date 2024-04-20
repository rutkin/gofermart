package config

import (
	"flag"
	"os"
)

type config struct {
	LogLevel             string
	RunAddress           string
	DatabaseURI          string
	AccrualSystemAddress string
}

func MakeConfig() *config {
	config := &config{LogLevel: "info"}
	flag.StringVar(&config.RunAddress, "a", "", "run address")
	flag.StringVar(&config.DatabaseURI, "d", "", "database uri")
	flag.StringVar(&config.AccrualSystemAddress, "r", "", "accrual system address")
	flag.Parse()

	if runAddress, ok := os.LookupEnv("RUN_ADDRESS"); ok {
		config.RunAddress = runAddress
	}

	if databaseURI, ok := os.LookupEnv("DATABASE_URI"); ok {
		config.DatabaseURI = databaseURI
	}

	if accrualSystemAddress, ok := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS"); ok {
		config.AccrualSystemAddress = accrualSystemAddress
	}

	return config
}
