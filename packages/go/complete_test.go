package spekoai_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	spekoai "github.com/ilaszajsenbaev/spekoai-go"
	"github.com/ilaszajsenbaev/spekoai-go/internal/testutil"
)

func TestComplete_Streaming(t *testing.T) {
	events := []string{
		`event: delta` + "\n" + `data: {"delta":"Hello"}`,
		`event: delta` + "\n" + `data: {"delta":", world"}`,
		`event: done` + "\n" + `data: {}`,
	}

	srv := testutil.NewServer(map[string]testutil.Handler{
		"POST /v1/complete": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Speko-Provider", "openai")
			w.Header().Set("X-Speko-Model", "gpt-4o-mini")
			testutil.SSE(w, events)
		},
	})
	defer srv.Close()

	client := spekoai.New("test-key", spekoai.WithBaseURL(srv.URL))
	result, err := client.Complete(context.Background(), spekoai.CompleteRequest{
		Messages: []spekoai.Message{
			{Role: "user", Content: "say hello"},
		},
		Language: "en-US",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Content != "Hello, world" {
		t.Errorf("content = %q, want %q", result.Content, "Hello, world")
	}
	if result.Routing == nil || result.Routing.Provider != "openai" {
		t.Errorf("routing provider = %v, want openai", result.Routing)
	}
}

func TestComplete_RateLimit(t *testing.T) {
	srv := testutil.NewServer(map[string]testutil.Handler{
		"POST /v1/complete": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Retry-After", "30")
			testutil.Error(w, http.StatusTooManyRequests, "RATE_LIMITED", "slow down")
		},
	})
	defer srv.Close()

	client := spekoai.New("test-key", spekoai.WithBaseURL(srv.URL))
	_, err := client.Complete(context.Background(), spekoai.CompleteRequest{
		Messages: []spekoai.Message{{Role: "user", Content: "hi"}},
		Language: "en-US",
	})

	var rlErr *spekoai.RateLimitError
	if !errors.As(err, &rlErr) {
		t.Fatalf("expected *RateLimitError, got %T: %v", err, err)
	}
	if rlErr.RetryAfter.Seconds() != 30 {
		t.Errorf("RetryAfter = %v, want 30s", rlErr.RetryAfter)
	}
}

func TestStreamComplete_CancelMidStream(t *testing.T) {
	done := make(chan struct{})
	srv := testutil.NewServer(map[string]testutil.Handler{
		"POST /v1/complete": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(http.StatusOK)
			flusher := w.(http.Flusher)

			for i := 0; i < 100; i++ {
				select {
				case <-r.Context().Done():
					close(done)
					return
				default:
					w.Write([]byte("event: delta\ndata: {\"delta\":\"x\"}\n\n")) //nolint:errcheck
					flusher.Flush()
				}
			}
		},
	})
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := spekoai.New("test-key", spekoai.WithBaseURL(srv.URL))

	ch, err := client.StreamComplete(ctx, spekoai.CompleteRequest{
		Messages: []spekoai.Message{{Role: "user", Content: "stream forever"}},
		Language: "en-US",
	})
	if err != nil {
		t.Fatalf("unexpected error starting stream: %v", err)
	}

	// Consume a few events then cancel
	count := 0
	for range ch {
		count++
		if count == 3 {
			cancel()
			break
		}
	}

	// Drain remaining events after cancel
	for range ch {
	}

	if count < 3 {
		t.Errorf("received %d events before cancel, want at least 3", count)
	}
}
