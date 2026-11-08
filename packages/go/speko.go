// Package spekoai is the official Go SDK for the Speko Voice AI platform.
// It provides typed clients for routing-aware STT, TTS, LLM completion,
// session management, telephony, and agent operations.
//
// Usage:
//
//	client := spekoai.New(os.Getenv("SPEKO_API_KEY"))
//
//	result, err := client.Transcribe(ctx, spekoai.TranscribeRequest{
//	    Audio:    audioBytes,
//	    Language: "es-MX",
//	    Optimize: spekoai.OptimizeAccuracy,
//	})
package spekoai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultBaseURL = "https://api.speko.dev"
	defaultTimeout = 30 * time.Second
	sdkVersion     = "0.1.0"
)

// Client is the Speko API client. Create one with New().
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// Option configures a Client.
type Option func(*Client)

// WithBaseURL overrides the API base URL (useful for testing).
func WithBaseURL(url string) Option {
	return func(c *Client) { c.baseURL = url }
}

// WithHTTPClient replaces the underlying HTTP client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpClient = hc }
}

// WithTimeout sets the HTTP request timeout.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) { c.httpClient.Timeout = d }
}

// New creates a new Speko client. The apiKey must be kept server-side;
// never embed it in browser bundles or mobile app binaries.
func New(apiKey string, opts ...Option) *Client {
	c := &Client{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// do sends an authenticated JSON request and decodes the response into dst.
// Pass dst=nil to discard the body (e.g. DELETE).
func (c *Client) do(ctx context.Context, method, path string, body, dst any) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("spekoai: marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("spekoai: build request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("User-Agent", "spekoai-go/"+sdkVersion)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("spekoai: http: %w", err)
	}

	if res.StatusCode >= 400 {
		defer res.Body.Close()
		return res, parseAPIError(res)
	}

	if dst != nil {
		defer res.Body.Close()
		if err := json.NewDecoder(res.Body).Decode(dst); err != nil {
			return res, fmt.Errorf("spekoai: decode response: %w", err)
		}
	}

	return res, nil
}

// doRaw sends a request and returns the raw response body. Caller must close it.
func (c *Client) doRaw(ctx context.Context, method, path string, body any) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("spekoai: marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("spekoai: build request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("User-Agent", "spekoai-go/"+sdkVersion)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("spekoai: http: %w", err)
	}

	if res.StatusCode >= 400 {
		defer res.Body.Close()
		return res, parseAPIError(res)
	}

	return res, nil
}
