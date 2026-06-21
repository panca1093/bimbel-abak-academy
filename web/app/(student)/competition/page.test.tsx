import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import CompetitionPage from "./page";

let uiStore = { lang: "id" as "id" | "en" };

vi.mock("@/stores/ui", () => ({
  useUIStore: (selector: (s: typeof uiStore) => unknown) => selector(uiStore),
}));

describe("CompetitionPage", () => {
  it("renders title from i18n", () => {
    render(<CompetitionPage />);
    expect(
      screen.getByRole("heading", { name: /Kompetisi & Tryout/i })
    ).toBeInTheDocument();
  });

  it("shows maintenance message", () => {
    render(<CompetitionPage />);
    expect(
      screen.getByText(/Halaman ini sedang dalam pengembangan/i)
    ).toBeInTheDocument();
  });

  it("does not render package data", () => {
    render(<CompetitionPage />);
    expect(
      screen.queryByText("Try Out UTBK Gratis #12")
    ).not.toBeInTheDocument();
  });

  it("translates copy when language is en", () => {
    uiStore.lang = "en";
    render(<CompetitionPage />);
    expect(
      screen.getByText(/This page is under development/i)
    ).toBeInTheDocument();
  });
});
