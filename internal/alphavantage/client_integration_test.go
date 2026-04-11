//go:build integration
// +build integration

package alphavantage

import (
	"os"
	"testing"

	"stock-ticker/internal/logging"
)

func TestFetchDailyTimeSeriesIntegration(t *testing.T) {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" || apiKey == "demo" {
		t.Skip("set API_KEY to a real Alpha Vantage key and run with -tags=integration")
	}

	symbol := os.Getenv("SYMBOL")
	if symbol == "" {
		symbol = "MSFT"
	}

	function := os.Getenv("FUNCTION")
	if function == "" {
		function = "TIME_SERIES_DAILY"
	}

	client := NewClient("https://www.alphavantage.co/query", apiKey, logging.New("ERROR"))
	resp, err := client.FetchDailyTimeSeries(function, symbol)
	if err != nil {
		t.Fatalf("expected successful upstream response, got error: %v", err)
	}
	if resp.MetaData.Symbol != symbol {
		t.Fatalf("expected symbol %q, got %q", symbol, resp.MetaData.Symbol)
	}
	if len(resp.TimeSeries) == 0 {
		t.Fatal("expected non-empty time series")
	}
}
