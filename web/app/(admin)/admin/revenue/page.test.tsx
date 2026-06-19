import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import RevenuePage from "./page";
import type { AdminRevenue } from "@/lib/types";

let revenueState = {
  data: null as AdminRevenue | null,
  isLoading: true,
  isError: false,
  error: null as Error | null,
};

vi.mock("@/lib/hooks/admin-revenue", () => ({
  useAdminRevenue: () => revenueState,
}));

describe("RevenuePage", () => {
  beforeEach(() => {
    revenueState = {
      data: {
        total: 2_500_000,
        by_type: {
          book: { total: 1_000_000, count: 10 },
          course: { total: 1_500_000, count: 3 },
        },
      },
      isLoading: false,
      isError: false,
      error: null,
    };
  });

  it("renders KPI tiles from revenue data", async () => {
    render(<RevenuePage />);

    await waitFor(() => {
      expect(screen.getByText("Rp2.500.000")).toBeInTheDocument();
      expect(screen.getByText("13")).toBeInTheDocument();
      expect(screen.getByText("Rp192.308")).toBeInTheDocument();
    });
  });

  it("renders revenue-by-type bars", async () => {
    render(<RevenuePage />);

    await waitFor(() => {
      expect(screen.getByText(/book/i)).toBeInTheDocument();
      expect(screen.getByText(/course/i)).toBeInTheDocument();
    });

    expect(screen.getByText((_, node) => node?.textContent === "Rp1.000.000 · 10 orders")).toBeInTheDocument();
    expect(screen.getByText((_, node) => node?.textContent === "Rp1.500.000 · 3 orders")).toBeInTheDocument();
  });

  it("renders top products table headers", async () => {
    render(<RevenuePage />);

    await waitFor(() => {
      expect(screen.getByText("Top products")).toBeInTheDocument();
    });

    const headers = screen.getAllByRole("columnheader");
    expect(headers.map((h) => h.textContent)).toEqual(expect.arrayContaining(["Product", "Orders", "Revenue"]));
  });

  it("shows loading skeletons while loading", () => {
    revenueState = {
      data: null,
      isLoading: true,
      isError: false,
      error: null,
    };

    render(<RevenuePage />);

    expect(document.querySelectorAll("[data-slot='skeleton']").length).toBeGreaterThan(0);
  });

  it("surfaces an API error as inline error text", async () => {
    revenueState = {
      data: null,
      isLoading: false,
      isError: true,
      error: new Error("gagal memuat"),
    };

    render(<RevenuePage />);

    await waitFor(() => {
      expect(screen.getByText(/gagal memuat/i)).toBeInTheDocument();
    });
  });
});
