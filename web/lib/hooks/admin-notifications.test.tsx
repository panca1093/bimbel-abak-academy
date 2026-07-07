import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor, act } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import {
  useAdminNotifications,
  useMarkNotificationRead,
  adminNotifsKeys,
} from "./admin-notifications";

const mockAuthFetch = vi.fn();

vi.mock("@/lib/api", () => ({
  authFetch: (...args: Parameters<typeof mockAuthFetch>) => mockAuthFetch(...args),
  ApiError: class extends Error {
    code: string;
    status: number;
    constructor(code: string, message: string, status: number) {
      super(message);
      this.code = code;
      this.status = status;
    }
  },
}));

vi.mock("@/stores/auth", () => ({
  useAuthStore: {
    getState: () => ({ token: "test-token" }),
  },
}));

const sampleNotification = {
  id: "notif-1",
  type: "order_confirmed",
  order_id: "ord-1",
  student_name: "Budi Santoso",
  amount: 150000,
  created_at: "2026-07-05T10:00:00Z",
  read: false,
};

function wrapperFactory() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return {
    wrapper: ({ children }: { children: React.ReactNode }) => (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    ),
    queryClient,
  };
}

describe("admin-notifications hooks", () => {
  beforeEach(() => {
    mockAuthFetch.mockReset();
  });

  describe("query keys", () => {
    it("adminNotifsKeys.all is stable", () => {
      expect(adminNotifsKeys.all).toEqual(["admin", "notifications"]);
    });

    it("adminNotifsKeys.list() returns default list key", () => {
      expect(adminNotifsKeys.list()).toEqual(["admin", "notifications", "list"]);
    });

    it("adminNotifsKeys.list({}) returns list key without filters", () => {
      expect(adminNotifsKeys.list({})).toEqual(["admin", "notifications", "list"]);
    });

    it("adminNotifsKeys.list({unreadOnly:true}) includes unreadOnly", () => {
      const key = adminNotifsKeys.list({ unreadOnly: true });
      expect(key).toEqual(["admin", "notifications", "list", "unread"]);
    });

    it("adminNotifsKeys.list({cursor:'5'}) includes cursor", () => {
      const key = adminNotifsKeys.list({ cursor: "5" });
      expect(key).toEqual(["admin", "notifications", "list", "5"]);
    });

    it("adminNotifsKeys.list({unreadOnly:true,cursor:'3'}) includes both", () => {
      const key = adminNotifsKeys.list({ unreadOnly: true, cursor: "3" });
      expect(key).toEqual(["admin", "notifications", "list", "unread", "3"]);
    });
  });

  describe("useAdminNotifications", () => {
    it("fetches GET /admin/notifications and returns data with next_cursor", async () => {
      const apiResponse = {
        data: [sampleNotification],
        next_cursor: "10",
      };
      mockAuthFetch.mockResolvedValueOnce(apiResponse);

      const { wrapper } = wrapperFactory();
      const { result } = renderHook(() => useAdminNotifications(), { wrapper });

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(mockAuthFetch).toHaveBeenCalledWith("/admin/notifications");
      expect(result.current.data).toEqual(apiResponse);
    });

    it("appends unread_only query param when unreadOnly is true", async () => {
      mockAuthFetch.mockResolvedValueOnce({ data: [], next_cursor: "" });

      const { wrapper } = wrapperFactory();
      renderHook(() => useAdminNotifications({ unreadOnly: true }), { wrapper });

      await waitFor(() =>
        expect(mockAuthFetch).toHaveBeenCalledWith("/admin/notifications?unread_only=true")
      );
    });

    it("appends cursor query param when provided", async () => {
      mockAuthFetch.mockResolvedValueOnce({ data: [], next_cursor: "" });

      const { wrapper } = wrapperFactory();
      renderHook(() => useAdminNotifications({ cursor: "5" }), { wrapper });

      await waitFor(() =>
        expect(mockAuthFetch).toHaveBeenCalledWith("/admin/notifications?cursor=5")
      );
    });

    it("appends both cursor and unread_only query params", async () => {
      mockAuthFetch.mockResolvedValueOnce({ data: [], next_cursor: "" });

      const { wrapper } = wrapperFactory();
      renderHook(() => useAdminNotifications({ unreadOnly: true, cursor: "3" }), { wrapper });

      await waitFor(() =>
        expect(mockAuthFetch).toHaveBeenCalledWith(
          "/admin/notifications?unread_only=true&cursor=3"
        )
      );
    });
  });

  describe("useMarkNotificationRead", () => {
    it("calls PATCH /admin/notifications/:id/read and invalidates query", async () => {
      mockAuthFetch.mockResolvedValueOnce({ message: "notification marked read" });

      const { wrapper, queryClient } = wrapperFactory();
      const spy = vi.spyOn(queryClient, "invalidateQueries");
      const { result } = renderHook(() => useMarkNotificationRead(), { wrapper });

      await act(async () => {
        await result.current.mutateAsync("notif-1");
      });

      expect(mockAuthFetch).toHaveBeenCalledWith("/admin/notifications/notif-1/read", {
        method: "PATCH",
      });
      expect(spy).toHaveBeenCalledWith({ queryKey: adminNotifsKeys.all });
    });
  });
});
