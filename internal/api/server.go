package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"time"

	"github.com/jedediah/wx01/internal/db"
)

type Server struct {
	httpServer *http.Server
	queries    *db.Queries
}

func NewServer(port int, pool db.Pool, distFS fs.FS) *Server {
	s := &Server{
		queries: db.NewQueries(pool),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", s.handleHealth)
	mux.HandleFunc("GET /api/current", s.handleCurrent)
	mux.HandleFunc("GET /api/temperature", s.handleTemperature)
	mux.HandleFunc("GET /api/wind", s.handleWind)
	mux.HandleFunc("GET /api/rain", s.handleRain)
	mux.HandleFunc("GET /api/pressure", s.handlePressure)
	mux.HandleFunc("GET /api/solar", s.handleSolar)
	mux.HandleFunc("GET /api/humidity", s.handleHumidity)
	mux.HandleFunc("GET /api/uv", s.handleUV)

	if distFS != nil {
		mux.Handle("/", frontendHandler(distFS))
	}

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	return s
}

func (s *Server) Run(ctx context.Context) {
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.httpServer.Shutdown(shutdownCtx)
	}()

	log.Printf("http server listening on %s", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
		log.Printf("http server error: %v", err)
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{"status": "ok"})
}

func (s *Server) handleCurrent(w http.ResponseWriter, r *http.Request) {
	cond, err := s.queries.CurrentConditions(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if cond == nil {
		writeJSON(w, nil)
		return
	}

	rainLastHour, _ := s.queries.RainLastHour(r.Context())
	rainToday, _ := s.queries.RainToday(r.Context())

	writeJSON(w, map[string]any{
		"observation":    cond,
		"rain_last_hour": rainLastHour,
		"rain_today":     rainToday,
	})
}

func (s *Server) handleTemperature(w http.ResponseWriter, r *http.Request) {
	since, interval := parseTimeRange(r)
	data, err := s.queries.TempHumidity(r.Context(), since, interval)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, data)
}

func (s *Server) handleWind(w http.ResponseWriter, r *http.Request) {
	since, interval := parseTimeRange(r)
	data, err := s.queries.Wind(r.Context(), since, interval)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, data)
}

func (s *Server) handleRain(w http.ResponseWriter, r *http.Request) {
	since, interval := parseTimeRange(r)
	data, err := s.queries.Rain(r.Context(), since, interval)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, data)
}

func (s *Server) handlePressure(w http.ResponseWriter, r *http.Request) {
	since, interval := parseTimeRange(r)
	data, err := s.queries.Pressure(r.Context(), since, interval)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, data)
}

func (s *Server) handleSolar(w http.ResponseWriter, r *http.Request) {
	since, interval := parseTimeRange(r)
	data, err := s.queries.Solar(r.Context(), since, interval)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, data)
}

func (s *Server) handleHumidity(w http.ResponseWriter, r *http.Request) {
	since, interval := parseTimeRange(r)
	data, err := s.queries.Humidity(r.Context(), since, interval)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, data)
}

func (s *Server) handleUV(w http.ResponseWriter, r *http.Request) {
	since, interval := parseTimeRange(r)
	data, err := s.queries.UV(r.Context(), since, interval)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, data)
}

// parseTimeRange reads ?range=24h (default) and returns a since time and
// an appropriate bucket interval.
func parseTimeRange(r *http.Request) (time.Time, string) {
	rangeStr := r.URL.Query().Get("range")
	if rangeStr == "" {
		rangeStr = "24h"
	}

	d, err := time.ParseDuration(rangeStr)
	if err != nil {
		d = 24 * time.Hour
	}

	since := time.Now().Add(-d)

	// Pick a bucket interval that gives ~100-200 points
	var interval string
	switch {
	case d <= 6*time.Hour:
		interval = "5 minutes"
	case d <= 24*time.Hour:
		interval = "15 minutes"
	case d <= 7*24*time.Hour:
		interval = "1 hour"
	case d <= 30*24*time.Hour:
		interval = "6 hours"
	default:
		interval = "1 day"
	}

	return since, interval
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
