export function AbakMark({ size = 28 }: { size?: number }) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 120 120"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      aria-label="abak academy"
      className="text-brand-600"
    >
      <circle cx="44" cy="34" r="15" fill="currentColor" />
      <path d="M22 104 Q22 64 44 64 Q66 64 66 104 Z" fill="currentColor" />
      <path d="M62 104 Q62 78 80 78 Q98 78 98 104 Z" fill="#1E978A" />
      <path d="M80 44 L96 51 L80 58 L64 51 Z" fill="#D99A2B" />
      <circle cx="80" cy="62" r="11" fill="#1E978A" />
      <rect x="79" y="44" width="2.5" height="9" fill="#D99A2B" />
    </svg>
  );
}
