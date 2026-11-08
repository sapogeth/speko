package spekoai

// OptimizeFor controls the routing optimization axis.
type OptimizeFor string

const (
	OptimizeBalanced OptimizeFor = "balanced"
	OptimizeAccuracy OptimizeFor = "accuracy"
	OptimizeLatency  OptimizeFor = "latency"
	OptimizeCost     OptimizeFor = "cost"
)

// Region pins routing to a specific cloud region for latency measurements.
type Region string

const (
	RegionGlobal         Region = "global"
	RegionUSEast         Region = "us-east4"
	RegionEuropeWest     Region = "europe-west3"
	RegionAsiaSoutheast  Region = "asia-southeast1"
)

// RoutingIntent describes how the router should select providers.
type RoutingIntent struct {
	// Language is a BCP-47 tag, e.g. "en-US", "es-MX", "ja-JP".
	Language string `json:"language"`
	// Region pins latency measurements to a cloud region. Defaults to global.
	Region Region `json:"region,omitempty"`
	// Optimize selects the scoring preset. Defaults to balanced.
	Optimize OptimizeFor `json:"optimizeFor,omitempty"`
}

// ProviderConstraints restricts which providers the router may select.
// Each field is an allowlist of vendor names or "vendor:model" pairs.
type ProviderConstraints struct {
	STT []string `json:"stt,omitempty"`
	TTS []string `json:"tts,omitempty"`
	LLM []string `json:"llm,omitempty"`
}

// RoutingMetadata describes which provider/model the router selected.
type RoutingMetadata struct {
	Provider     string `json:"provider"`
	Model        string `json:"model"`
	FailoverCount int   `json:"failoverCount"`
}

// extractRouting reads routing decision headers from an HTTP response.
func extractRouting(h interface{ Get(string) string }) *RoutingMetadata {
	provider := h.Get("X-Speko-Provider")
	if provider == "" {
		return nil
	}
	return &RoutingMetadata{
		Provider: provider,
		Model:    h.Get("X-Speko-Model"),
	}
}

// Message is a chat-style message used in LLM completion requests.
type Message struct {
	Role    string `json:"role"`    // "system" | "user" | "assistant"
	Content string `json:"content"`
}

// Tool describes a function the LLM may call.
type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

// SessionMode selects the voice session architecture.
type SessionMode string

const (
	SessionModeCascade  SessionMode = "cascade"  // STT → LLM → TTS pipeline
	SessionModeS2S      SessionMode = "s2s"       // end-to-end speech-to-speech
	SessionModeRealtime SessionMode = "realtime"  // direct provider realtime API
)
