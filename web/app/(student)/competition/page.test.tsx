import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import CompetitionPage from "./page";

vi.mock("@/components/shell/ComingSoon", () => ({
  ComingSoon: ({ title }: { title?: string }) => (
    <div data-testid="coming-soon" data-title={title}>
      ComingSoon
    </div>
  ),
}));

let authStore = {
  user: null as { name?: string } | null,
};

vi.mock("@/stores/auth", () => ({
  useAuthStore: (selector: (s: typeof authStore) => unknown) => selector(authStore),
}));

describe("CompetitionPage", () => {
  beforeEach(() => {
    authStore = { user: { name: "Budi" } };
  });

  it("renders the ComingSoon screen", () => {
    render(<CompetitionPage />);
    expect(screen.getByTestId("coming-soon")).toBeInTheDocument();
  });
});
