# spekoai-go

The Go SDK for [Speko Voice AI](https://speko.ai) — routing-aware STT, TTS, and LLM completion with streaming support.

```bash
go get github.com/ilaszajsenbaev/spekoai-go
```

Requires Go 1.21+.

## Quickstart

```go
package main

import (
    "context"
    "fmt"
    "os"

    spekoai "github.com/ilaszajsenbaev/spekoai-go"
)

func main() {
    client := spekoai.New(os.Getenv("SPEKO_API_KEY"))

    // Transcribe — routes to best STT for es-MX based on live benchmarks
    result, err := client.Transcribe(context.Background(), spekoai.TranscribeRequest{
        Audio:    audioBytes,
        Language: "es-MX",
        Optimize: spekoai.OptimizeAccuracy,
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(result.Text)
    fmt.Printf("routed to: %s/%s\n", result.Routing.Provider, result.Routing.Model)
    // → deepgram/nova-3
}
```

## Routing

Every API call accepts a `RoutingIntent` that tells Speko how to score providers:

| Field | Type | Default |
|---|---|---|
| `Language` | BCP-47 string — `"en-US"`, `"ja-JP"`, `"es-MX"` | required |
| `Region` | `RegionGlobal`, `RegionUSEast`, `RegionEuropeWest`, `RegionAsiaSoutheast` | `global` |
| `Optimize` | `OptimizeBalanced`, `OptimizeAccuracy`, `OptimizeLatency`, `OptimizeCost` | `balanced` |

Routing metadata comes back on every result:

```go
result.Routing.Provider      // "deepgram"
result.Routing.Model         // "nova-3"
result.Routing.FailoverCount // 0
```

## STT — Transcribe

```go
// Batch
result, err := client.Transcribe(ctx, spekoai.TranscribeRequest{
    Audio:    audioBytes,     // webm, wav, mp3, ogg, flac
    Language: "de-DE",
    Optimize: spekoai.OptimizeLatency,
    Constraints: &spekoai.ProviderConstraints{
        STT: []string{"deepgram", "assemblyai"},
    },
})

// Streaming (live words as they arrive)
ch, err := client.StreamTranscribe(ctx, spekoai.StreamTranscribeRequest{
    Audio:    micStream,
    Language: "ja-JP",
})
for evt := range ch {
    if evt.Err != nil { break }
    fmt.Printf("%s (final=%v)\n", evt.Text, evt.IsFinal)
}
```

## TTS — Synthesize

```go
// Batch
result, err := client.Synthesize(ctx, spekoai.SynthesizeRequest{
    Text:     "Hola, ¿en qué puedo ayudarte?",
    Language: "es-MX",
    Format:   spekoai.AudioFormatWAV,
    Optimize: spekoai.OptimizeCost,
})
os.WriteFile("output.wav", result.Audio, 0644)

// Streaming (chunks arrive before full synthesis completes)
ch, err := client.StreamSynthesize(ctx, spekoai.SynthesizeRequest{
    Text:     "...",
    Language: "ko-KR",
    Format:   spekoai.AudioFormatPCM,
})
for evt := range ch {
    if evt.Done { break }
    audioPlayer.Write(evt.Chunk)
}
```

## LLM — Complete

```go
// Batch (accumulates streaming internally)
result, err := client.Complete(ctx, spekoai.CompleteRequest{
    Messages: []spekoai.Message{
        {Role: "system", Content: "You are a voice assistant."},
        {Role: "user", Content: userSpeech},
    },
    Language:  "en-US",
    MaxTokens: 200,
    Tools: []spekoai.Tool{{
        Name:        "book_appointment",
        Description: "Book a meeting slot",
        Parameters:  schema,
    }},
})

// Streaming
ch, err := client.StreamComplete(ctx, spekoai.CompleteRequest{...})
for evt := range ch {
    if evt.Done { break }
    fmt.Print(evt.Delta)
}
```

## Sessions

```go
// Create a session — returns browser transport credentials
// Call this in your HTTP handler; never in the browser.
session, err := client.CreateSession(ctx, spekoai.CreateSessionRequest{
    AgentID: os.Getenv("SPEKO_AGENT_ID"),
    Mode:    spekoai.SessionModeCascade,
    Variables: map[string]string{
        "user_name": "Alice",
    },
})
// Pass session.TransportToken and session.TransportURL to @spekoai/client
```

## Agents

```go
agent, err := client.CreateAgent(ctx, spekoai.CreateAgentRequest{
    Name:   "Support Bot",
    Prompt: "You are a helpful support agent...",
    Intent: &spekoai.RoutingIntent{
        Language: "en-US",
        Optimize: spekoai.OptimizeLatency,
    },
})

// Outbound call
call, err := client.DialPhone(ctx, spekoai.PhoneDialRequest{
    To:             "+1555...",
    AgentID:        agent.ID,
    IdempotencyKey: "booking-123", // safe to retry
})
```

## Error handling

```go
_, err := client.Transcribe(ctx, req)
var authErr *spekoai.AuthError
var rlErr *spekoai.RateLimitError
var npErr *spekoai.NoProviderError

switch {
case errors.As(err, &authErr):
    // Check SPEKO_API_KEY
case errors.As(err, &rlErr):
    time.Sleep(rlErr.RetryAfter)
    // retry
case errors.As(err, &npErr):
    // No provider supports this language/region combination
default:
    log.Fatal(err)
}
```

## CLI example

```bash
SPEKO_API_KEY=sk_... go run ./examples/route-benchmark \
  -text "Buenos días, ¿cómo puedo ayudarle?" \
  -lang es-MX \
  -optimize accuracy

  spekoai-go route-benchmark
  language:    es-MX
  optimize:    accuracy
  region:      global
  text:        "Buenos días, ¿cómo puedo ayudarle?"

  Synthesize  elevenlabs           multilingual-v2          342ms
  Complete    openai               gpt-4o-mini              891ms

  reply: "¡Buenos días! Estoy aquí para ayudarle..."
```

## License

MIT — built by [Ilasz Ajsenbaev](https://github.com/ilaszajsenbaev) for [Speko (YC S26)](https://speko.ai).
