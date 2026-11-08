"use client";

interface WaveformProps {
  active: boolean;
  mode: "listening" | "speaking" | null;
  audioLevel?: number;
}

const BAR_COUNT = 20;

export function Waveform({ active, mode, audioLevel = 0 }: WaveformProps) {
  const color =
    mode === "speaking" ? "#6366f1" : mode === "listening" ? "#22c55e" : "#52525b";

  return (
    <div className="flex items-center justify-center gap-[3px] h-12">
      {Array.from({ length: BAR_COUNT }).map((_, i) => {
        const delay = `${(i * 0.04).toFixed(2)}s`;
        const baseHeight = active ? 24 + Math.sin((i / BAR_COUNT) * Math.PI) * 20 : 6;
        const levelBoost = active ? audioLevel * 32 * Math.sin((i / BAR_COUNT) * Math.PI) : 0;
        const height = Math.round(baseHeight + levelBoost);

        return (
          <div
            key={i}
            className={active ? "bar-animate" : ""}
            style={{
              width: 3,
              height,
              borderRadius: 2,
              backgroundColor: color,
              animationDelay: delay,
              transition: "height 0.1s ease, background-color 0.3s ease",
              opacity: active ? 0.9 : 0.3,
            }}
          />
        );
      })}
    </div>
  );
}
