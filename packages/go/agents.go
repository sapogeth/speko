package spekoai

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Agent is a saved voice persona with a system prompt and routing config.
type Agent struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Prompt      string         `json:"systemPrompt"`
	Intent      *RoutingIntent `json:"intent,omitempty"`
	Constraints *ProviderConstraints `json:"constraints,omitempty"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
}

// CreateAgentRequest defines a new voice agent.
type CreateAgentRequest struct {
	Name        string               `json:"name"`
	Prompt      string               `json:"systemPrompt"`
	Intent      *RoutingIntent       `json:"intent,omitempty"`
	Constraints *ProviderConstraints `json:"constraints,omitempty"`
}

// UpdateAgentRequest partially updates an agent. Zero-value fields are ignored.
type UpdateAgentRequest struct {
	Name        string               `json:"name,omitempty"`
	Prompt      string               `json:"systemPrompt,omitempty"`
	Intent      *RoutingIntent       `json:"intent,omitempty"`
	Constraints *ProviderConstraints `json:"constraints,omitempty"`
}

// ListAgents returns all agents in the workspace.
func (c *Client) ListAgents(ctx context.Context) ([]Agent, error) {
	var out struct {
		Agents []Agent `json:"agents"`
	}
	_, err := c.do(ctx, http.MethodGet, "/v1/agents", nil, &out)
	return out.Agents, err
}

// CreateAgent creates a new voice agent.
func (c *Client) CreateAgent(ctx context.Context, req CreateAgentRequest) (*Agent, error) {
	var agent Agent
	_, err := c.do(ctx, http.MethodPost, "/v1/agents", req, &agent)
	return &agent, err
}

// GetAgent fetches a single agent by ID.
func (c *Client) GetAgent(ctx context.Context, id string) (*Agent, error) {
	var agent Agent
	_, err := c.do(ctx, http.MethodGet, "/v1/agents/"+id, nil, &agent)
	return &agent, err
}

// UpdateAgent partially updates an agent.
func (c *Client) UpdateAgent(ctx context.Context, id string, req UpdateAgentRequest) (*Agent, error) {
	var agent Agent
	_, err := c.do(ctx, http.MethodPatch, "/v1/agents/"+id, req, &agent)
	return &agent, err
}

// DeleteAgent removes an agent permanently.
func (c *Client) DeleteAgent(ctx context.Context, id string) error {
	res, err := c.do(ctx, http.MethodDelete, "/v1/agents/"+id, nil, nil)
	if err != nil {
		return err
	}
	res.Body.Close()
	return nil
}

// PhoneDialRequest places an outbound call with an agent.
type PhoneDialRequest struct {
	// To is the destination phone number in E.164 format.
	To string `json:"to"`
	// AgentID selects the agent that will handle the call.
	AgentID string `json:"agentId,omitempty"`
	// From overrides the caller ID. Leave empty to use the workspace default.
	From string `json:"from,omitempty"`
	// RingTimeout is the maximum seconds to wait for the call to be answered.
	RingTimeout int `json:"ringTimeout,omitempty"`
	// Variables are injected into the agent prompt for this call.
	Variables map[string]string `json:"variables,omitempty"`
	// IdempotencyKey prevents duplicate dials within 24 hours.
	IdempotencyKey string `json:"-"`
}

// Call is a telephony session.
type Call struct {
	ID        string     `json:"id"`
	AgentID   string     `json:"agentId"`
	To        string     `json:"to"`
	Status    string     `json:"status"`
	StartedAt *time.Time `json:"startedAt,omitempty"`
}

// DialPhone places an outbound call. Provide IdempotencyKey to safely retry.
func (c *Client) DialPhone(ctx context.Context, req PhoneDialRequest) (*Call, error) {
	httpReq, err := buildJSONRequest(ctx, http.MethodPost, c.baseURL+"/v1/sessions/phone", req, c.apiKey)
	if err != nil {
		return nil, err
	}
	if req.IdempotencyKey != "" {
		httpReq.Header.Set("Idempotency-Key", req.IdempotencyKey)
	}

	res, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("spekoai: http: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		return nil, parseAPIError(res)
	}

	var call Call
	if err := decodeJSON(res, &call); err != nil {
		return nil, err
	}
	return &call, nil
}
