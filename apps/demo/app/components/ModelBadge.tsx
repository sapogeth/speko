import type { RoutingMetadata } from "spekoai-react";

const PROVIDER_COLORS: Record<string, string> = {
  deepgram: "bg-blue-500/15 text-blue-300 border-blue-500/30",
  openai: "bg-emerald-500/15 text-emerald-300 border-emerald-500/30",
  assemblyai: "bg-purple-500/15 text-purple-300 border-purple-500/30",
  elevenlabs: "bg-orange-500/15 text-orange-300 border-orange-500/30",
  cartesia: "bg-pink-500/15 text-pink-300 border-pink-500/30",
  google: "bg-yellow-500/15 text-yellow-300 border-yellow-500/30",
  "xai": "bg-zinc-500/15 text-zinc-300 border-zinc-500/30",
  default: "bg-indigo-500/15 text-indigo-300 border-indigo-500/30",
};

const MODALITY_LABEL: Record<string, string> = {
  stt: "STT",
  tts: "TTS",
  llm: "LLM",
  s2s: "S2S",
};

interface ModelBadgeProps {
  metadata: RoutingMetadata;
}

export function ModelBadge({ metadata }: ModelBadgeProps) {
  const colorClass =
    PROVIDER_COLORS[metadata.provider.toLowerCase()] ?? PROVIDER_COLORS.default;

  return (
    <span
      className={`inline-flex items-center gap-1.5 text-xs font-mono px-2 py-1 rounded-md border ${colorClass}`}
    >
      <span className="opacity-60 font-sans font-medium">
        {MODALITY_LABEL[metadata.modality] ?? metadata.modality}
      </span>
      <span className="opacity-40">·</span>
      <span>{metadata.provider}</span>
      <span className="opacity-40">/</span>
      <span className="opacity-80">{metadata.model}</span>
      {metadata.failoverCount != null && metadata.failoverCount > 0 && (
        <span className="opacity-50 text-[10px]">(+{metadata.failoverCount} failover)</span>
      )}
    </span>
  );
}
