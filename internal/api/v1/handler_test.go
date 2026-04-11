package v1

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"stock-ticker/internal/alphavantage"
	"stock-ticker/internal/logging"
	"stock-ticker/internal/metrics"
)

func TestGetTickerRecordsSuccessMetrics(t *testing.T) {
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
				"2026-04-11": {"1. open": "100.00", "2. high": "101.00", "3. low": "99.00", "4. close": "10.00", "5. volume": "1000"},
				"2026-04-10": {"1. open": "100.00", "2. high": "101.00", "3. low": "99.00", "4. close": "20.00", "5. volume": "1000"}
			}
		}`))
	}))
	defer server.Close()

	m := metrics.New()
	h := &Handler{
		Client:   alphavantage.NewClientWithHTTPClient(server.URL, "test-key", logging.New("ERROR"), server.Client(), m),
		Log:      logging.New("ERROR"),
		Metrics:  m,
		Function: "TIME_SERIES_DAILY",
		Symbol:   "MSFT",
		NDays:    2,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ticker", nil)
	rec := httptest.NewRecorder()

	h.GetTicker(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp TickerResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.DailyAverage != "15.00" {
		t.Fatalf("expected average 15.00, got %q", resp.DailyAverage)
	}
	if got := testutil.ToFloat64(m.TickerRequestsTotal.WithLabelValues("/api/v1/ticker", "200")); got != 1 {
		t.Fatalf("expected ticker request count 1, got %v", got)
	}
}

func TestGetTickerRecordsErrorMetrics(t *testing.T) {
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
				"2026-04-11": {"1. open": "100.00", "2. high": "101.00", "3. low": "99.00", "4. close": "not-a-number", "5. volume": "1000"}
			}
		}`))
	}))
	defer server.Close()

	m := metrics.New()
	h := &Handler{
		Client:   alphavantage.NewClientWithHTTPClient(server.URL, "test-key", logging.New("ERROR"), server.Client(), m),
		Log:      logging.New("ERROR"),
		Metrics:  m,
		Function: "TIME_SERIES_DAILY",
		Symbol:   "MSFT",
		NDays:    1,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ticker", nil)
	rec := httptest.NewRecorder()

	h.GetTicker(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
	if got := testutil.ToFloat64(m.TickerRequestsTotal.WithLabelValues("/api/v1/ticker", "500")); got != 1 {
		t.Fatalf("expected ticker 500 count 1, got %v", got)
	}
	if got := testutil.ToFloat64(m.TickerErrorsTotal.WithLabelValues("parse_close_value")); got != 1 {
		t.Fatalf("expected parse_close_value error count 1, got %v", got)
	}
}
