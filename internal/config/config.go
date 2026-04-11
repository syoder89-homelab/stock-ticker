package config

import (
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
}

func LoadFromEnv() Config {
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
		}
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		cfg.LogLevel = v
	}

	log.Printf("Config: quoteServiceURL=%s serveAddr=%s otlpEndpoint=%s function=%s symbol=%s NDays=%d logLevel=%s",
		cfg.QuoteServiceURL, cfg.ServeAddr, cfg.OTLPEndpoint, cfg.Function, cfg.Symbol, cfg.NDays, cfg.LogLevel)

	return cfg
}
