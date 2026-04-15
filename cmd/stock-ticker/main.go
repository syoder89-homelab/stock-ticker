package main

import (
	"log"

	"stock-ticker/internal/alphavantage"
	v1 "stock-ticker/internal/api/v1"
	"stock-ticker/internal/config"
	"stock-ticker/internal/logging"
	"stock-ticker/internal/metrics"
	"stock-ticker/internal/server"
)

var (
	version = "dev"
	commit  = "unknown"
)

func main() {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		// All individual config errors logged are logged in the config package
		log.Fatalf("Failed to load configuration!")
	}
	logger := logging.New(cfg.LogLevel)

	var appMetrics *metrics.Metrics
	if !cfg.DisableMetrics {
		appMetrics = metrics.New(version, commit)
	}

	client := alphavantage.NewClient(cfg.QuoteServiceURL, cfg.APIKey, logger, appMetrics)

	handler := &v1.Handler{
		Client:   client,
		Log:      logger,
		Metrics:  appMetrics,
		Function: cfg.Function,
		Symbol:   cfg.Symbol,
		NDays:    cfg.NDays,
	}

	srv := server.New(cfg.ServeAddr, handler, appMetrics, logger)
	if err := srv.Serve(); err != nil {
		log.Fatalf("Server exited: %v", err)
	}
}
