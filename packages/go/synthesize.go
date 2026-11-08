package spekoai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// AudioFormat is the output encoding for synthesized speech.
type AudioFormat string

const (
	AudioFormatWAV AudioFormat = "wav"
	AudioFormatPCM AudioFormat = "pcm"
	AudioFormatMP3 AudioFormat = "mp3"
)

// SynthesizeRequest is the input for a text-to-speech call.
type SynthesizeRequest struct {
	// Text is the content to synthesize. Required.
	Text string
	// Language is a BCP-47 tag. Required.
	Language string
	// Voice is a provider-specific voice ID. Optional; router selects if omitted.
	Voice string
	// Speed multiplies playback rate (0.5–2.0). Defaults to 1.0.
	Speed float64
	// Format selects the output encoding. Defaults to WAV.
	Format AudioFormat
	// Optimize selects the routing axis.
	Optimize OptimizeFor
	// Region pins routing to a cloud region.
	Region Region
}

// SynthesizeResult is the response from a TTS call.
type SynthesizeResult struct {
	// Audio contains the raw audio bytes in the requested format.
	Audio []byte
	// Routing describes which provider/model was selected.
	Routing *RoutingMetadata
}

// Synthesize converts text to speech using the best-scoring TTS provider
// for the given language, as ranked by Speko's live benchmarks.
func (c *Client) Synthesize(ctx context.Context, req SynthesizeRequest) (*SynthesizeResult, error) {
	if req.Speed == 0 {
		req.Speed = 1.0
	}
	if req.Format == "" {
		req.Format = AudioFormatWAV
	}

	payload := map[string]any{
		"text":         req.Text,
		"language":     req.Language,
		"speed":        req.Speed,
		"outputFormat": string(req.Format),
	}
	if req.Voice != "" {
		payload["voice"] = req.Voice
	}
	if req.Optimize != "" {
		payload["optimizeFor"] = string(req.Optimize)
	}
	if req.Region != "" {
		payload["region"] = string(req.Region)
	}

	b, _ := json.Marshal(payload)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/synthesize", bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("spekoai: build request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "spekoai-go/"+sdkVersion)

	res, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("spekoai: http: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return nil, parseAPIError(res)
	}

	audio, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("spekoai: read audio: %w", err)
	}

	return &SynthesizeResult{
		Audio:   audio,
		Routing: extractRouting(res.Header),
	}, nil
}

// SpeechEvent is a streaming TTS audio chunk.
type SpeechEvent struct {
	// Chunk is a fragment of audio data.
	Chunk []byte
	// Done is true on the final event (Chunk will be empty).
	Done bool
	// Routing is populated on the first event.
	Routing *RoutingMetadata
	// Err is set on stream errors.
	Err error
}

// StreamSynthesize streams audio chunks from the TTS router as they are generated.
// This reduces time-to-first-audio compared to waiting for the full Synthesize call.
func (c *Client) StreamSynthesize(ctx context.Context, req SynthesizeRequest) (<-chan SpeechEvent, error) {
	if req.Speed == 0 {
		req.Speed = 1.0
	}
	if req.Format == "" {
		req.Format = AudioFormatPCM // PCM preferred for streaming
	}

	payload := map[string]any{
		"text":         req.Text,
		"language":     req.Language,
		"speed":        req.Speed,
		"outputFormat": string(req.Format),
		"delivery":     "stream",
	}
	if req.Voice != "" {
		payload["voice"] = req.Voice
	}
	if req.Optimize != "" {
		payload["optimizeFor"] = string(req.Optimize)
	}

	b, _ := json.Marshal(payload)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/synthesize", bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("spekoai: build request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
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

	ch := make(chan SpeechEvent, 32)
	routing := extractRouting(res.Header)

	go func() {
		defer res.Body.Close()
		defer close(ch)

		scanner := bufio.NewScanner(res.Body)
		scanner.Buffer(make([]byte, 1<<20), 1<<20) // 1 MB max line
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
				evt := parseSpeechEvent(eventType, dataLine, routing)
				routing = nil
				select {
				case ch <- evt:
				case <-ctx.Done():
					return
				}
				eventType, dataLine = "", ""
			}
		}
	}()

	return ch, nil
}

func parseSpeechEvent(eventType, data string, routing *RoutingMetadata) SpeechEvent {
	switch eventType {
	case "done":
		return SpeechEvent{Done: true}
	case "error":
		var e struct{ Message string }
		json.Unmarshal([]byte(data), &e) //nolint:errcheck
		return SpeechEvent{Err: fmt.Errorf("spekoai: tts stream error: %s", e.Message)}
	default:
		// data is base64-encoded audio chunk
		var payload struct {
			Audio []byte `json:"audio"` // already decoded by json.Unmarshal
		}
		json.Unmarshal([]byte(data), &payload) //nolint:errcheck
		return SpeechEvent{Chunk: payload.Audio, Routing: routing}
	}
}
