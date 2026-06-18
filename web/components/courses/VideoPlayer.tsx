"use client";

interface VideoPlayerProps {
  videoRef?: string;
  title?: string;
}

function toYoutubeId(value?: string): string | null {
  if (!value) return null;
  const trimmed = value.trim();
  if (!trimmed) return null;
  if (/^https?:\/\//i.test(trimmed)) {
    try {
      const url = new URL(trimmed);
      if (/youtube\.com$/i.test(url.hostname)) {
        const v = url.searchParams.get("v");
        if (v) return v;
      }
      if (/youtu\.be$/i.test(url.hostname)) {
        const id = url.pathname.replace(/^\//, "");
        if (id) return id;
      }
      return null;
    } catch {
      return null;
    }
  }
  return trimmed;
}

export function VideoPlayer({ videoRef, title }: VideoPlayerProps) {
  const id = toYoutubeId(videoRef);
  return (
    <div
      className="overflow-hidden rounded-lg border border-line bg-ink-900"
      style={{ aspectRatio: "16 / 9" }}
    >
      {id ? (
        <iframe
          title={title ?? "Lesson video"}
          src={`https://www.youtube.com/embed/${encodeURIComponent(id)}?rel=0`}
          allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
          allowFullScreen
          className="block size-full border-0"
        />
      ) : (
        <div className="flex size-full items-center justify-center text-sm text-ink-300">
          Video belum tersedia.
        </div>
      )}
    </div>
  );
}