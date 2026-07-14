import "@testing-library/jest-dom/vitest";
import { vi } from "vitest";

// JSDOM stubs for Radix UI components
Element.prototype.scrollIntoView = vi.fn();
Element.prototype.hasPointerCapture = vi.fn();

// JSDOM does not implement the deprecated contentEditable execCommand API.
// Provide a no-op default so tests can spy on it.
if (typeof document !== "undefined" && !document.execCommand) {
  document.execCommand = vi.fn(() => true);
}
