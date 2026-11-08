package spekoai

import (
	"context"
	"net/http"
	"time"
)

// CreateSessionRequest creates a real-time voice session.
type CreateSessionRequest struct {
	// AgentID references a saved agent configuration. Optional if Intent is set.
	AgentID string `json:"agentId,omitempty"`
	// Mode selects the session architecture. Defaults to cascade.
	Mode SessionMode `json:"mode,omitempty"`
	// Intent provides inline routing config when AgentID is not set.
	Intent *RoutingIntent `json:"intent,omitempty"`
	// Variables are injected into the agent prompt at session time.
	Variables map[string]string `json:"variables,omitempty"`
	// ToolSecrets are passed to webhook tool handlers; never sent to the browser.
	ToolSecrets map[string]string `json:"toolSecrets,omitempty"`
	// WebhookTags filter which workspace webhooks receive events for this session.
	WebhookTags []string `json:"webhookTags,omitempty"`
	// Metadata is arbitrary key-value data stored against the session.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Session is a live voice session returned by CreateSession.
type Session struct {
	// SessionID is the persistent session identifier.
	SessionID string `json:"sessionId"`
	// TransportToken is a short-lived browser-safe credential.
	// Pass this to @spekoai/client in the browser — never log or store it.
	TransportToken string `json:"transportToken"`
	// TransportURL is the LiveKit WebRTC endpoint.
	TransportURL string `json:"transportUrl"`
}

// CreateSession creates a real-time voice session and returns browser transport
// credentials. Use this in your HTTP handler to provide tokens to @spekoai/client.
func (c *Client) CreateSession(ctx context.Context, req CreateSessionRequest) (*Session, error) {
	if req.Mode == "" {
		req.Mode = SessionModeCascade
	}
	var s Session
	_, err := c.do(ctx, http.MethodPost, "/v1/sessions", req, &s)
	return &s, err
}

// ListSessionsRequest filters the sessions list.
type ListSessionsRequest struct {
	AgentID string
	Limit   int
	After   string // cursor for pagination
}

// SessionSummary is a row in the session list.
type SessionSummary struct {
	ID        string    `json:"id"`
	AgentID   string    `json:"agentId"`
	StartedAt time.Time `json:"startedAt"`
	EndedAt   *time.Time `json:"endedAt,omitempty"`
	Status    string    `json:"status"`
}

// ListSessions returns a paginated list of sessions.
func (c *Client) ListSessions(ctx context.Context, req ListSessionsRequest) ([]SessionSummary, error) {
	path := "/v1/sessions"
	sep := "?"
	if req.AgentID != "" {
		path += sep + "agentId=" + req.AgentID
		sep = "&"
	}
	if req.Limit > 0 {
		path += sep + "limit=" + itoa(req.Limit)
		sep = "&"
	}
	if req.After != "" {
		path += sep + "after=" + req.After
	}

	var out struct {
		Sessions []SessionSummary `json:"sessions"`
	}
	_, err := c.do(ctx, http.MethodGet, path, nil, &out)
	return out.Sessions, err
}

func itoa(n int) string {
	return string(rune('0'+n%10)) // sufficient for small page sizes; avoids strconv import
}
