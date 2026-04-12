package config

import (
	"errors"
	"flag"
	"os"

	"github.com/google/uuid"
)

type Config struct {
	RunAddr              string
	DatabaseDSN          string
	AccrualSystemAddress string
	CookieSecret         string
}

const (
	envRunAddr              = "RUN_ADDRESS"
	envDatabaseDSN          = "DATABASE_URI"
	envAccrualSystemAddress = "ACCRUAL_SYSTEM_ADDRESS"
	envCookieSecret         = "COOKIE_SECRET"

	defaultRunAddr = "localhost:8080"
)

func New() (*Config, error) {
	cfg := &Config{
		RunAddr:              defaultRunAddr,
		DatabaseDSN:          "",
		AccrualSystemAddress: "",
		CookieSecret:         "",
	}

	// parsing flags
	flag.StringVar(&cfg.RunAddr, "a", cfg.RunAddr, "Service run address and port")
	flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "DB connection address")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", cfg.AccrualSystemAddress, "Accrual system address")
	flag.StringVar(&cfg.CookieSecret, "s", cfg.CookieSecret, "Secret key for cookie signing")
	flag.Parse()

	// parsing env
	if v, ok := os.LookupEnv(envRunAddr); ok {
		cfg.RunAddr = v
	}
	if v, ok := os.LookupEnv(envDatabaseDSN); ok {
		cfg.DatabaseDSN = v
	}
	if v, ok := os.LookupEnv(envAccrualSystemAddress); ok {
		cfg.AccrualSystemAddress = v
	}
	if v, ok := os.LookupEnv(envCookieSecret); ok {
		cfg.CookieSecret = v
	}

	if cfg.RunAddr == "" {
		return nil, errors.New("server address is empty")
	}

	if cfg.CookieSecret == "" {
		cfg.CookieSecret = uuid.NewString()
	}

	return cfg, nil
}
