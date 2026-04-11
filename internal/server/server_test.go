package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"stock-ticker/internal/logging"
	"stock-ticker/internal/metrics"
)

func TestBuildMuxServesMetrics(t *testing.T) {
	t.Parallel()

	m := metrics.New()
	m.IncTickerError("test")
	s := New(":8080", nil, m, logging.New("ERROR"))
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()

	s.buildMux().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	body, err := io.ReadAll(rec.Body)
	if err != nil {
		t.Fatalf("failed reading metrics response: %v", err)
	}
	if !strings.Contains(string(body), "stock_ticker_ticker_errors_total") {
		t.Fatalf("expected custom metric name in response, got %q", string(body))
	}
}
