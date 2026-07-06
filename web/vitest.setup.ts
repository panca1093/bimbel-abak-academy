import "@testing-library/jest-dom/vitest";
import { vi } from "vitest";

// JSDOM stubs for Radix UI components
Element.prototype.scrollIntoView = vi.fn();
Element.prototype.hasPointerCapture = vi.fn();
