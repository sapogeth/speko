// Package testutil provides a mock Speko API server for unit tests.
package testutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
)

// Server is a mock Speko API server.
type Server struct {
	*httptest.Server
	Requests []*http.Request
}

// Handler is a function that handles a mock API request.
type Handler func(w http.ResponseWriter, r *http.Request)

// NewServer starts a mock server with the given route handlers.
// Routes is a map of "METHOD /path" → handler.
func NewServer(routes map[string]Handler) *Server {
	s := &Server{}
	mux := http.NewServeMux()

	for route, h := range routes {
		h := h
		mux.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
			s.Requests = append(s.Requests, r)
			h(w, r)
		})
	}

	s.Server = httptest.NewServer(mux)
	return s
}

// JSON writes a JSON response.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

// JSONWithRouting writes a JSON response with Speko routing headers.
func JSONWithRouting(w http.ResponseWriter, status int, v any, provider, model string) {
	w.Header().Set("X-Speko-Provider", provider)
	w.Header().Set("X-Speko-Model", model)
	JSON(w, status, v)
}

// SSE streams server-sent events.
func SSE(w http.ResponseWriter, events []string) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	for _, e := range events {
		w.Write([]byte(e + "\n\n")) //nolint:errcheck
	}
}

// Error writes an API error response.
func Error(w http.ResponseWriter, status int, code, message string) {
	JSON(w, status, map[string]any{
		"code":    code,
		"message": message,
	})
}
