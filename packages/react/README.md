# @spekoai/react

React hooks for [Speko Voice AI](https://speko.ai) — build multilingual voice agents with zero boilerplate.

```bash
npm install @spekoai/react @spekoai/client
```

## Hooks

### `useVoiceConversation`

Full-duplex voice session. One hook handles transport, transcripts, mode changes, and routing metadata.

```tsx
import { useVoiceConversation } from "@spekoai/react";

function VoiceAgent() {
  const { status, mode, messages, routing, start, end, setMuted } =
    useVoiceConversation({
      sessionEndpoint: "/api/speko-session",
    });

  return (
    <>
      <button onClick={status === "idle" ? start : end}>
        {status === "idle" ? "Talk" : "End"}
      </button>

      {routing && (
        <p>
          {routing.provider}/{routing.model} ({routing.modality})
        </p>
      )}

      {messages.map((m, i) => (
        <div key={i} className={m.source}>
          {m.text}
        </div>
      ))}
    </>
  );
}
```

The `sessionEndpoint` is a route on **your** server that calls `POST /v1/sessions` with your `SPEKO_API_KEY` and returns `{ transportToken, transportUrl }`. Never expose the key to the browser.

```ts
// app/api/speko-session/route.ts
export async function POST(req: Request) {
  const { language, optimizeFor } = await req.json();
  const res = await fetch("https://api.speko.dev/v1/sessions", {
    method: "POST",
    headers: { Authorization: `Bearer ${process.env.SPEKO_API_KEY}` },
    body: JSON.stringify({
      agentId: process.env.SPEKO_AGENT_ID,
      mode: "cascade",
      intent: { language, optimizeFor },
    }),
  });
  return Response.json(await res.json());
}
```

### `useTranscription`

One-shot speech-to-text with routing.

```tsx
import { useTranscription, useMicrophoneStream } from "@spekoai/react";

function Transcriber() {
  const { startRecording, stopRecording, isRecording, audioLevel } =
    useMicrophoneStream();
  const { transcribe, isLoading } = useTranscription({ language: "es-MX" });

  const handleStop = async () => {
    const blob = stopRecording();
    if (!blob) return;
    const { text, routing } = await transcribe(blob);
    // routing → { provider: "deepgram", model: "nova-3", modality: "stt" }
  };

  return (
    <button onClick={isRecording ? handleStop : startRecording}>
      {isRecording ? `Stop (level: ${audioLevel.toFixed(2)})` : "Record"}
    </button>
  );
}
```

### `useSynthesis`

One-shot text-to-speech with routing.

```tsx
import { useSynthesis } from "@spekoai/react";

function Speaker() {
  const { synthesize, isLoading } = useSynthesis({ language: "ja-JP" });

  const speak = async () => {
    const { audioUrl, routing } = await synthesize("こんにちは");
    new Audio(audioUrl).play();
    // routing → { provider: "elevenlabs", model: "multilingual-v2", modality: "tts" }
  };

  return <button onClick={speak} disabled={isLoading}>Speak</button>;
}
```

### `useMicrophoneStream`

Browser mic access with real-time audio level for waveform visualisation.

```tsx
const { startRecording, stopRecording, isRecording, audioLevel, error } =
  useMicrophoneStream();
```

Returns:
- `audioLevel` — `0–1` amplitude, updates at 60fps
- `stopRecording()` — returns a `Blob` of recorded audio
- `error` — set if mic permission was denied

## Routing metadata

Every hook that calls the Speko API surfaces a `RoutingMetadata` object:

```ts
interface RoutingMetadata {
  provider: string;      // e.g. "deepgram"
  model: string;         // e.g. "nova-3"
  modality: "stt" | "tts" | "llm" | "s2s";
  failoverCount?: number;
  scoreBreakdown?: {
    quality: number;
    latency: number;
    cost: number;
    composite: number;
  };
}
```

Speko selects the top-ranked provider from [benchmarks.speko.ai](https://benchmarks.speko.ai) for your `language` + `optimizeFor` combination. No manual configuration needed.

## RoutingIntent options

Pass these in `sessionPayload` (for voice sessions) or directly to `useTranscription` / `useSynthesis`:

| Field | Values | Default |
|---|---|---|
| `language` | BCP-47 (e.g. `"en-US"`, `"es-MX"`) | required |
| `region` | `"global"` `"us-east4"` `"europe-west3"` `"asia-southeast1"` | `"global"` |
| `optimizeFor` | `"balanced"` `"accuracy"` `"latency"` `"cost"` | `"balanced"` |

## TypeScript

All types are exported:

```ts
import type {
  ConversationMessage,
  ConversationStatus,
  ConversationMode,
  RoutingMetadata,
  RoutingIntent,
  VoiceConversationOptions,
} from "@spekoai/react";
```

## Peer dependencies

| Package | Required |
|---|---|
| `react` | ≥ 18 |
| `@spekoai/client` | ≥ 0.0.1 |
| `@spekoai/sdk` | optional (needed for `useTranscription`, `useSynthesis`) |

## License

MIT
