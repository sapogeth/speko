package spekoai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// CompleteRequest is the input for a single-turn LLM completion.
type CompleteRequest struct {
	// Messages is the conversation history. Required.
	Messages []Message
	// Language is a BCP-47 tag used for LLM routing. Required.
	Language string
	// Temperature controls randomness (0.0–2.0). Defaults to 0.7.
	Temperature float64
	// MaxTokens limits the response length.
	MaxTokens int
	// Tools are function definitions the model may invoke.
	Tools []Tool
	// Optimize selects the routing axis.
	Optimize OptimizeFor
	// Region pins routing to a cloud region.
	Region Region
}

// CompleteResult is the response from a non-streaming LLM completion.
type CompleteResult struct {
	// Content is the model's reply.
	Content string
	// ToolCalls contains any function calls the model requested.
	ToolCalls []ToolCall
	// Routing describes which LLM provider/model was selected.
	Routing *RoutingMetadata
}

// ToolCall represents a function the model requested to call.
type ToolCall struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

// Complete sends messages to the Speko LLM router and returns the full response.
func (c *Client) Complete(ctx context.Context, req CompleteRequest) (*CompleteResult, error) {
	chunks, routing, err := c.streamComplete(ctx, req)
	if err != nil {
		return nil, err
	}

	var sb strings.Builder
	var toolCalls []ToolCall
	for evt := range chunks {
		if evt.Err != nil {
			return nil, evt.Err
		}
		sb.WriteString(evt.Delta)
		toolCalls = append(toolCalls, evt.ToolCalls...)
	}

	return &CompleteResult{
		Content:   sb.String(),
		ToolCalls: toolCalls,
		Routing:   routing,
	}, nil
}

// CompletionEvent is a streaming delta from the LLM router.
type CompletionEvent struct {
	// Delta is an incremental text fragment.
	Delta string
	// ToolCalls are function calls emitted in this chunk.
	ToolCalls []ToolCall
	// Done is true on the final event.
	Done bool
	// Routing is set on the first event.
	Routing *RoutingMetadata
	// Err signals a stream error.
	Err error
}

// StreamComplete streams LLM completion deltas. The returned channel is closed
// when the stream ends or ctx is cancelled.
func (c *Client) StreamComplete(ctx context.Context, req CompleteRequest) (<-chan CompletionEvent, error) {
	ch, _, err := c.streamComplete(ctx, req)
	return ch, err
}

func (c *Client) streamComplete(ctx context.Context, req CompleteRequest) (<-chan CompletionEvent, *RoutingMetadata, error) {
	if req.Temperature == 0 {
		req.Temperature = 0.7
	}

	payload := map[string]any{
		"messages":    req.Messages,
		"language":    req.Language,
		"temperature": req.Temperature,
	}
	if req.MaxTokens > 0 {
		payload["maxTokens"] = req.MaxTokens
	}
	if len(req.Tools) > 0 {
		payload["tools"] = req.Tools
	}
	if req.Optimize != "" {
		payload["optimizeFor"] = string(req.Optimize)
	}
	if req.Region != "" {
		payload["region"] = string(req.Region)
	}

	b, _ := json.Marshal(payload)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/complete", bytes.NewReader(b))
	if err != nil {
		return nil, nil, fmt.Errorf("spekoai: build request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("User-Agent", "spekoai-go/"+sdkVersion)

	res, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, nil, fmt.Errorf("spekoai: http: %w", err)
	}
	if res.StatusCode >= 400 {
		defer res.Body.Close()
		return nil, nil, parseAPIError(res)
	}

	routing := extractRouting(res.Header)
	ch := make(chan CompletionEvent, 64)

	go func() {
		defer res.Body.Close()
		defer close(ch)

		scanner := bufio.NewScanner(res.Body)
		var eventType, dataLine string
		firstRouting := routing

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
				evt := parseCompletionEvent(eventType, dataLine, firstRouting)
				firstRouting = nil
				select {
				case ch <- evt:
				case <-ctx.Done():
					return
				}
				eventType, dataLine = "", ""
			}
		}
	}()

	return ch, routing, nil
}

func parseCompletionEvent(eventType, data string, routing *RoutingMetadata) CompletionEvent {
	switch eventType {
	case "done":
		return CompletionEvent{Done: true}
	case "error":
		var e struct{ Message string }
		json.Unmarshal([]byte(data), &e) //nolint:errcheck
		return CompletionEvent{Err: fmt.Errorf("spekoai: completion stream error: %s", e.Message)}
	default:
		var payload struct {
			Delta     string     `json:"delta"`
			ToolCalls []ToolCall `json:"toolCalls,omitempty"`
		}
		json.Unmarshal([]byte(data), &payload) //nolint:errcheck
		return CompletionEvent{
			Delta:     payload.Delta,
			ToolCalls: payload.ToolCalls,
			Routing:   routing,
		}
	}
}
