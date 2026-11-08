# @spekoai/react

> React hooks for [Speko Voice AI](https://speko.ai) — build multilingual voice agents with zero boilerplate.

Speko routes each STT / LLM / TTS call to the best-scoring provider for your language and optimization target, based on live benchmarks. This package wraps `@spekoai/client` in idiomatic React hooks so you don't have to manage transport, transcript accumulation, or provider-key security yourself.

## Packages

| Package | Description |
|---|---|
| [`packages/react`](./packages/react) | React hooks library — `useVoiceConversation`, `useTranscription`, `useSynthesis`, `useMicrophoneStream` |
| [`apps/demo`](./apps/demo) | Interactive Next.js demo — live routing visualization, language switcher, code examples |

## Quickstart

```bash
npm install @spekoai/react @spekoai/client
```

```tsx
import { useVoiceConversation } from "@spekoai/react";

function VoiceAgent() {
  const { status, mode, messages, routing, start, end } =
    useVoiceConversation({ sessionEndpoint: "/api/speko-session" });

  return (
    <>
      <button onClick={status === "idle" ? start : end}>
        {status === "idle" ? "Talk" : "End"}
      </button>

      {routing && <p>Routed to: {routing.provider}/{routing.model}</p>}

      {messages.map((m, i) => (
        <div key={i}>{m.source}: {m.text}</div>
      ))}
    </>
  );
}
```

Add a session endpoint on your server (keeps your API key safe):

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

## Hooks

### `useVoiceConversation(options)`

Full-duplex voice session over LiveKit. Returns `status`, `mode`, `messages`, `routing`, `start()`, `end()`, `setMuted()`, `setVolume()`, `sendMessage()`.

### `useTranscription(options?)`

One-shot speech-to-text with routing. Returns `transcribe(audio)`, `isLoading`, `error`.
Routing metadata tells you which provider/model won for your language.

### `useSynthesis(options?)`

One-shot text-to-speech with routing. Returns `synthesize(text)`, `isLoading`, `error`.

### `useMicrophoneStream()`

Browser microphone access with real-time `audioLevel` (0–1) for waveform visualisation.

## Routing metadata

Every hook surfaces a `RoutingMetadata` object so you can show users which provider was selected:

```ts
{
  provider: "deepgram",
  model: "nova-3",
  modality: "stt",
  failoverCount: 0,
}
```

## Run the demo locally

```bash
git clone https://github.com/ilaszajsenbaev/spekoai-react
cd spekoai-react
pnpm install
cp apps/demo/.env.example apps/demo/.env.local
# optionally add SPEKO_API_KEY to run against the real API
pnpm dev
```

Open [http://localhost:3000](http://localhost:3000).

## License

MIT — built by [Ilasz Ajsenbaev](https://github.com/ilaszajsenbaev) for [Speko (YC S26)](https://speko.ai).
