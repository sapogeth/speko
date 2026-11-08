export type ConversationStatus =
  | "idle"
  | "connecting"
  | "connected"
  | "disconnecting"
  | "disconnected";

export type ConversationMode = "listening" | "speaking";

export type OptimizeFor = "balanced" | "accuracy" | "latency" | "cost";

export type SpekoRegion =
  | "global"
  | "us-east4"
  | "europe-west3"
  | "asia-southeast1";

export interface RoutingIntent {
  language: string;
  region?: SpekoRegion;
  optimizeFor?: OptimizeFor;
}

export interface ProviderConstraints {
  stt?: string[];
  tts?: string[];
  llm?: string[];
}

export interface ConversationMessage {
  source: "user" | "agent";
  text: string;
  isFinal: boolean;
  segmentId?: string;
}

export interface RoutingMetadata {
  provider: string;
  model: string;
  modality: "stt" | "tts" | "llm" | "s2s";
  failoverCount?: number;
  scoreBreakdown?: {
    quality: number;
    latency: number;
    cost: number;
    composite: number;
  };
}

export interface SessionResponse {
  transportToken: string;
  transportUrl: string;
  sessionId: string;
}

export interface VoiceConversationOptions {
  /** Endpoint on your server that returns { transportToken, transportUrl } */
  sessionEndpoint: string;
  /** Override the session POST body */
  sessionPayload?: Record<string, unknown>;
  /** Audio device constraints */
  audioConstraints?: {
    echoCancellation?: boolean;
    noiseSuppression?: boolean;
    autoGainControl?: boolean;
  };
  onMessage?: (message: ConversationMessage) => void;
  onStatusChange?: (status: ConversationStatus) => void;
  onModeChange?: (mode: ConversationMode) => void;
  onRoutingMetadata?: (metadata: RoutingMetadata) => void;
  onError?: (error: Error) => void;
}

export interface UseVoiceConversationReturn {
  status: ConversationStatus;
  mode: ConversationMode | null;
  messages: ConversationMessage[];
  isMuted: boolean;
  routing: RoutingMetadata | null;
  error: Error | null;
  start: () => Promise<void>;
  end: () => Promise<void>;
  setMuted: (muted: boolean) => Promise<void>;
  setVolume: (volume: number) => void;
  sendMessage: (text: string) => void;
}

export interface TranscriptionOptions {
  language: string;
  region?: SpekoRegion;
  optimizeFor?: OptimizeFor;
  constraints?: { allowedProviders?: ProviderConstraints };
}

export interface UseTranscriptionReturn {
  transcribe: (audio: Blob | ArrayBuffer) => Promise<{ text: string; routing: RoutingMetadata | null }>;
  isLoading: boolean;
  error: Error | null;
}

export interface SynthesisOptions {
  language: string;
  voice?: string;
  speed?: number;
  outputFormat?: "wav" | "pcm" | "mp3";
  optimizeFor?: OptimizeFor;
}

export interface UseSynthesisReturn {
  synthesize: (text: string) => Promise<{ audioUrl: string; routing: RoutingMetadata | null }>;
  isLoading: boolean;
  error: Error | null;
}
