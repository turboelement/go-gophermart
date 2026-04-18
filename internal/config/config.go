package config

import (
	"errors"
	"flag"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

type Config struct {
	RunAddr              string
	DatabaseDSN          string
	AccrualSystemAddress string
	CookieSecret         string
	CookieSecure         bool
}

const (
	envRunAddr              = "RUN_ADDRESS"
	envDatabaseDSN          = "DATABASE_URI"
	envAccrualSystemAddress = "ACCRUAL_SYSTEM_ADDRESS"
	envCookieSecret         = "COOKIE_SECRET"
	envCookieSecure         = "COOKIE_SECURE"

	defaultRunAddr = "localhost:8080"
)

func New() (*Config, error) {
	cfg := &Config{
		RunAddr:              defaultRunAddr,
		DatabaseDSN:          "",
		AccrualSystemAddress: "",
		CookieSecret:         "",
		CookieSecure:         false,
	}

	// parsing flags
	flag.StringVar(&cfg.RunAddr, "a", cfg.RunAddr, "Service run address and port")
	flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "DB connection address")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", cfg.AccrualSystemAddress, "Accrual system address")
	flag.StringVar(&cfg.CookieSecret, "s", cfg.CookieSecret, "Secret key for cookie signing")
	flag.BoolVar(&cfg.CookieSecure, "secure", cfg.CookieSecure, "Secure flag for cookies (true for HTTPS)")
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
	if v, ok := os.LookupEnv(envCookieSecure); ok {
		if secure, err := strconv.ParseBool(v); err == nil {
			cfg.CookieSecure = secure
		}
	}

	if cfg.RunAddr == "" {
		return nil, errors.New("server address is empty")
	}

	if cfg.CookieSecret == "" {
		cfg.CookieSecret = uuid.NewString()
	}

	if !cfg.CookieSecure {
		cfg.CookieSecure = shouldUseSecureCookie(cfg.RunAddr)
	}

	return cfg, nil
}

func shouldUseSecureCookie(runAddr string) bool {
	if strings.HasPrefix(runAddr, "https://") {
		return true
	}

	if strings.Contains(runAddr, ":443") {
		return true
	}

	if !strings.Contains(runAddr, "localhost") &&
		!strings.Contains(runAddr, "127.0.0.1") &&
		!strings.Contains(runAddr, "0.0.0.0") &&
		!strings.Contains(runAddr, "::1") {
		return true
	}

	return false
}
