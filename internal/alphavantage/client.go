package alphavantage

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"stock-ticker/internal/metrics"
)

type Client struct {
	baseURL    string
	apiKey     string
	log        *slog.Logger
	httpClient *http.Client
	metrics    *metrics.Metrics
}

func NewClient(baseURL, apiKey string, log *slog.Logger, metrics *metrics.Metrics) *Client {
	return NewClientWithHTTPClient(baseURL, apiKey, log, http.DefaultClient, metrics)
}

func NewClientWithHTTPClient(baseURL, apiKey string, log *slog.Logger, httpClient *http.Client, metrics *metrics.Metrics) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{
		baseURL:    baseURL,
		apiKey:     apiKey,
		log:        log,
		httpClient: httpClient,
		metrics:    metrics,
	}
}

func (c *Client) FetchDailyTimeSeries(function, symbol string) (*DailyTSResponse, error) {
	start := time.Now()
	status := "success"
	defer func() {
		c.metrics.ObserveUpstreamRequest(function, symbol, status, time.Since(start))
	}()

	reqURL := fmt.Sprintf("%s?apikey=%s&function=%s&symbol=%s",
		c.baseURL, c.apiKey, function, symbol)
	c.log.Debug("Fetching data from upstream",
		"url", c.baseURL, "function", function, "symbol", symbol)

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		status = "request_build_error"
		c.metrics.IncUpstreamError(function, symbol, "request_build")
		return nil, fmt.Errorf("building request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		status = "transport_error"
		c.metrics.IncUpstreamError(function, symbol, "transport")
		return nil, fmt.Errorf("GET failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		status = httpStatusClass(resp.StatusCode)
		c.metrics.IncUpstreamError(function, symbol, "http_status")
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		status = "read_error"
		c.metrics.IncUpstreamError(function, symbol, "read")
		return nil, fmt.Errorf("reading response body: %w", err)
	}
	c.log.Debug("Received upstream response", "bytes", len(body))

	var result DailyTSResponse
	if err := json.Unmarshal(body, &result); err != nil {
		status = "decode_error"
		c.metrics.IncUpstreamError(function, symbol, "decode")
		return nil, fmt.Errorf("unmarshaling response: %w", err)
	}

	c.log.Debug("Parsed time series entries", "count", len(result.TimeSeries), "symbol", symbol)

	if len(result.TimeSeries) == 0 {
		status = "empty_response"
		c.metrics.IncUpstreamError(function, symbol, "empty_response")
		return nil, fmt.Errorf("upstream returned no time series data, raw response: %s", string(body))
	}

	if result.MetaData.Symbol != symbol {
		status = "symbol_mismatch"
		c.metrics.IncUpstreamError(function, symbol, "symbol_mismatch")
		return nil, fmt.Errorf("symbol mismatch: expected %s, got %s", symbol, result.MetaData.Symbol)
	}

	return &result, nil
}

func httpStatusClass(code int) string {
	switch {
	case code >= 200 && code < 300:
		return "2xx"
	case code >= 300 && code < 400:
		return "3xx"
	case code >= 400 && code < 500:
		return "4xx"
	case code >= 500 && code < 600:
		return "5xx"
	default:
		return "other"
	}
}
