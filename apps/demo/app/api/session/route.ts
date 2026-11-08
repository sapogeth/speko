import { NextRequest, NextResponse } from "next/server";

const SPEKO_API_KEY = process.env.SPEKO_API_KEY;
const SPEKO_AGENT_ID = process.env.SPEKO_AGENT_ID;

export async function POST(req: NextRequest) {
  const body = (await req.json().catch(() => ({}))) as {
    language?: string;
    optimizeFor?: string;
  };

  // Demo mode — no API key configured
  if (!SPEKO_API_KEY) {
    return NextResponse.json(
      {
        transportToken: "demo-token",
        transportUrl: "wss://demo.speko.dev",
        sessionId: `demo-${Date.now()}`,
        _demo: true,
      },
      {
        headers: {
          "x-speko-provider": "deepgram",
          "x-speko-model": "nova-3",
        },
      }
    );
  }

  const res = await fetch("https://api.speko.dev/v1/sessions", {
    method: "POST",
    headers: {
      Authorization: `Bearer ${SPEKO_API_KEY}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      agentId: SPEKO_AGENT_ID,
      mode: "cascade",
      variables: {},
      intent: {
        language: body.language ?? "en-US",
        region: "global",
        optimizeFor: body.optimizeFor ?? "balanced",
      },
    }),
  });

  if (!res.ok) {
    const text = await res.text();
    return NextResponse.json(
      { error: "Failed to create session", detail: text },
      { status: res.status }
    );
  }

  const data = await res.json() as {
    transportToken: string;
    transportUrl: string;
    sessionId: string;
  };

  return NextResponse.json(data, {
    headers: {
      "x-speko-provider": res.headers.get("x-speko-provider") ?? "",
      "x-speko-model": res.headers.get("x-speko-model") ?? "",
    },
  });
}
