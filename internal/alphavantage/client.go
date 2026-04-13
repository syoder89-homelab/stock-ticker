package alphavantage

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

type DailyTSResponse struct {
	MetaData   MetaData                  `json:"Meta Data"`
	TimeSeries map[string]DailyDataPoint `json:"Time Series (Daily)"`
}

type MetaData struct {
	Information   string `json:"1. Information"`
	Symbol        string `json:"2. Symbol"`
	LastRefreshed string `json:"3. Last Refreshed"`
	OutputSize    string `json:"4. Output Size"`
	TimeZone      string `json:"5. Time Zone"`
}

type DailyDataPoint struct {
	Open   string `json:"1. open"`
	High   string `json:"2. high"`
	Low    string `json:"3. low"`
	Close  string `json:"4. close"`
	Volume string `json:"5. volume"`
}

type Client struct {
	baseURL string
	apiKey  string
	log     *slog.Logger
}

func NewClient(baseURL, apiKey string, log *slog.Logger) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		log:     log,
	}
}

func (c *Client) FetchDailyTimeSeries(function, symbol string) (*DailyTSResponse, error) {
	reqURL := fmt.Sprintf("%s?apikey=%s&function=%s&symbol=%s",
		c.baseURL, c.apiKey, function, symbol)
	c.log.Debug("Fetching data from upstream",
		"url", c.baseURL, "function", function, "symbol", symbol)

	resp, err := http.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("GET failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}
	c.log.Debug("Received upstream response", "bytes", len(body))

	var result DailyTSResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling response: %w", err)
	}

	c.log.Debug("Parsed time series entries", "count", len(result.TimeSeries), "symbol", symbol)

	if len(result.TimeSeries) == 0 {
		return nil, fmt.Errorf("upstream returned no time series data, raw response: %s", string(body))
	}

	if result.MetaData.Symbol != symbol {
		return nil, fmt.Errorf("symbol mismatch: expected %s, got %s", symbol, result.MetaData.Symbol)
	}

	return &result, nil
}
