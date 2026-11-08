export { useVoiceConversation } from "./hooks/useVoiceConversation.js";
export { useTranscription, type UseTranscriptionConfig } from "./hooks/useTranscription.js";
export { useSynthesis, type UseSynthesisConfig } from "./hooks/useSynthesis.js";
export { useMicrophoneStream } from "./hooks/useMicrophoneStream.js";

export type {
  ConversationMessage,
  ConversationMode,
  ConversationStatus,
  OptimizeFor,
  ProviderConstraints,
  RoutingIntent,
  RoutingMetadata,
  SessionResponse,
  SpekoRegion,
  SynthesisOptions,
  TranscriptionOptions,
  UseTranscriptionReturn,
  UseSynthesisReturn,
  UseVoiceConversationReturn,
  VoiceConversationOptions,
} from "./types.js";
