import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";

import ExamSessionLayout from "./layout";

const routerReplace = vi.fn();

vi.mock("next/navigation", () => ({
  useRouter: () => ({ replace: routerReplace }),
}));

let authStore = { token: null as string | null };

vi.mock("@/stores/auth", () => ({
  useAuthStore: (selector: (s: typeof authStore) => unknown) => selector(authStore),
}));

describe("ExamSessionLayout", () => {
  beforeEach(() => {
    authStore = { token: null };
    routerReplace.mockReset();
  });

  it("shows loading placeholder while unhydrated", () => {
    render(
      <ExamSessionLayout>
        <div>Test content</div>
      </ExamSessionLayout>
    );
    expect(screen.getByText("Memuat…")).toBeInTheDocument();
  });

  it("redirects to login when token is falsy", async () => {
    authStore = { token: null };
    render(
      <ExamSessionLayout>
        <div>Test content</div>
      </ExamSessionLayout>
    );

    await waitFor(() => {
      expect(routerReplace).toHaveBeenCalledWith("/login");
    });
  });

  it("renders children when token is present and hydrated", async () => {
    authStore = { token: "valid-token" };
    render(
      <ExamSessionLayout>
        <div>Test content</div>
      </ExamSessionLayout>
    );

    await waitFor(() => {
      expect(screen.getByText("Test content")).toBeInTheDocument();
    });
  });

  it("does not render AppShell or navigation elements", async () => {
    authStore = { token: "valid-token" };
    render(
      <ExamSessionLayout>
        <div data-testid="child-content">Test content</div>
      </ExamSessionLayout>
    );

    await waitFor(() => {
      expect(screen.getByTestId("child-content")).toBeInTheDocument();
    });

    // Ensure no navigation sidebar or navigation-like elements are present
    // AppShell renders a sidebar with navigation links like "Beranda", "Kompetisi", etc.
    expect(screen.queryByText("Beranda")).not.toBeInTheDocument();
    expect(screen.queryByText("Kompetisi")).not.toBeInTheDocument();
    expect(screen.queryByText("Kursus Saya")).not.toBeInTheDocument();
  });
});
