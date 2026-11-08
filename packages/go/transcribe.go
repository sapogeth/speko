package spekoai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
)

// TranscribeRequest is the input for a batch speech-to-text call.
type TranscribeRequest struct {
	// Audio is the raw audio bytes. Supported formats: webm, wav, mp3, ogg, flac.
	Audio []byte
	// Filename hints the content type (e.g. "audio.webm"). Optional.
	Filename string
	// Language is a BCP-47 tag. Required.
	Language string
	// Region pins the routing decision to a cloud region. Optional.
	Region Region
	// Optimize selects the scoring axis. Defaults to balanced.
	Optimize OptimizeFor
	// Constraints restricts which STT providers may be selected.
	Constraints *ProviderConstraints
}

// TranscribeResult is the response from a batch transcription.
type TranscribeResult struct {
	// Text is the transcribed speech.
	Text string
	// Routing describes which provider/model was selected.
	Routing *RoutingMetadata
}

// Transcribe sends audio to the Speko STT router and returns the transcript.
// The router selects the best-scoring provider for the given language using
// live benchmark data from benchmarks.speko.ai.
func (c *Client) Transcribe(ctx context.Context, req TranscribeRequest) (*TranscribeResult, error) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	filename := req.Filename
	if filename == "" {
		filename = "audio.webm"
	}
	fw, err := mw.CreateFormFile("audio", filename)
	if err != nil {
		return nil, fmt.Errorf("spekoai: create form file: %w", err)
	}
	if _, err := fw.Write(req.Audio); err != nil {
		return nil, fmt.Errorf("spekoai: write audio: %w", err)
	}

	addFormField(mw, "language", req.Language)
	if req.Region != "" {
		addFormField(mw, "region", string(req.Region))
	}
	if req.Optimize != "" {
		addFormField(mw, "optimizeFor", string(req.Optimize))
	}
	if req.Constraints != nil {
		if b, err := json.Marshal(req.Constraints); err == nil {
			addFormField(mw, "constraints", string(b))
		}
	}
	mw.Close()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/transcribe", &buf)
	if err != nil {
		return nil, fmt.Errorf("spekoai: build request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", mw.FormDataContentType())
	httpReq.Header.Set("User-Agent", "spekoai-go/"+sdkVersion)

	res, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("spekoai: http: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return nil, parseAPIError(res)
	}

	var out struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("spekoai: decode response: %w", err)
	}

	return &TranscribeResult{
		Text:    out.Text,
		Routing: extractRouting(res.Header),
	}, nil
}

// TranscriptEvent is a streaming event from the STT router.
type TranscriptEvent struct {
	// Text is the current transcript fragment.
	Text string
	// IsFinal is true when this segment is committed and won't be revised.
	IsFinal bool
	// Routing is populated on the first event with provider/model info.
	Routing *RoutingMetadata
	// Err is set if the stream encountered an error.
	Err error
}

// StreamTranscribeRequest is the input for a streaming speech-to-text call.
type StreamTranscribeRequest struct {
	Audio    io.Reader
	Filename string
	Language string
	Region   Region
	Optimize OptimizeFor
}

// StreamTranscribe sends audio to the Speko STT router and returns a channel
// of incremental transcript events. The channel is closed when the stream ends.
// Cancel ctx to abort early.
func (c *Client) StreamTranscribe(ctx context.Context, req StreamTranscribeRequest) (<-chan TranscriptEvent, error) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	filename := req.Filename
	if filename == "" {
		filename = "audio.webm"
	}
	fw, err := mw.CreateFormFile("audio", filename)
	if err != nil {
		return nil, fmt.Errorf("spekoai: create form file: %w", err)
	}
	if _, err := io.Copy(fw, req.Audio); err != nil {
		return nil, fmt.Errorf("spekoai: read audio: %w", err)
	}

	addFormField(mw, "language", req.Language)
	addFormField(mw, "stream", "true")
	if req.Region != "" {
		addFormField(mw, "region", string(req.Region))
	}
	if req.Optimize != "" {
		addFormField(mw, "optimizeFor", string(req.Optimize))
	}
	mw.Close()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/transcribe", &buf)
	if err != nil {
		return nil, fmt.Errorf("spekoai: build request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", mw.FormDataContentType())
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("User-Agent", "spekoai-go/"+sdkVersion)

	res, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("spekoai: http: %w", err)
	}
	if res.StatusCode >= 400 {
		defer res.Body.Close()
		return nil, parseAPIError(res)
	}

	ch := make(chan TranscriptEvent, 32)
	routing := extractRouting(res.Header)

	go func() {
		defer res.Body.Close()
		defer close(ch)

		scanner := bufio.NewScanner(res.Body)
		var eventType, dataLine string

		for scanner.Scan() {
			line := scanner.Text()
			switch {
			case strings.HasPrefix(line, "event:"):
				eventType = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			case strings.HasPrefix(line, "data:"):
				dataLine = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			case line == "":
				if dataLine == "" {
					continue
				}
				evt := parseTranscriptEvent(eventType, dataLine, routing)
				routing = nil // only send on first event
				select {
				case ch <- evt:
				case <-ctx.Done():
					return
				}
				eventType, dataLine = "", ""
			}
		}

		if err := scanner.Err(); err != nil && ctx.Err() == nil {
			select {
			case ch <- TranscriptEvent{Err: fmt.Errorf("spekoai: stream read: %w", err)}:
			default:
			}
		}
	}()

	return ch, nil
}

func parseTranscriptEvent(eventType, data string, routing *RoutingMetadata) TranscriptEvent {
	switch eventType {
	case "error":
		var e struct{ Message string }
		json.Unmarshal([]byte(data), &e) //nolint:errcheck
		return TranscriptEvent{Err: fmt.Errorf("spekoai: stream error: %s", e.Message)}
	case "done":
		return TranscriptEvent{IsFinal: true, Routing: routing}
	default:
		var payload struct {
			Text    string `json:"text"`
			IsFinal bool   `json:"isFinal"`
		}
		json.Unmarshal([]byte(data), &payload) //nolint:errcheck
		return TranscriptEvent{
			Text:    payload.Text,
			IsFinal: payload.IsFinal,
			Routing: routing,
		}
	}
}

func addFormField(mw *multipart.Writer, key, value string) {
	if value == "" {
		return
	}
	fw, _ := mw.CreateFormField(key)
	io.WriteString(fw, value) //nolint:errcheck
}
