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
	"time"

	"stock-ticker/internal/alphavantage"
	"stock-ticker/internal/metrics"
)

type Handler struct {
	Client   *alphavantage.Client
	Log      *slog.Logger
	Metrics  *metrics.Metrics
	Function string
	Symbol   string
	NDays    int
	sr       *statusRecorder
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.status = code
	sr.ResponseWriter.WriteHeader(code)
}

func (h *Handler) beginRequest(w http.ResponseWriter) time.Time {
	h.sr = &statusRecorder{ResponseWriter: w, status: http.StatusOK}
	h.sr.Header().Set("Content-Type", "application/json")
	return time.Now()
}

func (h *Handler) endRequest(start time.Time) {
	h.Metrics.ObserveTickerRequest(h.sr.status, time.Since(start))
}

func (h *Handler) writeJSON(data []byte) {
	h.Log.Debug("Sending response to client", "bytes", len(data))
	h.sr.Write(data)
}

func (h *Handler) httpError(msg string, code int, metricKind string) {
	h.Log.Error(msg, "status", code)
	h.Metrics.IncTickerError(metricKind)
	http.Error(h.sr, msg, code)
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
	start := h.beginRequest(w)
	defer h.endRequest(start)

	data, err := h.Client.FetchDailyTimeSeries(h.Function, h.Symbol)
	if err != nil {
		// Note: this isn't really double-counting errors - it's incrementing the ticker_errors_total not upstream_errors_total
		h.httpError("Failed to fetch upstream data", http.StatusBadGateway, "upstream_fetch")
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
			h.httpError("Failed to parse close value", http.StatusInternalServerError, "parse_close_value")
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
		h.httpError("Failed to marshal response", http.StatusInternalServerError, "marshal_response")
		return
	}
	h.writeJSON(out)
}
