package v1

import (
	"cmp"
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
	"net/http"
	"slices"
	"strconv"

	"stock-ticker/internal/alphavantage"
)

type Handler struct {
	Client   *alphavantage.Client
	Log      *slog.Logger
	Function string
	Symbol   string
	NDays    int
}

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

func (h *Handler) GetTicker(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	data, err := h.Client.FetchDailyTimeSeries(h.Function, h.Symbol)
	if err != nil {
		h.Log.Error("Failed to fetch upstream data", "error", err)
		http.Error(w, "Failed to fetch upstream data", http.StatusBadGateway)
		return
	}

	keys := slices.SortedFunc(maps.Keys(data.TimeSeries), func(a, b string) int {
		return cmp.Compare(b, a)
	})
	ndays := h.NDays
	if ndays > len(keys) {
		ndays = len(keys)
	}

	ts := make(map[string]alphavantage.DailyDataPoint, ndays)
	var dailyAverage float64
	for _, k := range keys[:ndays] {
		ts[k] = data.TimeSeries[k]
		closeValue, err := strconv.ParseFloat(data.TimeSeries[k].Close, 64)
		if err != nil {
			h.Log.Error("Failed to parse close value", "key", k, "error", err)
			http.Error(w, "Failed to parse close value", http.StatusInternalServerError)
			return
		}
		dailyAverage += closeValue
	}
	if ndays > 0 {
		dailyAverage /= float64(ndays)
	}

	resp := TickerResponse{
		MetaData: TickerMetaData{
			NDays:         strconv.Itoa(ndays),
			Symbol:        h.Symbol,
			LastRefreshed: data.MetaData.LastRefreshed,
			TimeZone:      data.MetaData.TimeZone,
		},
		DailyAverage: fmt.Sprintf("%.2f", dailyAverage),
		TimeSeries:   ts,
	}

	out, err := json.Marshal(resp)
	if err != nil {
		h.Log.Error("Failed to marshal response", "error", err)
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}
	h.Log.Debug("Sending response to client", "bytes", len(out))
	w.Write(out)
}
