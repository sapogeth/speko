import { Github, ExternalLink, Package } from "lucide-react";
import { VoiceDemo } from "./components/VoiceDemo";
import { CodeBlock } from "./components/CodeBlock";

const INSTALL_CODE = `npm install @spekoai/react @spekoai/client`;

const QUICKSTART_CODE = `import { useVoiceConversation } from "@spekoai/react";

export function VoiceAgent() {
  const { status, mode, messages, start, end, setMuted } =
    useVoiceConversation({
      sessionEndpoint: "/api/speko-session",
    });

  return (
    <div>
      <p>Status: {status} · Mode: {mode ?? "—"}</p>

      {status === "idle" && (
        <button onClick={start}>Start conversation</button>
      )}
      {status === "connected" && (
        <button onClick={end}>End session</button>
      )}

      {messages.map((msg, i) => (
        <div key={i} className={msg.source}>
          {msg.text}
        </div>
      ))}
    </div>
  );
}`;

const TRANSCRIPTION_CODE = `import { useTranscription, useMicrophoneStream } from "@spekoai/react";

export function TranscribeTool() {
  const { startRecording, stopRecording, isRecording } = useMicrophoneStream();
  const { transcribe, isLoading } = useTranscription({ language: "es-MX" });

  const handleStop = async () => {
    const blob = stopRecording();
    if (!blob) return;

    const { text, routing } = await transcribe(blob);
    // routing → { provider: "deepgram", model: "nova-3", modality: "stt" }
    console.log(text, routing);
  };

  return (
    <button onClick={isRecording ? handleStop : startRecording}>
      {isRecording ? "Stop & Transcribe" : "Record"}
    </button>
  );
}`;

const SESSION_ROUTE_CODE = `// app/api/speko-session/route.ts
import { NextRequest, NextResponse } from "next/server";

export async function POST(req: NextRequest) {
  const { language, optimizeFor } = await req.json();

  const res = await fetch("https://api.speko.dev/v1/sessions", {
    method: "POST",
    headers: {
      Authorization: \`Bearer \${process.env.SPEKO_API_KEY}\`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      agentId: process.env.SPEKO_AGENT_ID,
      mode: "cascade",
      intent: { language, optimizeFor },
    }),
  });

  const data = await res.json();
  return NextResponse.json(data);
}`;

export default function HomePage() {
  return (
    <main className="min-h-screen" style={{ background: "var(--bg)" }}>
      {/* Nav */}
      <nav
        className="sticky top-0 z-50 border-b"
        style={{
          background: "rgba(9,9,11,0.85)",
          borderColor: "var(--border)",
          backdropFilter: "blur(12px)",
        }}
      >
        <div className="max-w-6xl mx-auto px-6 py-3.5 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <span
              className="font-mono text-sm font-semibold px-2 py-0.5 rounded-md"
              style={{ background: "var(--accent-glow)", color: "#a5b4fc" }}
            >
              @spekoai/react
            </span>
            <span
              className="text-xs px-2 py-0.5 rounded-full border"
              style={{ color: "var(--muted)", borderColor: "var(--border)" }}
            >
              v0.1.0
            </span>
          </div>
          <div className="flex items-center gap-4">
            <a
              href="https://docs.speko.ai"
              target="_blank"
              rel="noopener noreferrer"
              className="text-sm flex items-center gap-1.5 transition-colors"
              style={{ color: "var(--muted)" }}
            >
              <ExternalLink size={13} />
              Speko docs
            </a>
            <a
              href="https://github.com/sapogeth/speko"
              target="_blank"
              rel="noopener noreferrer"
              className="text-sm flex items-center gap-1.5 transition-colors"
              style={{ color: "var(--muted)" }}
            >
              <Github size={14} />
              GitHub
            </a>
          </div>
        </div>
      </nav>

      <div className="max-w-6xl mx-auto px-6 py-16 space-y-24">
        {/* Hero */}
        <section className="text-center space-y-6 pt-4">
          <div className="inline-flex items-center gap-2 text-xs font-mono px-3 py-1.5 rounded-full border" style={{ background: "var(--surface)", borderColor: "var(--border)", color: "var(--muted)" }}>
            <Package size={12} className="text-indigo-400" />
            React hooks for{" "}
            <a
              href="https://speko.ai"
              target="_blank"
              rel="noopener noreferrer"
              className="text-indigo-400 hover:text-indigo-300"
            >
              Speko Voice AI
            </a>
          </div>

          <h1
            className="text-5xl sm:text-6xl font-bold tracking-tight leading-tight"
            style={{ color: "var(--text)" }}
          >
            Build voice agents
            <br />
            <span
              style={{
                background: "linear-gradient(135deg, #6366f1 0%, #a78bfa 50%, #38bdf8 100%)",
                WebkitBackgroundClip: "text",
                WebkitTextFillColor: "transparent",
              }}
            >
              in 10 lines of React.
            </span>
          </h1>

          <p
            className="text-lg max-w-2xl mx-auto leading-relaxed"
            style={{ color: "var(--muted)" }}
          >
            <code
              className="font-mono text-sm px-1.5 py-0.5 rounded"
              style={{ background: "var(--surface-elevated)", color: "#a5b4fc" }}
            >
              @spekoai/react
            </code>{" "}
            wraps{" "}
            <a
              href="https://docs.speko.ai/client"
              target="_blank"
              rel="noopener noreferrer"
              className="underline underline-offset-2 decoration-zinc-600 hover:decoration-zinc-400 transition-colors"
              style={{ color: "var(--muted)" }}
            >
              @spekoai/client
            </a>{" "}
            in routing-aware React hooks — so you get automatic{" "}
            <strong style={{ color: "var(--text)" }}>
              STT → LLM → TTS model selection
            </strong>{" "}
            per language, without thinking about providers.
          </p>

          <div className="flex flex-wrap justify-center gap-3 pt-2">
            {[
              "useVoiceConversation",
              "useTranscription",
              "useSynthesis",
              "useMicrophoneStream",
            ].map((h) => (
              <span
                key={h}
                className="font-mono text-sm px-3 py-1 rounded-lg border"
                style={{
                  background: "var(--surface)",
                  borderColor: "var(--border)",
                  color: "#a5b4fc",
                }}
              >
                {h}
              </span>
            ))}
          </div>
        </section>

        {/* Live demo */}
        <section className="space-y-5">
          <div className="space-y-1">
            <h2 className="text-xl font-semibold" style={{ color: "var(--text)" }}>
              Live demo
            </h2>
            <p className="text-sm" style={{ color: "var(--muted)" }}>
              Switch language and watch Speko route to the best-scoring providers
              from{" "}
              <a
                href="https://benchmarks.speko.ai"
                target="_blank"
                rel="noopener noreferrer"
                className="underline underline-offset-2 hover:text-indigo-300 transition-colors"
              >
                benchmarks.speko.ai
              </a>
              .
            </p>
          </div>
          <VoiceDemo />
        </section>

        {/* Install */}
        <section className="space-y-5">
          <h2 className="text-xl font-semibold" style={{ color: "var(--text)" }}>
            Install
          </h2>
          <CodeBlock code={INSTALL_CODE} language="bash" />
        </section>

        {/* Quickstart grid */}
        <section className="space-y-5">
          <h2 className="text-xl font-semibold" style={{ color: "var(--text)" }}>
            Quickstart
          </h2>
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-5">
            <div className="space-y-3">
              <h3 className="text-sm font-medium" style={{ color: "var(--muted)" }}>
                1. Full voice conversation
              </h3>
              <CodeBlock
                code={QUICKSTART_CODE}
                language="tsx"
                filename="VoiceAgent.tsx"
              />
            </div>
            <div className="space-y-3">
              <h3 className="text-sm font-medium" style={{ color: "var(--muted)" }}>
                2. One-shot transcription
              </h3>
              <CodeBlock
                code={TRANSCRIPTION_CODE}
                language="tsx"
                filename="TranscribeTool.tsx"
              />
            </div>
          </div>
          <div className="space-y-3">
            <h3 className="text-sm font-medium" style={{ color: "var(--muted)" }}>
              3. Session API route (Next.js)
            </h3>
            <CodeBlock
              code={SESSION_ROUTE_CODE}
              language="ts"
              filename="app/api/speko-session/route.ts"
            />
          </div>
        </section>

        {/* Hook reference */}
        <section className="space-y-5">
          <h2 className="text-xl font-semibold" style={{ color: "var(--text)" }}>
            Hook reference
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {[
              {
                name: "useVoiceConversation",
                desc: "Full duplex voice session over LiveKit. Handles transport, transcripts, mode changes, and routing metadata in one hook.",
                returns: ["status", "mode", "messages", "routing", "start()", "end()", "setMuted()"],
              },
              {
                name: "useTranscription",
                desc: "One-shot speech-to-text with Speko routing. Returns the transcript and the winning provider/model.",
                returns: ["transcribe(audio)", "isLoading", "error"],
              },
              {
                name: "useSynthesis",
                desc: "One-shot text-to-speech with routing. Returns a blob URL and the winning provider/model.",
                returns: ["synthesize(text)", "isLoading", "error"],
              },
              {
                name: "useMicrophoneStream",
                desc: "Browser microphone access with real-time audio level for waveform visualisation.",
                returns: ["startRecording()", "stopRecording()", "isRecording", "audioLevel"],
              },
            ].map((hook) => (
              <div
                key={hook.name}
                className="rounded-xl border p-5 space-y-3"
                style={{
                  background: "var(--surface)",
                  borderColor: "var(--border)",
                }}
              >
                <div className="space-y-1">
                  <h3
                    className="font-mono text-sm font-semibold"
                    style={{ color: "#a5b4fc" }}
                  >
                    {hook.name}
                  </h3>
                  <p className="text-sm leading-relaxed" style={{ color: "var(--muted)" }}>
                    {hook.desc}
                  </p>
                </div>
                <div className="flex flex-wrap gap-1.5">
                  {hook.returns.map((r) => (
                    <span
                      key={r}
                      className="font-mono text-xs px-2 py-0.5 rounded-md border"
                      style={{
                        background: "var(--surface-elevated)",
                        borderColor: "var(--border)",
                        color: "var(--muted)",
                      }}
                    >
                      {r}
                    </span>
                  ))}
                </div>
              </div>
            ))}
          </div>
        </section>

        {/* Why section */}
        <section className="space-y-5">
          <h2 className="text-xl font-semibold" style={{ color: "var(--text)" }}>
            Why this exists
          </h2>
          <div
            className="rounded-xl border p-6 space-y-4"
            style={{ background: "var(--surface)", borderColor: "var(--border)" }}
          >
            <p className="text-sm leading-7" style={{ color: "var(--muted)" }}>
              Speko&apos;s TypeScript SDK and browser client are excellent, but wiring them into
              React still takes boilerplate: session lifecycle, state management, transcript
              accumulation, provider-key protection. <strong style={{ color: "var(--text)" }}>
                @spekoai/react
              </strong>{" "}
              collapses that into four hooks so you can focus on your product, not plumbing.
            </p>
            <p className="text-sm leading-7" style={{ color: "var(--muted)" }}>
              Routing metadata — which STT/LLM/TTS provider won for your language — surfaces
              as a typed return value, so you can visualise it, log it, or gate on it without
              digging through response headers manually.
            </p>
          </div>
        </section>

        {/* Footer */}
        <footer
          className="border-t pt-8 pb-4 flex flex-col sm:flex-row items-center justify-between gap-4"
          style={{ borderColor: "var(--border)" }}
        >
          <p className="text-sm" style={{ color: "var(--muted)" }}>
            Built for{" "}
            <a
              href="https://speko.ai"
              target="_blank"
              rel="noopener noreferrer"
              className="underline underline-offset-2 hover:text-indigo-300 transition-colors"
            >
              Speko (YC S26)
            </a>{" "}
            by Ilyas
          </p>
          <div className="flex items-center gap-4">
            <a
              href="https://speko.ai"
              target="_blank"
              rel="noopener noreferrer"
              className="text-sm flex items-center gap-1.5 transition-colors"
              style={{ color: "var(--muted)" }}
            >
              <ExternalLink size={12} />
              speko.ai
            </a>
            <a
              href="https://benchmarks.speko.ai"
              target="_blank"
              rel="noopener noreferrer"
              className="text-sm flex items-center gap-1.5 transition-colors"
              style={{ color: "var(--muted)" }}
            >
              <ExternalLink size={12} />
              Benchmarks
            </a>
          </div>
        </footer>
      </div>
    </main>
  );
}
