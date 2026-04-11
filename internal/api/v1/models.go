package v1

import "stock-ticker/internal/alphavantage"

type TickerResponse struct {
	MetaData     TickerMetaData                         `json:"Meta Data"`
	DailyAverage string                                 `json:"Daily Average"`
	TimeSeries   map[string]alphavantage.DailyDataPoint `json:"Time Series (Daily)"`
}

type TickerMetaData struct {
	NDays         string `json:"NDAYS"`
	Symbol        string `json:"Symbol"`
	LastRefreshed string `json:"Last Refreshed"`
	TimeZone      string `json:"Time Zone"`
}
