package alphavantage

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type Client struct {
	baseURL string
	apiKey  string
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
	}
}

func (c *Client) FetchDailyTimeSeries(function, symbol string) (*DailyTSResponse, error) {
	reqURL := fmt.Sprintf("%s?apikey=%s&function=%s&symbol=%s",
		c.baseURL, c.apiKey, function, symbol)
	log.Printf("DEBUG: Fetching data from %s?apikey=REDACTED&function=%s&symbol=%s",
		c.baseURL, function, symbol)

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
	log.Printf("DEBUG: Received %d bytes from upstream", len(body))

	var result DailyTSResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshaling response: %w", err)
	}

	log.Printf("DEBUG: Parsed %d time series entries for %s", len(result.TimeSeries), symbol)

	if len(result.TimeSeries) == 0 {
		return nil, fmt.Errorf("upstream returned no time series data, raw response: %s", string(body))
	}

	if result.MetaData.Symbol != symbol {
		return nil, fmt.Errorf("symbol mismatch: expected %s, got %s", symbol, result.MetaData.Symbol)
	}

	return &result, nil
}
