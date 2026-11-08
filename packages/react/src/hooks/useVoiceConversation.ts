import { useCallback, useEffect, useRef, useState } from "react";
import type {
  ConversationMessage,
  ConversationMode,
  ConversationStatus,
  RoutingMetadata,
  UseVoiceConversationReturn,
  VoiceConversationOptions,
} from "../types.js";

type VoiceConversationInstance = {
  getId: () => string;
  isOpen: () => boolean;
  setMicMuted: (muted: boolean) => Promise<void>;
  setVolume: (volume: number) => void;
  sendChatMessage: (text: string) => Promise<void>;
  endSession: () => Promise<void>;
};

export function useVoiceConversation(
  options: VoiceConversationOptions
): UseVoiceConversationReturn {
  const [status, setStatus] = useState<ConversationStatus>("idle");
  const [mode, setMode] = useState<ConversationMode | null>(null);
  const [messages, setMessages] = useState<ConversationMessage[]>([]);
  const [isMuted, setIsMuted] = useState(false);
  const [routing, setRouting] = useState<RoutingMetadata | null>(null);
  const [error, setError] = useState<Error | null>(null);

  const conversationRef = useRef<VoiceConversationInstance | null>(null);
  const optionsRef = useRef(options);
  optionsRef.current = options;

  useEffect(() => {
    return () => {
      conversationRef.current?.endSession().catch(() => {});
    };
  }, []);

  const start = useCallback(async () => {
    if (status === "connecting" || status === "connected") return;

    setError(null);
    setStatus("connecting");
    setMessages([]);

    try {
      const res = await fetch(optionsRef.current.sessionEndpoint, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(optionsRef.current.sessionPayload ?? {}),
      });

      if (!res.ok) {
        throw new Error(`Session creation failed: ${res.status} ${res.statusText}`);
      }

      const { transportToken, transportUrl } = await res.json() as {
        transportToken: string;
        transportUrl: string;
      };

      // Dynamic import keeps @spekoai/client out of SSR bundles
      const { VoiceConversation } = await import("@spekoai/client");

      const conv = await VoiceConversation.create({
        transportToken,
        transportUrl,
        audioConstraints: optionsRef.current.audioConstraints,

        onConnect: () => {
          setStatus("connected");
        },

        onDisconnect: () => {
          setStatus("disconnected");
          setMode(null);
          conversationRef.current = null;
        },

        onMessage: (message: ConversationMessage) => {
          setMessages((prev) => {
            if (!message.isFinal) {
              const lastIdx = prev.findLastIndex(
                (m) => m.source === message.source && !m.isFinal
              );
              if (lastIdx !== -1) {
                const next = [...prev];
                next[lastIdx] = message;
                return next;
              }
            }
            return [...prev, message];
          });
          optionsRef.current.onMessage?.(message);
        },

        onStatusChange: (s: ConversationStatus) => {
          setStatus(s);
          optionsRef.current.onStatusChange?.(s);
        },

        onModeChange: (m: ConversationMode) => {
          setMode(m);
          optionsRef.current.onModeChange?.(m);
        },

        onError: (err: Error) => {
          setError(err);
          optionsRef.current.onError?.(err);
        },
      });

      conversationRef.current = conv;

      // Extract routing header from the session response
      const provider = res.headers.get("x-speko-provider");
      const model = res.headers.get("x-speko-model");
      if (provider && model) {
        const meta: RoutingMetadata = { provider, model, modality: "s2s" };
        setRouting(meta);
        optionsRef.current.onRoutingMetadata?.(meta);
      }
    } catch (err) {
      const e = err instanceof Error ? err : new Error(String(err));
      setError(e);
      setStatus("idle");
      optionsRef.current.onError?.(e);
    }
  }, [status]);

  const end = useCallback(async () => {
    if (!conversationRef.current) return;
    setStatus("disconnecting");
    await conversationRef.current.endSession();
  }, []);

  const setMuted = useCallback(async (muted: boolean) => {
    if (!conversationRef.current) return;
    await conversationRef.current.setMicMuted(muted);
    setIsMuted(muted);
  }, []);

  const setVolume = useCallback((volume: number) => {
    conversationRef.current?.setVolume(Math.max(0, Math.min(1, volume)));
  }, []);

  const sendMessage = useCallback((text: string) => {
    conversationRef.current?.sendChatMessage(text).catch((err: unknown) => {
      const e = err instanceof Error ? err : new Error(String(err));
      setError(e);
    });
  }, []);

  return {
    status,
    mode,
    messages,
    isMuted,
    routing,
    error,
    start,
    end,
    setMuted,
    setVolume,
    sendMessage,
  };
}
