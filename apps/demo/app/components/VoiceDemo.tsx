"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { Mic, MicOff, PhoneOff, RotateCcw } from "lucide-react";
import type { ConversationMessage, RoutingMetadata } from "spekoai-react";
import { Waveform } from "./Waveform";
import { ModelBadge } from "./ModelBadge";

const LANGUAGES = [
  { code: "en-US", label: "English", flag: "🇺🇸" },
  { code: "es-MX", label: "Español", flag: "🇲🇽" },
  { code: "de-DE", label: "Deutsch", flag: "🇩🇪" },
  { code: "fr-FR", label: "Français", flag: "🇫🇷" },
  { code: "ar-SA", label: "العربية", flag: "🇸🇦" },
  { code: "ja-JP", label: "日本語", flag: "🇯🇵" },
  { code: "ko-KR", label: "한국어", flag: "🇰🇷" },
  { code: "hi-IN", label: "हिन्दी", flag: "🇮🇳" },
];

type OptimizeFor = "balanced" | "accuracy" | "latency" | "cost";

const OPTIMIZE_OPTIONS: { value: OptimizeFor; label: string }[] = [
  { value: "balanced", label: "Balanced" },
  { value: "accuracy", label: "Accuracy" },
  { value: "latency", label: "Latency" },
  { value: "cost", label: "Cost" },
];

// Simulated routing decisions for demo mode (no API key)
const DEMO_ROUTING: Record<string, RoutingMetadata[]> = {
  "en-US": [
    { provider: "deepgram", model: "nova-3", modality: "stt" },
    { provider: "openai", model: "gpt-4o-mini", modality: "llm" },
    { provider: "cartesia", model: "sonic-english", modality: "tts" },
  ],
  "es-MX": [
    { provider: "deepgram", model: "nova-3", modality: "stt" },
    { provider: "openai", model: "gpt-4o-mini", modality: "llm" },
    { provider: "elevenlabs", model: "multilingual-v2", modality: "tts" },
  ],
  "de-DE": [
    { provider: "assemblyai", model: "universal-3.6-pro", modality: "stt" },
    { provider: "openai", model: "gpt-4o-mini", modality: "llm" },
    { provider: "elevenlabs", model: "multilingual-v2", modality: "tts" },
  ],
  "fr-FR": [
    { provider: "assemblyai", model: "universal-3.6-pro", modality: "stt" },
    { provider: "openai", model: "gpt-4o-mini", modality: "llm" },
    { provider: "elevenlabs", model: "multilingual-v2", modality: "tts" },
  ],
  "ja-JP": [
    { provider: "google", model: "chirp-3-hd", modality: "stt" },
    { provider: "openai", model: "gpt-4o-mini", modality: "llm" },
    { provider: "elevenlabs", model: "multilingual-v2", modality: "tts" },
  ],
  "ko-KR": [
    { provider: "google", model: "chirp-3-hd", modality: "stt" },
    { provider: "openai", model: "gpt-4o-mini", modality: "llm" },
    { provider: "elevenlabs", model: "multilingual-v2", modality: "tts" },
  ],
  default: [
    { provider: "deepgram", model: "nova-3", modality: "stt" },
    { provider: "openai", model: "gpt-4o-mini", modality: "llm" },
    { provider: "elevenlabs", model: "multilingual-v2", modality: "tts" },
  ],
};

const DEMO_TRANSCRIPT: Record<string, { user: string; agent: string }> = {
  "en-US": {
    user: "Hi, can you help me schedule a meeting?",
    agent: "Of course! I'd be happy to help. What day and time works best for you?",
  },
  "es-MX": {
    user: "Hola, ¿puedes ayudarme a programar una reunión?",
    agent: "¡Por supuesto! Estaré encantado de ayudar. ¿Qué día y hora te viene mejor?",
  },
  "de-DE": {
    user: "Hallo, können Sie mir helfen, ein Meeting zu planen?",
    agent: "Natürlich! Ich helfe Ihnen gerne. Welcher Tag und welche Uhrzeit passt Ihnen?",
  },
  "fr-FR": {
    user: "Bonjour, pouvez-vous m'aider à planifier une réunion?",
    agent: "Bien sûr! Je serais ravi de vous aider. Quel jour et quelle heure vous convient?",
  },
  "ja-JP": {
    user: "こんにちは、会議のスケジュールを立てるのを手伝ってもらえますか？",
    agent: "もちろんです！喜んでお手伝いします。ご都合のよい日程と時間帯はいつですか？",
  },
  "ko-KR": {
    user: "안녕하세요, 회의 일정을 잡는 것을 도와주실 수 있나요?",
    agent: "물론이죠! 기꺼이 도와드리겠습니다. 어떤 날짜와 시간이 편하신가요?",
  },
  default: {
    user: "Hello, can you help me?",
    agent: "Of course! How can I assist you today?",
  },
};

type DemoStatus = "idle" | "connecting" | "connected" | "done";

export function VoiceDemo() {
  const [language, setLanguage] = useState("en-US");
  const [optimizeFor, setOptimizeFor] = useState<OptimizeFor>("balanced");
  const [status, setStatus] = useState<DemoStatus>("idle");
  const [mode, setMode] = useState<"listening" | "speaking" | null>(null);
  const [isMuted, setIsMuted] = useState(false);
  const [messages, setMessages] = useState<ConversationMessage[]>([]);
  const [routing, setRouting] = useState<RoutingMetadata[]>([]);
  const [latency, setLatency] = useState<number | null>(null);
  const transcriptRef = useRef<HTMLDivElement>(null);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const clearTimers = useCallback(() => {
    if (timerRef.current) clearTimeout(timerRef.current);
  }, []);

  useEffect(() => {
    if (transcriptRef.current) {
      transcriptRef.current.scrollTop = transcriptRef.current.scrollHeight;
    }
  }, [messages]);

  useEffect(() => () => clearTimers(), [clearTimers]);

  const runDemo = useCallback(() => {
    clearTimers();
    setStatus("connecting");
    setMessages([]);
    setRouting([]);
    setLatency(null);
    setMode(null);

    const demoRouting = DEMO_ROUTING[language] ?? DEMO_ROUTING.default;
    const demoScript = DEMO_TRANSCRIPT[language] ?? DEMO_TRANSCRIPT.default;
    const connectDelay = 600 + Math.random() * 400;
    const userDelay = 1200;
    const agentThinkDelay = 800 + Math.random() * 400;
    const agentSpeakDelay = 400;

    timerRef.current = setTimeout(() => {
      setStatus("connected");
      setRouting(demoRouting);
      const t0 = Date.now();

      timerRef.current = setTimeout(() => {
        setMode("listening");

        timerRef.current = setTimeout(() => {
          setMessages([{ source: "user", text: demoScript.user, isFinal: true }]);
          setMode(null);
          const elapsed = Date.now() - t0;
          setLatency(elapsed);

          timerRef.current = setTimeout(() => {
            setMode("speaking");

            timerRef.current = setTimeout(() => {
              setMessages((prev) => [
                ...prev,
                { source: "agent", text: demoScript.agent, isFinal: true },
              ]);
              setMode(null);
              setStatus("done");
            }, agentSpeakDelay + demoScript.agent.length * 18);
          }, agentThinkDelay);
        }, userDelay);
      }, 400);
    }, connectDelay);
  }, [language, clearTimers]);

  const reset = useCallback(() => {
    clearTimers();
    setStatus("idle");
    setMode(null);
    setMessages([]);
    setRouting([]);
    setLatency(null);
    setIsMuted(false);
  }, [clearTimers]);

  const isActive = status === "connected" || status === "done";
  const isConnecting = status === "connecting";

  return (
    <div className="rounded-2xl border border-[var(--border)] bg-[var(--surface)] overflow-hidden">
      {/* Header */}
      <div className="flex items-center justify-between px-5 py-4 border-b border-[var(--border)]">
        <div className="flex items-center gap-3">
          <div
            className={`w-2.5 h-2.5 rounded-full transition-colors duration-300 ${
              status === "connected"
                ? "bg-[var(--green)]"
                : status === "connecting"
                ? "bg-[var(--amber)]"
                : status === "done"
                ? "bg-indigo-400"
                : "bg-[var(--border)]"
            }`}
          />
          <span className="text-sm font-medium text-[var(--muted)]">
            {status === "connecting"
              ? "Connecting…"
              : status === "connected"
              ? mode === "listening"
                ? "Listening…"
                : mode === "speaking"
                ? "Agent speaking…"
                : "Connected"
              : status === "done"
              ? "Session ended"
              : "Ready"}
          </span>
        </div>

        <div className="flex items-center gap-2">
          {latency !== null && (
            <span className="text-xs text-[var(--muted)] font-mono">
              {latency}ms first token
            </span>
          )}
          {status !== "idle" && (
            <button
              onClick={reset}
              className="p-1.5 rounded-lg hover:bg-[var(--surface-elevated)] text-[var(--muted)] hover:text-[var(--text)] transition-colors"
              title="Reset demo"
            >
              <RotateCcw size={14} />
            </button>
          )}
        </div>
      </div>

      {/* Controls */}
      <div className="px-5 py-4 border-b border-[var(--border)] flex flex-wrap gap-3 items-center">
        <div className="flex items-center gap-2 flex-wrap">
          <span className="text-xs text-[var(--muted)] shrink-0">Language:</span>
          <div className="flex flex-wrap gap-1.5">
            {LANGUAGES.map((lang) => (
              <button
                key={lang.code}
                onClick={() => { setLanguage(lang.code); reset(); }}
                className={`text-xs px-2.5 py-1 rounded-lg border transition-all ${
                  language === lang.code
                    ? "bg-indigo-500/20 border-indigo-500/50 text-indigo-300"
                    : "border-[var(--border)] text-[var(--muted)] hover:border-zinc-500 hover:text-[var(--text)]"
                }`}
              >
                {lang.flag} {lang.label}
              </button>
            ))}
          </div>
        </div>

        <div className="flex items-center gap-2 ml-auto">
          <span className="text-xs text-[var(--muted)] shrink-0">Optimize:</span>
          <select
            value={optimizeFor}
            onChange={(e) => setOptimizeFor(e.target.value as OptimizeFor)}
            className="text-xs bg-[var(--surface-elevated)] border border-[var(--border)] rounded-lg px-2.5 py-1.5 text-[var(--text)] cursor-pointer"
          >
            {OPTIMIZE_OPTIONS.map((o) => (
              <option key={o.value} value={o.value}>
                {o.label}
              </option>
            ))}
          </select>
        </div>
      </div>

      {/* Waveform */}
      <div className="flex flex-col items-center py-8 gap-4 border-b border-[var(--border)]">
        <div className="relative">
          {isActive && mode && (
            <div
              className="pulse-ring absolute inset-0 rounded-full border-2"
              style={{
                borderColor: mode === "speaking" ? "#6366f1" : "#22c55e",
              }}
            />
          )}
          <Waveform active={isActive && mode !== null} mode={mode} />
        </div>

        <div className="flex gap-3">
          {status === "idle" || status === "done" ? (
            <button
              onClick={runDemo}
              className="flex items-center gap-2 px-5 py-2.5 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-sm font-medium transition-colors"
            >
              <Mic size={16} />
              {status === "done" ? "Try again" : "Start demo"}
            </button>
          ) : (
            <>
              <button
                disabled={isConnecting}
                onClick={() => setIsMuted((m) => !m)}
                className={`flex items-center gap-2 px-4 py-2 rounded-xl border text-sm font-medium transition-all ${
                  isMuted
                    ? "bg-red-500/20 border-red-500/50 text-red-300"
                    : "border-[var(--border)] text-[var(--muted)] hover:border-zinc-500 hover:text-[var(--text)]"
                }`}
              >
                {isMuted ? <MicOff size={15} /> : <Mic size={15} />}
                {isMuted ? "Unmute" : "Mute"}
              </button>
              <button
                onClick={reset}
                className="flex items-center gap-2 px-4 py-2 rounded-xl border border-red-500/30 text-red-400 hover:bg-red-500/10 text-sm font-medium transition-colors"
              >
                <PhoneOff size={15} />
                End
              </button>
            </>
          )}
        </div>
      </div>

      {/* Routing metadata */}
      {routing.length > 0 && (
        <div className="px-5 py-3 border-b border-[var(--border)] flex flex-wrap gap-2 items-center">
          <span className="text-xs text-[var(--muted)] shrink-0 mr-1">Routing →</span>
          {routing.map((r) => (
            <ModelBadge key={r.modality} metadata={r} />
          ))}
          <span className="text-xs text-[var(--muted)] ml-auto">
            best for <span className="text-[var(--text)]">{language}</span> ·{" "}
            <span className="text-[var(--text)]">{optimizeFor}</span>
          </span>
        </div>
      )}

      {/* Transcript */}
      <div
        ref={transcriptRef}
        className="px-5 py-4 min-h-[160px] max-h-64 overflow-y-auto space-y-3"
      >
        {messages.length === 0 ? (
          <p className="text-sm text-[var(--muted)] text-center pt-6">
            {isConnecting ? "Establishing session…" : 'Press "Start demo" to see routing in action'}
          </p>
        ) : (
          messages.map((msg, i) => (
            <div
              key={i}
              className={`flex gap-3 animate-fade-up ${
                msg.source === "user" ? "justify-end" : "justify-start"
              }`}
            >
              {msg.source === "agent" && (
                <div className="w-6 h-6 rounded-full bg-indigo-500/20 border border-indigo-500/30 flex items-center justify-center shrink-0 mt-0.5">
                  <span className="text-[10px] text-indigo-300 font-bold">AI</span>
                </div>
              )}
              <div
                className={`max-w-[75%] rounded-2xl px-4 py-2.5 text-sm leading-relaxed ${
                  msg.source === "user"
                    ? "bg-indigo-600/30 border border-indigo-500/30 text-indigo-100"
                    : "bg-[var(--surface-elevated)] border border-[var(--border)] text-[var(--text)]"
                }`}
              >
                {msg.text}
              </div>
            </div>
          ))
        )}
      </div>

      {/* Footer note */}
      <div className="px-5 py-3 border-t border-[var(--border)]">
        <p className="text-xs text-[var(--muted)]">
          Demo mode — routing decisions are simulated from{" "}
          <a
            href="https://benchmarks.speko.ai"
            target="_blank"
            rel="noopener noreferrer"
            className="text-indigo-400 hover:text-indigo-300 underline underline-offset-2"
          >
            benchmarks.speko.ai
          </a>
          . Add a{" "}
          <code className="font-mono bg-[var(--surface-elevated)] px-1 rounded">
            SPEKO_API_KEY
          </code>{" "}
          env var to run live.
        </p>
      </div>
    </div>
  );
}
