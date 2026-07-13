import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { PurchaseNotificationFeed } from "./PurchaseNotificationFeed";

const mockMutate = vi.fn();
const mockUseAdminNotifications = vi.fn();
const mockUseMarkNotificationRead = vi.fn(() => ({
  mutate: mockMutate,
  isPending: false,
}));

vi.mock("@/lib/hooks/admin-notifications", () => ({
  useAdminNotifications: (...args: Parameters<typeof mockUseAdminNotifications>) =>
    mockUseAdminNotifications(...args),
  useMarkNotificationRead: () => mockUseMarkNotificationRead(),
}));

vi.mock("@/lib/i18n", () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

const sampleNotif = {
  id: "notif-1",
  type: "order_confirmed",
  order_id: "ord-1",
  student_name: "Budi Santoso",
  amount: 150000,
  created_at: "2026-07-05T10:00:00Z",
  read: false,
};

const sampleNotifRead = {
  ...sampleNotif,
  id: "notif-2",
  student_name: "Siti Rahma",
  order_id: "ord-2",
  read: true,
};

function buildQueryResult(overrides: Record<string, unknown> = {}) {
  return {
    data: { data: [sampleNotif, sampleNotifRead], next_cursor: "10" },
    isSuccess: true,
    isFetching: false,
    ...overrides,
  };
}

describe("PurchaseNotificationFeed", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders notification list with student names and order ids", () => {
    mockUseAdminNotifications.mockReturnValue(buildQueryResult());

    render(<PurchaseNotificationFeed />);

    expect(screen.getByText("Budi Santoso")).toBeInTheDocument();
    expect(screen.getByText("Siti Rahma")).toBeInTheDocument();
    expect(screen.getByText(/ord-1/i)).toBeInTheDocument();
    expect(screen.getByText(/ord-2/i)).toBeInTheDocument();
  });

  it("renders formatted amount for each row", () => {
    mockUseAdminNotifications.mockReturnValue(buildQueryResult());

    render(<PurchaseNotificationFeed />);

    const amounts = screen.getAllByText("Rp150.000");
    expect(amounts).toHaveLength(2);
  });

  it("shows Load More button when next_cursor is present", () => {
    mockUseAdminNotifications.mockReturnValue(buildQueryResult());

    render(<PurchaseNotificationFeed />);

    expect(screen.getByRole("button", { name: /muat lebih banyak/i })).toBeInTheDocument();
  });

  it("hides Load More button when next_cursor is empty", () => {
    mockUseAdminNotifications.mockReturnValue(
      buildQueryResult({ data: { data: [sampleNotif], next_cursor: "" } })
    );

    render(<PurchaseNotificationFeed />);

    expect(screen.queryByRole("button", { name: /muat lebih banyak/i })).not.toBeInTheDocument();
  });

  it("shows loading indicator when fetching without data", () => {
    mockUseAdminNotifications.mockReturnValue(
      buildQueryResult({ isFetching: true, isSuccess: false, data: undefined })
    );

    render(<PurchaseNotificationFeed />);

    expect(screen.getByText("Memuat...")).toBeInTheDocument();
  });

  it("shows unread-only toggle and activates on click", () => {
    mockUseAdminNotifications.mockReturnValue(buildQueryResult());

    render(<PurchaseNotificationFeed />);

    const toggle = screen.getByRole("button", { name: /notification_unread_only/i });
    expect(toggle).toBeInTheDocument();

    // Click toggle and verify visual state changes
    fireEvent.click(toggle);

    // After click the unreadOnly state should be true, which sets bg-primary class
    expect(toggle.className).toContain("bg-primary");
  });

  it("calls mark-read mutation when an unread notification is clicked", async () => {
    mockUseAdminNotifications.mockReturnValue(buildQueryResult());

    render(<PurchaseNotificationFeed />);

    const notifElement = screen.getByText("Budi Santoso");
    fireEvent.click(notifElement);

    await waitFor(() => {
      expect(mockMutate).toHaveBeenCalledWith("notif-1");
    });
  });

  it("shows empty state when no notifications", () => {
    mockUseAdminNotifications.mockReturnValue(
      buildQueryResult({ data: { data: [], next_cursor: "" } })
    );

    render(<PurchaseNotificationFeed />);

    expect(screen.getByText("Belum ada notifikasi")).toBeInTheDocument();
  });

  it("renders empty state without crash when data.data is null", () => {
    mockUseAdminNotifications.mockReturnValue(
      buildQueryResult({ data: { data: null, next_cursor: "" } })
    );

    render(<PurchaseNotificationFeed />);

    expect(screen.getByText("Belum ada notifikasi")).toBeInTheDocument();
  });
});
