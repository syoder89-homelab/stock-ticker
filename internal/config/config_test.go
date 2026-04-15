package config

import (
	"strings"
	"testing"
)

func TestLoadFromEnvDefaults(t *testing.T) {
	t.Setenv("QUOTE_SERVICE_URL", "")
	t.Setenv("SERVICE_ADDR", "")
	t.Setenv("OTLP_ENDPOINT", "")
	t.Setenv("API_KEY", "")
	t.Setenv("FUNCTION", "")
	t.Setenv("SYMBOL", "")
	t.Setenv("NDAYS", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("DISABLE_METRICS", "")

	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.QuoteServiceURL != "https://www.alphavantage.co/query" {
		t.Fatalf("unexpected QuoteServiceURL: %q", cfg.QuoteServiceURL)
	}
	if cfg.ServeAddr != ":8080" {
		t.Fatalf("unexpected ServeAddr: %q", cfg.ServeAddr)
	}
	if cfg.APIKey != "demo" {
		t.Fatalf("unexpected APIKey: %q", cfg.APIKey)
	}
	if cfg.Function != "TIME_SERIES_DAILY" {
		t.Fatalf("unexpected Function: %q", cfg.Function)
	}
	if cfg.Symbol != "MSFT" {
		t.Fatalf("unexpected Symbol: %q", cfg.Symbol)
	}
	if cfg.NDays != 7 {
		t.Fatalf("unexpected NDays: %d", cfg.NDays)
	}
	if cfg.LogLevel != "DEBUG" {
		t.Fatalf("unexpected LogLevel: %q", cfg.LogLevel)
	}
	if cfg.DisableMetrics != false {
		t.Fatal("expected DisableMetrics to be false by default")
	}
}

func TestLoadFromEnvOverrides(t *testing.T) {
	t.Setenv("QUOTE_SERVICE_URL", "http://localhost:9090/query")
	t.Setenv("SERVICE_ADDR", ":9090")
	t.Setenv("OTLP_ENDPOINT", "otel-collector:4317")
	t.Setenv("API_KEY", "secret")
	t.Setenv("FUNCTION", "TIME_SERIES_DAILY_ADJUSTED")
	t.Setenv("SYMBOL", "AAPL")
	t.Setenv("NDAYS", "30")
	t.Setenv("LOG_LEVEL", "INFO")
	t.Setenv("DISABLE_METRICS", "true")

	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.QuoteServiceURL != "http://localhost:9090/query" {
		t.Fatalf("unexpected QuoteServiceURL: %q", cfg.QuoteServiceURL)
	}
	if cfg.ServeAddr != ":9090" {
		t.Fatalf("unexpected ServeAddr: %q", cfg.ServeAddr)
	}
	if cfg.OTLPEndpoint != "otel-collector:4317" {
		t.Fatalf("unexpected OTLPEndpoint: %q", cfg.OTLPEndpoint)
	}
	if cfg.APIKey != "secret" {
		t.Fatalf("unexpected APIKey: %q", cfg.APIKey)
	}
	if cfg.Function != "TIME_SERIES_DAILY_ADJUSTED" {
		t.Fatalf("unexpected Function: %q", cfg.Function)
	}
	if cfg.Symbol != "AAPL" {
		t.Fatalf("unexpected Symbol: %q", cfg.Symbol)
	}
	if cfg.NDays != 30 {
		t.Fatalf("unexpected NDays: %d", cfg.NDays)
	}
	if cfg.LogLevel != "INFO" {
		t.Fatalf("unexpected LogLevel: %q", cfg.LogLevel)
	}
	if cfg.DisableMetrics != true {
		t.Fatal("expected DisableMetrics to be true")
	}
}

func TestLoadFromEnvInvalidNDays(t *testing.T) {
	t.Setenv("NDAYS", "not-a-number")

	_, err := LoadFromEnv()

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "Invalid (integer) NDAYS value") {
		t.Fatalf("unexpected error message: %v", err)
	}
}
