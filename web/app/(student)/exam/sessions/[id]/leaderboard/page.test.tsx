import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { ApiError } from "@/lib/api";

import SessionLeaderboardPage from "./page";

let lbState = {
  data: null as { data: { rank: number; student_id: string; student_name: string; score: number }[] } | null,
  isLoading: true,
  isError: false,
  error: null as Error | null,
  refetch: vi.fn(),
};

vi.mock("@/lib/hooks/exam", () => ({
  useSessionLeaderboard: () => lbState,
}));

const sampleEntries = [
  { rank: 1, student_id: "s1", student_name: "Budi Santoso", score: 95 },
  { rank: 2, student_id: "s2", student_name: "Siti Aminah", score: 88 },
  { rank: 3, student_id: "s3", student_name: "Agus Wijaya", score: 82 },
];

describe("SessionLeaderboardPage", () => {
  beforeEach(() => {
    lbState = {
      data: { data: sampleEntries },
      isLoading: false,
      isError: false,
      error: null,
      refetch: vi.fn(),
    };
  });

  it("renders ranked rows on success", async () => {
    render(<SessionLeaderboardPage />);

    await waitFor(() => {
      expect(screen.getByText("Budi Santoso")).toBeInTheDocument();
    });
    expect(screen.getByText("Siti Aminah")).toBeInTheDocument();
    expect(screen.getByText("Agus Wijaya")).toBeInTheDocument();
    expect(screen.getByText("95")).toBeInTheDocument();
  });

  it("renders the not-available message on 403 leaderboard_not_available", () => {
    lbState = {
      data: null,
      isLoading: false,
      isError: true,
      error: new ApiError("leaderboard_not_available", "Peringkat belum tersedia", 403),
      refetch: vi.fn(),
    };

    render(<SessionLeaderboardPage />);

    expect(screen.getByText("Peringkat belum tersedia")).toBeInTheDocument();
  });
});
