package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
)

type Config struct {
	QuoteServiceURL string
	ServeAddr       string
	OTLPEndpoint    string
	APIKey          string
	Function        string
	Symbol          string
	NDays           int
	LogLevel        string
	DisableMetrics  bool
}

func LoadFromEnv() (Config, error) {
	cfg := Config{
		QuoteServiceURL: "https://www.alphavantage.co/query",
		ServeAddr:       ":8080",
		OTLPEndpoint:    "",
		APIKey:          "demo",
		Function:        "TIME_SERIES_DAILY",
		Symbol:          "MSFT",
		NDays:           7,
		LogLevel:        "DEBUG",
	}

	errorList := []error{}

	// Override defaults with environment variables if set and not empty
	if v := os.Getenv("QUOTE_SERVICE_URL"); v != "" {
		cfg.QuoteServiceURL = v
	}
	if v := os.Getenv("API_KEY"); v != "" {
		cfg.APIKey = v
	}
	if v := os.Getenv("SERVICE_ADDR"); v != "" {
		cfg.ServeAddr = v
	}
	if v := os.Getenv("OTLP_ENDPOINT"); v != "" {
		cfg.OTLPEndpoint = v
	}
	if v := os.Getenv("FUNCTION"); v != "" {
		cfg.Function = v
	}
	if v := os.Getenv("SYMBOL"); v != "" {
		cfg.Symbol = v
	}
	if v := os.Getenv("NDAYS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.NDays = n
		} else {
			errorList = append(errorList, fmt.Errorf("Invalid (integer) NDAYS value %q: %v", v, err))
		}

	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		cfg.LogLevel = v
	}
	if v := os.Getenv("DISABLE_METRICS"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			cfg.DisableMetrics = b
		} else {
			errorList = append(errorList, fmt.Errorf("Invalid (boolean) DISABLE_METRICS value %q: %v", v, err))
		}
	}

	if len(errorList) > 0 {
		for _, err := range errorList {
			log.Printf("%v", err)
		}
	} else {
		log.Printf("Config: quoteServiceURL=%s serveAddr=%s otlpEndpoint=%s function=%s symbol=%s NDays=%d logLevel=%s disableMetrics=%v",
			cfg.QuoteServiceURL, cfg.ServeAddr, cfg.OTLPEndpoint, cfg.Function, cfg.Symbol, cfg.NDays, cfg.LogLevel, cfg.DisableMetrics)
	}

	return cfg, errors.Join(errorList...)
}
