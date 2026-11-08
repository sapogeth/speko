import { useCallback, useRef, useState } from "react";
import type {
  RoutingMetadata,
  SynthesisOptions,
  UseSynthesisReturn,
} from "../types.js";

export interface UseSynthesisConfig extends Partial<SynthesisOptions> {
  /**
   * Your backend endpoint that calls POST /v1/synthesize with SPEKO_API_KEY.
   * Never call the Speko API directly from the browser — the key must stay server-side.
   */
  endpoint: string;
}

export function useSynthesis(config: UseSynthesisConfig): UseSynthesisReturn {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const abortRef = useRef<AbortController | null>(null);

  const synthesize = useCallback(
    async (
      text: string,
      overrides?: Partial<SynthesisOptions>
    ): Promise<{ audioUrl: string; routing: RoutingMetadata | null }> => {
      abortRef.current?.abort();
      const controller = new AbortController();
      abortRef.current = controller;

      setIsLoading(true);
      setError(null);

      try {
        const opts = { ...config, ...overrides };

        const res = await fetch(config.endpoint, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            text,
            language: opts.language ?? "en",
            voice: opts.voice,
            speed: opts.speed ?? 1.0,
            outputFormat: opts.outputFormat ?? "wav",
            optimizeFor: opts.optimizeFor,
          }),
          signal: controller.signal,
        });

        if (!res.ok) {
          throw new Error(`Synthesis failed: ${res.status} ${res.statusText}`);
        }

        const audioBuffer = await res.arrayBuffer();
        const mimeType =
          opts.outputFormat === "mp3" ? "audio/mpeg" : "audio/wav";
        const audioUrl = URL.createObjectURL(
          new Blob([audioBuffer], { type: mimeType })
        );

        const provider = res.headers.get("x-speko-provider");
        const model = res.headers.get("x-speko-model");
        const routing: RoutingMetadata | null = provider
          ? { provider, model: model ?? provider, modality: "tts" }
          : null;

        return { audioUrl, routing };
      } catch (err) {
        if (err instanceof Error && err.name === "AbortError") {
          return { audioUrl: "", routing: null };
        }
        const e = err instanceof Error ? err : new Error(String(err));
        setError(e);
        throw e;
      } finally {
        setIsLoading(false);
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [config.endpoint, config.language, config.voice, config.outputFormat, config.optimizeFor]
  );

  return { synthesize, isLoading, error };
}
