package spekoai

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// APIError is returned for 4xx/5xx responses from the Speko API.
type APIError struct {
	StatusCode int
	Code       string `json:"code"`
	Message    string `json:"message"`
	Detail     any    `json:"detail,omitempty"`
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("spekoai: %s (%s, status %d)", e.Message, e.Code, e.StatusCode)
	}
	return fmt.Sprintf("spekoai: HTTP %d", e.StatusCode)
}

// AuthError is returned for 401 responses.
type AuthError struct{ *APIError }

func (e *AuthError) Error() string {
	return fmt.Sprintf("spekoai: authentication failed — check your SPEKO_API_KEY (%s)", e.APIError.Message)
}

// RateLimitError is returned for 429 responses.
type RateLimitError struct {
	*APIError
	RetryAfter time.Duration
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("spekoai: rate limited — retry after %s", e.RetryAfter)
}

// BillingError is returned for 402 responses (insufficient credits / spending cap).
type BillingError struct{ *APIError }

func (e *BillingError) Error() string {
	return fmt.Sprintf("spekoai: billing refused — %s", e.APIError.Message)
}

// NoProviderError is returned for 422 responses (no capable provider for the routing intent).
type NoProviderError struct{ *APIError }

func (e *NoProviderError) Error() string {
	return fmt.Sprintf("spekoai: no provider available for this routing intent — %s", e.APIError.Message)
}

func parseAPIError(res *http.Response) error {
	base := &APIError{StatusCode: res.StatusCode}

	var payload struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Detail  any    `json:"detail"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err == nil {
		base.Code = payload.Code
		base.Message = payload.Message
		base.Detail = payload.Detail
	}

	switch res.StatusCode {
	case http.StatusUnauthorized:
		return &AuthError{base}
	case http.StatusPaymentRequired:
		return &BillingError{base}
	case http.StatusTooManyRequests:
		ra := parseDuration(res.Header.Get("Retry-After"))
		return &RateLimitError{APIError: base, RetryAfter: ra}
	case http.StatusUnprocessableEntity:
		return &NoProviderError{base}
	default:
		return base
	}
}

func parseDuration(s string) time.Duration {
	var secs float64
	if _, err := fmt.Sscanf(s, "%f", &secs); err == nil {
		return time.Duration(secs * float64(time.Second))
	}
	return 0
}
