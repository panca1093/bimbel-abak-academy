export function stripHtmlToPlainText(html?: string | null): string {
  if (!html) return "";
  const d = document.createElement("div");
  d.innerHTML = html;
  d.querySelectorAll("script, style").forEach((el) => el.remove());
  return (d.textContent || "").replace(/\s+/g, " ").trim();
}
