package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	Registry                       *prometheus.Registry
	TickerRequestsTotal            *prometheus.CounterVec
	TickerRequestDurationSeconds   prometheus.Histogram
	TickerErrorsTotal              *prometheus.CounterVec
	UpstreamRequestsTotal          *prometheus.CounterVec
	UpstreamRequestDurationSeconds *prometheus.HistogramVec
	UpstreamErrorsTotal            *prometheus.CounterVec
}

func New(version, commit string) *Metrics {
	registry := prometheus.NewRegistry()

	buildInfo := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "stock_ticker",
		Name:      "build_info",
		Help:      "Build information for the stock-ticker service.",
	}, []string{"version", "commit"})
	buildInfo.WithLabelValues(version, commit).Set(1)

	m := &Metrics{
		Registry: registry,
		TickerRequestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "stock_ticker",
			Name:      "ticker_requests_total",
			Help:      "Total number of stock ticker API requests.",
		}, []string{"status"}),
		TickerRequestDurationSeconds: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: "stock_ticker",
			Name:      "ticker_request_duration_seconds",
			Help:      "Latency of stock ticker API requests.",
			Buckets:   prometheus.DefBuckets,
		}),
		TickerErrorsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "stock_ticker",
			Name:      "ticker_errors_total",
			Help:      "Total number of stock ticker handler errors by kind.",
		}, []string{"kind"}),
		UpstreamRequestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "stock_ticker",
			Name:      "upstream_requests_total",
			Help:      "Total number of Alpha Vantage requests by outcome.",
		}, []string{"function", "symbol", "status"}),
		UpstreamRequestDurationSeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "stock_ticker",
			Name:      "upstream_request_duration_seconds",
			Help:      "Latency of Alpha Vantage requests.",
			Buckets:   []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 30, 60},
		}, []string{"function", "symbol"}),
		UpstreamErrorsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "stock_ticker",
			Name:      "upstream_errors_total",
			Help:      "Total number of Alpha Vantage request errors by kind.",
		}, []string{"function", "symbol", "kind"}),
	}

	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		buildInfo,
		m.TickerRequestsTotal,
		m.TickerRequestDurationSeconds,
		m.TickerErrorsTotal,
		m.UpstreamRequestsTotal,
		m.UpstreamRequestDurationSeconds,
		m.UpstreamErrorsTotal,
	)

	return m
}

func (m *Metrics) Handler() http.Handler {
	if m == nil {
		return http.NotFoundHandler()
	}

	return promhttp.HandlerFor(m.Registry, promhttp.HandlerOpts{})
}

func (m *Metrics) ObserveTickerRequest(statusCode int, duration time.Duration) {
	if m == nil {
		return
	}

	status := strconv.Itoa(statusCode)
	m.TickerRequestsTotal.WithLabelValues(status).Inc()
	m.TickerRequestDurationSeconds.Observe(duration.Seconds())
}

func (m *Metrics) IncTickerError(kind string) {
	if m == nil {
		return
	}

	m.TickerErrorsTotal.WithLabelValues(kind).Inc()
}

func (m *Metrics) ObserveUpstreamRequest(function, symbol, status string, duration time.Duration) {
	if m == nil {
		return
	}

	m.UpstreamRequestsTotal.WithLabelValues(function, symbol, status).Inc()
	m.UpstreamRequestDurationSeconds.WithLabelValues(function, symbol).Observe(duration.Seconds())
}

func (m *Metrics) IncUpstreamError(function, symbol, kind string) {
	if m == nil {
		return
	}

	m.UpstreamErrorsTotal.WithLabelValues(function, symbol, kind).Inc()
}
