// route-benchmark is a CLI that shows which Speko provider wins for a given
// language and optimization target.
//
// Usage:
//
//	SPEKO_API_KEY=sk_... go run ./examples/route-benchmark \
//	  -text "Buenos días, ¿cómo puedo ayudarle?" \
//	  -lang es-MX \
//	  -optimize accuracy
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	spekoai "github.com/ilaszajsenbaev/spekoai-go"
)

func main() {
	text := flag.String("text", "Hello, how can I help you today?", "Text to synthesize")
	lang := flag.String("lang", "en-US", "BCP-47 language tag (e.g. es-MX, ja-JP, de-DE)")
	optimize := flag.String("optimize", "balanced", "balanced | accuracy | latency | cost")
	region := flag.String("region", "global", "global | us-east4 | europe-west3 | asia-southeast1")
	flag.Parse()

	apiKey := os.Getenv("SPEKO_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "error: SPEKO_API_KEY environment variable is not set")
		fmt.Fprintln(os.Stderr, "  get a key at https://app.speko.ai")
		os.Exit(1)
	}

	client := spekoai.New(apiKey)

	fmt.Printf("\n  spekoai-go route-benchmark\n")
	fmt.Printf("  %-12s %s\n", "language:", *lang)
	fmt.Printf("  %-12s %s\n", "optimize:", *optimize)
	fmt.Printf("  %-12s %s\n", "region:", *region)
	fmt.Printf("  %-12s %q\n\n", "text:", truncate(*text, 60))

	// --- TTS routing ---
	fmt.Print("  Synthesize  ")
	t0 := time.Now()

	synthResult, err := client.Synthesize(context.Background(), spekoai.SynthesizeRequest{
		Text:     *text,
		Language: *lang,
		Optimize: spekoai.OptimizeFor(*optimize),
		Region:   spekoai.Region(*region),
		Format:   spekoai.AudioFormatWAV,
	})

	ttsLatency := time.Since(t0)
	if err != nil {
		fmt.Printf("FAILED: %v\n\n", err)
	} else {
		printRouting(synthResult.Routing, ttsLatency, len(synthResult.Audio))
	}

	// --- LLM routing ---
	fmt.Print("  Complete    ")
	t0 = time.Now()

	llmResult, err := client.Complete(context.Background(), spekoai.CompleteRequest{
		Messages: []spekoai.Message{
			{Role: "system", Content: "You are a helpful assistant. Reply in the same language as the user."},
			{Role: "user", Content: *text},
		},
		Language: *lang,
		Optimize: spekoai.OptimizeFor(*optimize),
		Region:   spekoai.Region(*region),
		MaxTokens: 60,
	})

	llmLatency := time.Since(t0)
	if err != nil {
		fmt.Printf("FAILED: %v\n\n", err)
	} else {
		printRouting(llmResult.Routing, llmLatency, len(llmResult.Content))
		fmt.Printf("\n  reply: %q\n", truncate(llmResult.Content, 80))
	}

	fmt.Println()
}

func printRouting(r *spekoai.RoutingMetadata, latency time.Duration, size int) {
	if r == nil {
		fmt.Printf("%-20s %-24s %8s   %d bytes\n", "(no routing header)", "", latency.Round(time.Millisecond), size)
		return
	}
	failover := ""
	if r.FailoverCount > 0 {
		failover = fmt.Sprintf(" (+%d failover)", r.FailoverCount)
	}
	fmt.Printf("%-20s %-24s %8s%s\n",
		r.Provider,
		r.Model,
		latency.Round(time.Millisecond),
		failover,
	)
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
