import { describe, it, expect } from "vitest";
import { stripHtmlToPlainText } from "./rich-text";

describe("stripHtmlToPlainText", () => {
  it("strips tags and collapses whitespace around a script element", () => {
    expect(stripHtmlToPlainText("<b>Hi</b> <script>alert(1)</script>there")).toBe(
      "Hi there",
    );
  });

  it("returns empty string for undefined without throwing", () => {
    expect(stripHtmlToPlainText(undefined)).toBe("");
  });

  it("returns empty string for empty string without throwing", () => {
    expect(stripHtmlToPlainText("")).toBe("");
  });

  it("passes plain text through unchanged", () => {
    expect(stripHtmlToPlainText("Apa ibu kota Indonesia?")).toBe(
      "Apa ibu kota Indonesia?",
    );
  });

  it("collapses internal whitespace runs to a single space and trims", () => {
    expect(
      stripHtmlToPlainText("<div>  hello\n\n   world  </div>"),
    ).toBe("hello world");
  });
});
