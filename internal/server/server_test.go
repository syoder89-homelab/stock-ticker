package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"stock-ticker/internal/logging"
)

func TestBuildMuxProbeEndpoints(t *testing.T) {
	t.Parallel()

	s := New(":8080", nil, logging.New("ERROR"))
	mux := s.buildMux()

	t.Run("healthz is always ok", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("readyz reflects readiness", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected status 503 before ready, got %d", rec.Code)
		}

		s.ready.Store(true)
		rec = httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200 after ready, got %d", rec.Code)
		}
	})

	t.Run("startupz reflects started state", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/startupz", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected status 503 before start, got %d", rec.Code)
		}

		s.started.Store(true)
		rec = httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200 after start, got %d", rec.Code)
		}
	})
}

func TestBuildMuxRejectsNonGetTickerRequests(t *testing.T) {
	t.Parallel()

	s := New(":8080", nil, logging.New("ERROR"))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ticker", nil)
	rec := httptest.NewRecorder()

	s.buildMux().ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", rec.Code)
	}
	if got := rec.Header().Get("Allow"); got != http.MethodGet {
		t.Fatalf("expected Allow GET, got %q", got)
	}
}