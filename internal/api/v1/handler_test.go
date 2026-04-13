package v1

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"stock-ticker/internal/alphavantage"
	"stock-ticker/internal/logging"
)

func TestGetTickerReturnsLatestNDaysAndAverage(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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
				"2026-04-10": {"1. open": "100.00", "2. high": "101.00", "3. low": "99.00", "4. close": "10.00", "5. volume": "1000"},
				"2026-04-09": {"1. open": "100.00", "2. high": "101.00", "3. low": "99.00", "4. close": "20.00", "5. volume": "1000"},
				"2026-04-08": {"1. open": "100.00", "2. high": "101.00", "3. low": "99.00", "4. close": "30.00", "5. volume": "1000"}
			}
		}`))
	}))
	defer server.Close()

	h := &Handler{
		Client:   alphavantage.NewClient(server.URL, "test-key", logging.New("ERROR")),
		Log:      logging.New("ERROR"),
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

	if resp.MetaData.NDays != "2" {
		t.Fatalf("expected NDAYS 2, got %q", resp.MetaData.NDays)
	}
	if resp.DailyAverage != "15.00" {
		t.Fatalf("expected average 15.00, got %q", resp.DailyAverage)
	}
	if len(resp.TimeSeries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(resp.TimeSeries))
	}
	if _, ok := resp.TimeSeries["2026-04-10"]; !ok {
		t.Fatal("expected latest date in response")
	}
	if _, ok := resp.TimeSeries["2026-04-09"]; !ok {
		t.Fatal("expected second latest date in response")
	}
	if _, ok := resp.TimeSeries["2026-04-08"]; ok {
		t.Fatal("did not expect older date in truncated response")
	}
}

func TestGetTickerReturnsBadGatewayOnUpstreamError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "rate limited", http.StatusTooManyRequests)
	}))
	defer server.Close()

	h := &Handler{
		Client:   alphavantage.NewClient(server.URL, "test-key", logging.New("ERROR")),
		Log:      logging.New("ERROR"),
		Function: "TIME_SERIES_DAILY",
		Symbol:   "MSFT",
		NDays:    2,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ticker", nil)
	rec := httptest.NewRecorder()

	h.GetTicker(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected status 502, got %d", rec.Code)
	}
}

func TestGetTickerReturnsInternalServerErrorOnInvalidCloseValue(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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
				"2026-04-10": {"1. open": "100.00", "2. high": "101.00", "3. low": "99.00", "4. close": "not-a-number", "5. volume": "1000"}
			}
		}`))
	}))
	defer server.Close()

	h := &Handler{
		Client:   alphavantage.NewClient(server.URL, "test-key", logging.New("ERROR")),
		Log:      logging.New("ERROR"),
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
}
