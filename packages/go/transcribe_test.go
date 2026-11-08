package spekoai_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	spekoai "github.com/ilaszajsenbaev/spekoai-go"
	"github.com/ilaszajsenbaev/spekoai-go/internal/testutil"
)

func TestTranscribe_Success(t *testing.T) {
	srv := testutil.NewServer(map[string]testutil.Handler{
		"POST /v1/transcribe": func(w http.ResponseWriter, r *http.Request) {
			testutil.JSONWithRouting(w, http.StatusOK,
				map[string]any{"text": "hola mundo"},
				"deepgram", "nova-3",
			)
		},
	})
	defer srv.Close()

	client := spekoai.New("test-key", spekoai.WithBaseURL(srv.URL))
	result, err := client.Transcribe(context.Background(), spekoai.TranscribeRequest{
		Audio:    []byte("fake-audio"),
		Language: "es-MX",
		Optimize: spekoai.OptimizeAccuracy,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Text != "hola mundo" {
		t.Errorf("text = %q, want %q", result.Text, "hola mundo")
	}
	if result.Routing == nil {
		t.Fatal("routing metadata is nil")
	}
	if result.Routing.Provider != "deepgram" {
		t.Errorf("provider = %q, want %q", result.Routing.Provider, "deepgram")
	}
	if result.Routing.Model != "nova-3" {
		t.Errorf("model = %q, want %q", result.Routing.Model, "nova-3")
	}
}

func TestTranscribe_AuthError(t *testing.T) {
	srv := testutil.NewServer(map[string]testutil.Handler{
		"POST /v1/transcribe": func(w http.ResponseWriter, r *http.Request) {
			testutil.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid api key")
		},
	})
	defer srv.Close()

	client := spekoai.New("bad-key", spekoai.WithBaseURL(srv.URL))
	_, err := client.Transcribe(context.Background(), spekoai.TranscribeRequest{
		Audio:    []byte("audio"),
		Language: "en-US",
	})

	var authErr *spekoai.AuthError
	if !errors.As(err, &authErr) {
		t.Errorf("expected *AuthError, got %T: %v", err, err)
	}
}

func TestTranscribe_NoProvider(t *testing.T) {
	srv := testutil.NewServer(map[string]testutil.Handler{
		"POST /v1/transcribe": func(w http.ResponseWriter, r *http.Request) {
			testutil.Error(w, http.StatusUnprocessableEntity,
				"NO_CAPABLE_PROVIDER", "no provider supports swahili in this region")
		},
	})
	defer srv.Close()

	client := spekoai.New("test-key", spekoai.WithBaseURL(srv.URL))
	_, err := client.Transcribe(context.Background(), spekoai.TranscribeRequest{
		Audio:    []byte("audio"),
		Language: "sw-KE",
		Region:   spekoai.RegionEuropeWest,
	})

	var npErr *spekoai.NoProviderError
	if !errors.As(err, &npErr) {
		t.Errorf("expected *NoProviderError, got %T: %v", err, err)
	}
}

func TestTranscribe_ContextCancellation(t *testing.T) {
	srv := testutil.NewServer(map[string]testutil.Handler{
		"POST /v1/transcribe": func(w http.ResponseWriter, r *http.Request) {
			// Block until client cancels
			<-r.Context().Done()
		},
	})
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	client := spekoai.New("test-key", spekoai.WithBaseURL(srv.URL))
	_, err := client.Transcribe(ctx, spekoai.TranscribeRequest{
		Audio:    []byte("audio"),
		Language: "en-US",
	})

	if err == nil {
		t.Error("expected error after context cancellation, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Logf("error (non-fatal): %v", err) // may be wrapped
	}
}

func TestTranscribe_AuthorizationHeader(t *testing.T) {
	var gotAuth string
	srv := testutil.NewServer(map[string]testutil.Handler{
		"POST /v1/transcribe": func(w http.ResponseWriter, r *http.Request) {
			gotAuth = r.Header.Get("Authorization")
			testutil.JSON(w, http.StatusOK, map[string]any{"text": "test"})
		},
	})
	defer srv.Close()

	client := spekoai.New("sk-test-123", spekoai.WithBaseURL(srv.URL))
	client.Transcribe(context.Background(), spekoai.TranscribeRequest{ //nolint:errcheck
		Audio:    []byte("audio"),
		Language: "en-US",
	})

	if gotAuth != "Bearer sk-test-123" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer sk-test-123")
	}
}
