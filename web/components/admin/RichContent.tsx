"use client";

import { useEffect, useRef } from "react";
import DOMPurify from "dompurify";
import renderMathInElement from "katex/contrib/auto-render";
import "katex/dist/katex.min.css";

interface RichContentProps {
  html: string;
  className?: string;
}

// Mirrors the backend's questionBodyPolicy allowlist (exam.go). Sanitizing
// here too — not just at write time — matters because rows persisted before
// write-time sanitization was added are untrusted and still render through
// this component.
const ALLOWED_TAGS = ["b", "i", "u", "ul", "ol", "li", "sup", "sub", "img"];
const ALLOWED_ATTR = ["src", "alt", "style"];

export function RichContent({ html, className }: RichContentProps) {
  const containerRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    const el = containerRef.current;
    if (!el) return;
    el.innerHTML = DOMPurify.sanitize(html, { ALLOWED_TAGS, ALLOWED_ATTR });
    renderMathInElement(el, {
      delimiters: [
        { left: "\\(", right: "\\)", display: false },
        { left: "\\[", right: "\\]", display: true },
      ],
      throwOnError: false,
    });
  }, [html]);

  return <div ref={containerRef} data-rich-content className={className} />;
}
