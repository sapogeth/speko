package spekoai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func buildJSONRequest(ctx context.Context, method, url string, body any, apiKey string) (*http.Request, error) {
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("spekoai: marshal: %w", err)
		}
		r = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, r)
	if err != nil {
		return nil, fmt.Errorf("spekoai: new request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "spekoai-go/"+sdkVersion)
	return req, nil
}

func decodeJSON(res *http.Response, dst any) error {
	if err := json.NewDecoder(res.Body).Decode(dst); err != nil {
		return fmt.Errorf("spekoai: decode response: %w", err)
	}
	return nil
}
