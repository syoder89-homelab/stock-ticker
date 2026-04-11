package server

import (
	"log/slog"
	"net/http"
	"sync/atomic"

	v1 "stock-ticker/internal/api/v1"
)

type Server struct {
	addr      string
	v1Handler *v1.Handler
	log       *slog.Logger
	started   atomic.Bool
	ready     atomic.Bool
}

func New(addr string, v1Handler *v1.Handler, log *slog.Logger) *Server {
	return &Server{
		addr:      addr,
		v1Handler: v1Handler,
		log:       log,
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
		s.v1Handler.GetTicker(w, r)
	})
	return mux
}

func (s *Server) Serve() error {
	srv := &http.Server{
		Addr:    s.addr,
		Handler: s.buildMux(),
	}
	s.started.Store(true)
	s.ready.Store(true)
	s.log.Info("Server listening", "addr", s.addr)
	return srv.ListenAndServe()
}
