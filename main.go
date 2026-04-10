package main

import (
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"maps"
	"net/http"
	"os"
	"slices"
	"strconv"
	"sync/atomic"
)

func main() {
	quoteServiceUrl := "https://www.alphavantage.co/query"
	if v := os.Getenv("QUOTE_SERVICE_URL"); v != "" {
		quoteServiceUrl = v
	}
	apiKey := "demo"
	if v := os.Getenv("API_KEY"); v != "" {
		apiKey = v
	}

	serveAddr := ":8080"
	if v := os.Getenv("SERVICE_ADDR"); v != "" {
		serveAddr = v
	}

	otlpEndpoint := ""
	if v := os.Getenv("OTLP_ENDPOINT"); v != "" {
		otlpEndpoint = v
	}

	function := "TIME_SERIES_DAILY_ADJUSTED"
	if v := os.Getenv("FUNCTION"); v != "" {
		function = v
	}

	symbol := "MSFT"
	if v := os.Getenv("SYMBOL"); v != "" {
		symbol = v
	}

	NDAYS := 7
	if v := os.Getenv("NDAYS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			NDAYS = n
		}
	}

	cfg := config{
		quoteServiceURL: quoteServiceUrl,
		serveAddr:       serveAddr,
		oltpEndpoint:    otlpEndpoint,
		apiKey:          apiKey,
		function:        function,
		symbol:          symbol,
		NDAYS:           NDAYS,
	}

	server := NewServer(cfg)
	server.serve()
}

type config struct {
	quoteServiceURL string
	serveAddr       string
	oltpEndpoint    string
	apiKey          string
	function        string
	symbol          string
	NDAYS           int
}

type Server struct {
	cfg     config
	started atomic.Bool
	ready   atomic.Bool
}

type StockTickerAPIMetaData struct {
	NDAYS         string `json:"NDAYS"`
	Symbol        string `json:"Symbol"`
	LastRefreshed string `json:"Last Refreshed"`
	TimeZone      string `json:"Time Zone"`
}

type StockTickerAPIResponse struct {
	MetaData     StockTickerAPIMetaData    `json:"Meta Data"`
	DailyAverage string                    `json:"Daily Average"`
	TimeSeries   map[string]DailyDataPoint `json:"Time Series (Daily)"`
}

type DailyAdjustedResponse struct {
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
	Open             string `json:"1. open"`
	High             string `json:"2. high"`
	Low              string `json:"3. low"`
	Close            string `json:"4. close"`
	AdjustedClose    string `json:"5. adjusted close"`
	Volume           string `json:"6. volume"`
	DividendAmount   string `json:"7. dividend amount"`
	SplitCoefficient string `json:"8. split coefficient"`
}

func NewServer(cfg config) *Server {
	return &Server{
		cfg: cfg,
	}
}

func (s *Server) buildMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		if !s.ready.Load() {
			http.Error(w, "not ready", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/startupz", func(w http.ResponseWriter, _ *http.Request) {
		if !s.started.Load() {
			http.Error(w, "not started", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/api/v1/ticker", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.Header().Set("Allow", "GET")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		s.getData(w, r)
	})
	return mux
}

func (s *Server) serve() {
	srv := &http.Server{
		Addr:    s.cfg.serveAddr,
		Handler: s.buildMux(),
	}
	// Mark us as started and ready
	s.started.Store(true)
	s.ready.Store(true)
	log.Printf("Server listening on %s", s.cfg.serveAddr)
	srv.ListenAndServe()
}

func (s *Server) getData(w http.ResponseWriter, r *http.Request) {
	// apikey=demo&function=TIME_SERIES_DAILY_ADJUSTED&symbol=MSFT
	reqUrl := fmt.Sprintf(
		"%s?apikey=%s&function=%s&symbol=%s",
		s.cfg.quoteServiceURL, s.cfg.apiKey, s.cfg.function, s.cfg.symbol)
	/* Need to sanitize / remove the apikey
	log.Printf("Fetching data from URL: %s", reqUrl)
	*/
	response, err := http.Get(reqUrl)
	if err != nil {
		log.Printf("ERROR: Failed to fetch data, GET failed: %v", err)
		http.Error(w, "Failed to fetch data", http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		log.Printf("ERROR: Failed to fetch data, status code: %d", response.StatusCode)
		http.Error(w, "Failed to fetch data", http.StatusInternalServerError)
		return
	}
	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		log.Printf("ERROR: Failed to read response: %v", err)
		http.Error(w, "Failed to read response", http.StatusInternalServerError)
		return
	}
	// FOR TESTING: read sample-reply.json instead of fetching from API
	/*
		responseData, err := os.ReadFile("sample-reply.json")
		if err != nil {
			log.Printf("ERROR: Failed to read sample-reply.json: %v", err)
			http.Error(w, "Failed to read sample data", http.StatusInternalServerError)
			return
		}
	*/
	// Now process the received data
	var responseObject DailyAdjustedResponse
	if err := json.Unmarshal(responseData, &responseObject); err != nil {
		log.Printf("ERROR: Failed to unmarshal response: %v", err)
		http.Error(w, "Failed to unmarshal response", http.StatusInternalServerError)
		return
	}
	var stResponseObject StockTickerAPIResponse
	stResponseObject.MetaData = StockTickerAPIMetaData{
		NDAYS:         strconv.Itoa(s.cfg.NDAYS),
		Symbol:        responseObject.MetaData.Symbol,
		LastRefreshed: responseObject.MetaData.LastRefreshed,
		TimeZone:      responseObject.MetaData.TimeZone,
	}
	// Sort date keys descending (most recent first), take first NDAYS
	keys := slices.SortedFunc(maps.Keys(responseObject.TimeSeries), func(a, b string) int {
		return cmp.Compare(b, a)
	})
	ndays := s.cfg.NDAYS
	if ndays > len(keys) {
		ndays = len(keys)
	}
	stResponseObject.TimeSeries = make(map[string]DailyDataPoint, ndays)
	var dailyAverage float64 = 0
	for _, k := range keys[:ndays] {
		stResponseObject.TimeSeries[k] = responseObject.TimeSeries[k]
		closeValue, err := strconv.ParseFloat(responseObject.TimeSeries[k].AdjustedClose, 64)
		if err != nil {
			log.Printf("ERROR: Failed to parse close value: %v", err)
			http.Error(w, "Failed to parse close value", http.StatusInternalServerError)
			return
		}
		dailyAverage += closeValue
	}
	if ndays > 0 {
		dailyAverage /= float64(ndays)
	}
	stResponseObject.DailyAverage = fmt.Sprintf("%.2f", dailyAverage)

	w.Header().Set("Content-Type", "application/json")
	stResponseData, err := json.Marshal(stResponseObject)
	if err != nil {
		log.Printf("ERROR: Failed to marshal response: %v", err)
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}
	w.Write(stResponseData)
}
