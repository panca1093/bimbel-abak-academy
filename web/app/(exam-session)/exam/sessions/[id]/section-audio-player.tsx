"use client";

import { useState } from "react";

interface SectionAudioPlayerProps {
  audioUrl: string;
  playLimit?: number | null;
  testId?: string;
}

export function SectionAudioPlayer({ audioUrl, playLimit, testId = "section-audio-player" }: SectionAudioPlayerProps) {
  const [playCount, setPlayCount] = useState(0);
  const limitReached = playLimit != null && playCount >= playLimit;

  return (
    <div className="mb-4 rounded-lg border border-line bg-background p-3">
      <audio
        data-testid={testId}
        src={audioUrl}
        controls
        onPlay={() => setPlayCount((c) => c + 1)}
        className="w-full"
      />
      {limitReached && (
        <p className="mt-1 text-xs text-warning">
          Audio play limit reached
        </p>
      )}
    </div>
  );
}
