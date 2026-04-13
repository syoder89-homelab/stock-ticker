package alphavantage

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"stock-ticker/internal/logging"
	"stock-ticker/internal/metrics"
)

func TestFetchDailyTimeSeriesSuccess(t *testing.T) {
	t.Parallel()

	var gotAPIKey string
	var gotFunction string
	var gotSymbol string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAPIKey = r.URL.Query().Get("apikey")
		gotFunction = r.URL.Query().Get("function")
		gotSymbol = r.URL.Query().Get("symbol")

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"Meta Data": {
				"1. Information": "Daily Prices",
				"2. Symbol": "MSFT",
				"3. Last Refreshed": "2026-04-10",
				"4. Output Size": "Compact",
				"5. Time Zone": "US/Eastern"
			},
			"Time Series (Daily)": {
				"2026-04-10": {
					"1. open": "100.00",
					"2. high": "101.00",
					"3. low": "99.00",
					"4. close": "100.50",
					"5. volume": "1000"
				}
			}
		}`))
	}))
	defer server.Close()

	m := metrics.New("test", "test")
	client := NewClientWithHTTPClient(server.URL, "test-key", logging.New("ERROR"), server.Client(), m)
	resp, err := client.FetchDailyTimeSeries("TIME_SERIES_DAILY", "MSFT")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if gotAPIKey != "test-key" {
		t.Fatalf("expected api key test-key, got %q", gotAPIKey)
	}
	if gotFunction != "TIME_SERIES_DAILY" {
		t.Fatalf("expected function TIME_SERIES_DAILY, got %q", gotFunction)
	}
	if gotSymbol != "MSFT" {
		t.Fatalf("expected symbol MSFT, got %q", gotSymbol)
	}
	if resp.MetaData.Symbol != "MSFT" {
		t.Fatalf("expected symbol MSFT, got %q", resp.MetaData.Symbol)
	}
	if len(resp.TimeSeries) != 1 {
		t.Fatalf("expected 1 time series entry, got %d", len(resp.TimeSeries))
	}
}

func TestFetchDailyTimeSeriesRecordsSuccessMetrics(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"Meta Data": {
				"1. Information": "Daily Prices",
				"2. Symbol": "MSFT",
				"3. Last Refreshed": "2026-04-11",
				"4. Output Size": "Compact",
				"5. Time Zone": "US/Eastern"
			},
			"Time Series (Daily)": {
				"2026-04-11": {
					"1. open": "100.00",
					"2. high": "101.00",
					"3. low": "99.00",
					"4. close": "100.50",
					"5. volume": "1000"
				}
			}
		}`))
	}))
	defer server.Close()

	m := metrics.New("test", "test")
	client := NewClientWithHTTPClient(server.URL, "test-key", logging.New("ERROR"), server.Client(), m)

	_, err := client.FetchDailyTimeSeries("TIME_SERIES_DAILY", "MSFT")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if got := testutil.ToFloat64(m.UpstreamRequestsTotal.WithLabelValues("TIME_SERIES_DAILY", "MSFT", "success")); got != 1 {
		t.Fatalf("expected upstream success count 1, got %v", got)
	}
}

func TestFetchDailyTimeSeriesRecordsErrorMetrics(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "rate limited", http.StatusTooManyRequests)
	}))
	defer server.Close()

	m := metrics.New("test", "test")
	client := NewClientWithHTTPClient(server.URL, "test-key", logging.New("ERROR"), server.Client(), m)

	_, err := client.FetchDailyTimeSeries("TIME_SERIES_DAILY", "MSFT")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if got := testutil.ToFloat64(m.UpstreamRequestsTotal.WithLabelValues("TIME_SERIES_DAILY", "MSFT", "4xx")); got != 1 {
		t.Fatalf("expected upstream 4xx count 1, got %v", got)
	}
	if got := testutil.ToFloat64(m.UpstreamErrorsTotal.WithLabelValues("TIME_SERIES_DAILY", "MSFT", "http_status")); got != 1 {
		t.Fatalf("expected upstream error count 1, got %v", got)
	}
}

func TestFetchDailyTimeSeriesUnexpectedStatus(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusTooManyRequests)
	}))
	defer server.Close()

	m := metrics.New("test", "test")
	client := NewClientWithHTTPClient(server.URL, "test-key", logging.New("ERROR"), server.Client(), m)
	_, err := client.FetchDailyTimeSeries("TIME_SERIES_DAILY", "MSFT")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "unexpected status code: 429") {
		t.Fatalf("expected status code error, got %v", err)
	}
}

func TestFetchDailyTimeSeriesRejectsSymbolMismatch(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"Meta Data": {
				"1. Information": "Daily Prices",
				"2. Symbol": "AAPL",
				"3. Last Refreshed": "2026-04-10",
				"4. Output Size": "Compact",
				"5. Time Zone": "US/Eastern"
			},
			"Time Series (Daily)": {
				"2026-04-10": {
					"1. open": "100.00",
					"2. high": "101.00",
					"3. low": "99.00",
					"4. close": "100.50",
					"5. volume": "1000"
				}
			}
		}`))
	}))
	defer server.Close()

	m := metrics.New("test", "test")
	client := NewClientWithHTTPClient(server.URL, "test-key", logging.New("ERROR"), server.Client(), m)
	_, err := client.FetchDailyTimeSeries("TIME_SERIES_DAILY", "MSFT")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "symbol mismatch") {
		t.Fatalf("expected symbol mismatch error, got %v", err)
	}
}
