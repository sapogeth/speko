import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "@spekoai/react — React hooks for Speko Voice AI",
  description:
    "Build multilingual voice agents in React with zero boilerplate. useVoiceConversation, useTranscription, useSynthesis — routing-aware hooks for Speko.",
  openGraph: {
    title: "@spekoai/react",
    description: "React hooks for Speko Voice AI",
    type: "website",
  },
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" className="h-full">
      <body className="min-h-full">{children}</body>
    </html>
  );
}
