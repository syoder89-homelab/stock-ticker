package main

import (
	"log"

	"stock-ticker/internal/alphavantage"
	v1 "stock-ticker/internal/api/v1"
	"stock-ticker/internal/config"
	"stock-ticker/internal/server"
)

func main() {
	cfg := config.LoadFromEnv()

	client := alphavantage.NewClient(cfg.QuoteServiceURL, cfg.APIKey)

	handler := &v1.Handler{
		Client:   client,
		Function: cfg.Function,
		Symbol:   cfg.Symbol,
		NDays:    cfg.NDays,
	}

	srv := server.New(cfg.ServeAddr, handler)
	if err := srv.Serve(); err != nil {
		log.Fatalf("Server exited: %v", err)
	}
}
