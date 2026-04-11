package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestMetricsRecordTickerAndUpstreamCounts(t *testing.T) {
	t.Parallel()

	m := New()
	m.ObserveTickerRequest(200, 0)
	m.IncTickerError("parse_close_value")
	m.ObserveUpstreamRequest("TIME_SERIES_DAILY", "MSFT", "success", 0)
	m.IncUpstreamError("TIME_SERIES_DAILY", "MSFT", "http_status")

	if got := testutil.ToFloat64(m.TickerRequestsTotal.WithLabelValues("/api/v1/ticker", "200")); got != 1 {
		t.Fatalf("expected ticker request count 1, got %v", got)
	}
	if got := testutil.ToFloat64(m.TickerErrorsTotal.WithLabelValues("parse_close_value")); got != 1 {
		t.Fatalf("expected ticker error count 1, got %v", got)
	}
	if got := testutil.ToFloat64(m.UpstreamRequestsTotal.WithLabelValues("TIME_SERIES_DAILY", "MSFT", "success")); got != 1 {
		t.Fatalf("expected upstream request count 1, got %v", got)
	}
	if got := testutil.ToFloat64(m.UpstreamErrorsTotal.WithLabelValues("TIME_SERIES_DAILY", "MSFT", "http_status")); got != 1 {
		t.Fatalf("expected upstream error count 1, got %v", got)
	}
}
