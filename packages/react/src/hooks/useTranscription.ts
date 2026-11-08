import { useCallback, useRef, useState } from "react";
import type {
  RoutingMetadata,
  TranscriptionOptions,
  UseTranscriptionReturn,
} from "../types.js";

export interface UseTranscriptionConfig extends Partial<TranscriptionOptions> {
  /**
   * Your backend endpoint that calls POST /v1/transcribe with SPEKO_API_KEY.
   * Never call the Speko API directly from the browser — the key must stay server-side.
   */
  endpoint: string;
}

export function useTranscription(
  config: UseTranscriptionConfig
): UseTranscriptionReturn {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const abortRef = useRef<AbortController | null>(null);

  const transcribe = useCallback(
    async (
      audio: Blob | ArrayBuffer,
      overrides?: Partial<TranscriptionOptions>
    ): Promise<{ text: string; routing: RoutingMetadata | null }> => {
      abortRef.current?.abort();
      const controller = new AbortController();
      abortRef.current = controller;

      setIsLoading(true);
      setError(null);

      try {
        const opts = { ...config, ...overrides };
        const audioBlob =
          audio instanceof Blob ? audio : new Blob([audio], { type: "audio/webm" });

        const formData = new FormData();
        formData.append("audio", audioBlob);
        if (opts.language) formData.append("language", opts.language);
        if (opts.region) formData.append("region", opts.region);
        if (opts.optimizeFor) formData.append("optimizeFor", opts.optimizeFor);

        const res = await fetch(config.endpoint, {
          method: "POST",
          body: formData,
          signal: controller.signal,
        });

        if (!res.ok) {
          throw new Error(`Transcription failed: ${res.status} ${res.statusText}`);
        }

        const data = (await res.json()) as { text: string; provider?: string; model?: string; failoverCount?: number };

        const routing: RoutingMetadata | null = data.provider
          ? {
              provider: data.provider,
              model: data.model ?? data.provider,
              modality: "stt",
              failoverCount: data.failoverCount,
            }
          : (res.headers.get("x-speko-provider")
              ? {
                  provider: res.headers.get("x-speko-provider")!,
                  model: res.headers.get("x-speko-model") ?? res.headers.get("x-speko-provider")!,
                  modality: "stt",
                }
              : null);

        return { text: data.text, routing };
      } catch (err) {
        if (err instanceof Error && err.name === "AbortError") {
          return { text: "", routing: null };
        }
        const e = err instanceof Error ? err : new Error(String(err));
        setError(e);
        throw e;
      } finally {
        setIsLoading(false);
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [config.endpoint, config.language, config.region, config.optimizeFor]
  );

  return { transcribe, isLoading, error };
}
