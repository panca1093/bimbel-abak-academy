import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, waitFor, act } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import {
  useAdminAnnouncements,
  useCreateAnnouncement,
  useUpdateAnnouncement,
  useDeleteAnnouncement,
  useSendAnnouncement,
  adminAnnouncementKeys,
} from "./admin-announcements";

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

const draftAnnouncement = {
  id: "ann-1",
  title: "Test Announcement",
  message: "Test message",
  type: "announcement",
  recipients: "all",
  status: "draft",
  scheduled_at: null,
  sent_at: null,
  recipient_count: null,
  created_by: "admin-u1",
  created_at: "2026-07-06T10:00:00Z",
  updated_at: "2026-07-06T10:00:00Z",
};

const sentAnnouncement = {
  ...draftAnnouncement,
  id: "ann-sent-1",
  status: "sent",
  sent_at: "2026-07-06T11:00:00Z",
  recipient_count: 10,
};

const scheduledAnnouncement = {
  ...draftAnnouncement,
  id: "ann-sched-1",
  status: "scheduled",
  scheduled_at: "2026-07-07T10:00:00Z",
};

const announcementList = [draftAnnouncement, sentAnnouncement, scheduledAnnouncement];

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

describe("admin-announcements hooks", () => {
  beforeEach(() => {
    mockAuthFetch.mockReset();
  });

  describe("query keys", () => {
    it("adminAnnouncementKeys.all is stable", () => {
      expect(adminAnnouncementKeys.all).toEqual(["admin", "announcements"]);
    });

    it("adminAnnouncementKeys.list() returns default list key", () => {
      expect(adminAnnouncementKeys.list()).toEqual(["admin", "announcements", "list"]);
    });
  });

  describe("useAdminAnnouncements", () => {
    it("fetches GET /admin/notifications and returns data array", async () => {
      mockAuthFetch.mockResolvedValueOnce({ data: announcementList });

      const { wrapper } = wrapperFactory();
      const { result } = renderHook(() => useAdminAnnouncements(), { wrapper });

      await waitFor(() => expect(result.current.isSuccess).toBe(true));

      expect(mockAuthFetch).toHaveBeenCalledWith("/admin/notifications");
      expect(result.current.data).toEqual(announcementList);
    });
  });

  describe("useCreateAnnouncement", () => {
    it("creates a draft announcement and invalidates list", async () => {
      mockAuthFetch.mockResolvedValueOnce(draftAnnouncement);

      const { wrapper, queryClient } = wrapperFactory();
      const spy = vi.spyOn(queryClient, "invalidateQueries");
      const { result } = renderHook(() => useCreateAnnouncement(), { wrapper });

      const input = {
        title: "Test Announcement",
        message: "Test message",
        type: "announcement",
        recipients: "all",
        status: "draft",
      };

      await act(async () => {
        await result.current.mutateAsync(input);
      });

      expect(mockAuthFetch).toHaveBeenCalledWith("/admin/notifications", {
        method: "POST",
        body: JSON.stringify(input),
      });
      expect(spy).toHaveBeenCalledWith({ queryKey: adminAnnouncementKeys.list() });
    });

    it("creates a scheduled announcement with scheduled_at", async () => {
      mockAuthFetch.mockResolvedValueOnce(scheduledAnnouncement);

      const { wrapper, queryClient } = wrapperFactory();
      const spy = vi.spyOn(queryClient, "invalidateQueries");
      const { result } = renderHook(() => useCreateAnnouncement(), { wrapper });

      const input = {
        title: "Scheduled Announcement",
        message: "Test",
        type: "announcement",
        recipients: "students",
        status: "scheduled",
        scheduled_at: "2026-07-07T10:00:00Z",
      };

      await act(async () => {
        const res = await result.current.mutateAsync(input);
      });

      expect(mockAuthFetch).toHaveBeenCalledWith("/admin/notifications", {
        method: "POST",
        body: JSON.stringify(input),
      });
      expect(spy).toHaveBeenCalledWith({ queryKey: adminAnnouncementKeys.list() });
    });

    it("creates a sent announcement (send now)", async () => {
      const sentNow = { ...draftAnnouncement, id: "ann-sent-2", status: "sent" };
      mockAuthFetch.mockResolvedValueOnce(sentNow);

      const { wrapper, queryClient } = wrapperFactory();
      const spy = vi.spyOn(queryClient, "invalidateQueries");
      const { result } = renderHook(() => useCreateAnnouncement(), { wrapper });

      const input = {
        title: "Send Now",
        message: "Test",
        type: "promo",
        recipients: "all",
        status: "sent",
      };

      await act(async () => {
        const res = await result.current.mutateAsync(input);
      });

      expect(mockAuthFetch).toHaveBeenCalledWith("/admin/notifications", {
        method: "POST",
        body: JSON.stringify(input),
      });
      expect(spy).toHaveBeenCalledWith({ queryKey: adminAnnouncementKeys.list() });
    });
  });

  describe("useUpdateAnnouncement", () => {
    it("updates a draft announcement and invalidates list", async () => {
      const updated = { ...draftAnnouncement, title: "Updated Title" };
      mockAuthFetch.mockResolvedValueOnce(updated);

      const { wrapper, queryClient } = wrapperFactory();
      const spy = vi.spyOn(queryClient, "invalidateQueries");
      const { result } = renderHook(() => useUpdateAnnouncement(), { wrapper });

      await act(async () => {
        await result.current.mutateAsync({
          id: "ann-1",
          input: { title: "Updated Title" },
        });
      });

      expect(mockAuthFetch).toHaveBeenCalledWith("/admin/notifications/ann-1", {
        method: "PATCH",
        body: JSON.stringify({ title: "Updated Title" }),
      });
      expect(spy).toHaveBeenCalledWith({ queryKey: adminAnnouncementKeys.list() });
    });

    it("rejects update on sent announcement with API error", async () => {
      const { ApiError: MockApiError } = await vi.importActual<{ ApiError: typeof Error }>("@/lib/api");
      // Use the mocked ApiError
      mockAuthFetch.mockRejectedValueOnce(
        new (class extends Error {
          code = "announcement_immutable";
          status = 409;
          constructor() {
            super("announcement is immutable: already sent");
          }
        })()
      );

      const { wrapper } = wrapperFactory();
      const { result } = renderHook(() => useUpdateAnnouncement(), { wrapper });

      await act(async () => {
        try {
          await result.current.mutateAsync({
            id: "ann-sent-1",
            input: { title: "Updated" },
          });
          expect.unreachable("should have thrown");
        } catch (e: any) {
          expect(e.status).toBe(409);
          expect(e.code).toBe("announcement_immutable");
        }
      });
    });
  });

  describe("useDeleteAnnouncement", () => {
    it("deletes a draft announcement and invalidates list", async () => {
      mockAuthFetch.mockResolvedValueOnce({ message: "announcement deleted" });

      const { wrapper, queryClient } = wrapperFactory();
      const spy = vi.spyOn(queryClient, "invalidateQueries");
      const { result } = renderHook(() => useDeleteAnnouncement(), { wrapper });

      await act(async () => {
        await result.current.mutateAsync("ann-1");
      });

      expect(mockAuthFetch).toHaveBeenCalledWith("/admin/notifications/ann-1", {
        method: "DELETE",
      });
      expect(spy).toHaveBeenCalledWith({ queryKey: adminAnnouncementKeys.list() });
    });

    it("rejects delete on sent announcement with API error", async () => {
      mockAuthFetch.mockRejectedValueOnce(
        new (class extends Error {
          code = "announcement_immutable";
          status = 409;
          constructor() {
            super("announcement is immutable: already sent");
          }
        })()
      );

      const { wrapper } = wrapperFactory();
      const { result } = renderHook(() => useDeleteAnnouncement(), { wrapper });

      await act(async () => {
        try {
          await result.current.mutateAsync("ann-sent-1");
          expect.unreachable("should have thrown");
        } catch (e: any) {
          expect(e.status).toBe(409);
          expect(e.code).toBe("announcement_immutable");
        }
      });
    });
  });

  describe("useSendAnnouncement", () => {
    it("sends a draft announcement and invalidates list", async () => {
      const sent = { ...draftAnnouncement, status: "sent", sent_at: "2026-07-06T12:00:00Z", recipient_count: 15 };
      mockAuthFetch.mockResolvedValueOnce(sent);

      const { wrapper, queryClient } = wrapperFactory();
      const spy = vi.spyOn(queryClient, "invalidateQueries");
      const { result } = renderHook(() => useSendAnnouncement(), { wrapper });

      await act(async () => {
        await result.current.mutateAsync("ann-1");
      });

      expect(mockAuthFetch).toHaveBeenCalledWith("/admin/notifications/ann-1/send", {
        method: "POST",
      });
      expect(spy).toHaveBeenCalledWith({ queryKey: adminAnnouncementKeys.list() });
    });

    it("rejects send on already-sent announcement with API error", async () => {
      mockAuthFetch.mockRejectedValueOnce(
        new (class extends Error {
          code = "announcement_immutable";
          status = 409;
          constructor() {
            super("announcement is immutable: already sent");
          }
        })()
      );

      const { wrapper } = wrapperFactory();
      const { result } = renderHook(() => useSendAnnouncement(), { wrapper });

      await act(async () => {
        try {
          await result.current.mutateAsync("ann-sent-1");
          expect.unreachable("should have thrown");
        } catch (e: any) {
          expect(e.status).toBe(409);
          expect(e.code).toBe("announcement_immutable");
        }
      });
    });
  });

  describe("list query invalidation on all mutations", () => {
    it("useUpdateAnnouncement invalidates the list query", async () => {
      mockAuthFetch.mockResolvedValueOnce({ ...draftAnnouncement });
      mockAuthFetch.mockResolvedValueOnce(draftAnnouncement);
      mockAuthFetch.mockResolvedValueOnce({ ...draftAnnouncement, title: "Updated" });

      const { wrapper, queryClient } = wrapperFactory();
      const spy = vi.spyOn(queryClient, "invalidateQueries");

      const { result: listResult } = renderHook(() => useAdminAnnouncements(), { wrapper });
      await waitFor(() => expect(listResult.current.isSuccess).toBe(true));

      const { result: updateResult } = renderHook(() => useUpdateAnnouncement(), { wrapper });
      await act(async () => {
        await updateResult.current.mutateAsync({
          id: "ann-1",
          input: { title: "Updated" },
        });
      });

      expect(spy).toHaveBeenCalledWith({ queryKey: adminAnnouncementKeys.list() });
    });

    it("useDeleteAnnouncement invalidates the list query", async () => {
      mockAuthFetch.mockResolvedValueOnce({ data: [draftAnnouncement] });
      mockAuthFetch.mockResolvedValueOnce({ message: "announcement deleted" });

      const { wrapper, queryClient } = wrapperFactory();
      const spy = vi.spyOn(queryClient, "invalidateQueries");

      const { result: listResult } = renderHook(() => useAdminAnnouncements(), { wrapper });
      await waitFor(() => expect(listResult.current.isSuccess).toBe(true));

      const { result: deleteResult } = renderHook(() => useDeleteAnnouncement(), { wrapper });
      await act(async () => {
        await deleteResult.current.mutateAsync("ann-1");
      });

      expect(spy).toHaveBeenCalledWith({ queryKey: adminAnnouncementKeys.list() });
    });

    it("useSendAnnouncement invalidates the list query", async () => {
      const sent = { ...draftAnnouncement, status: "sent" };
      mockAuthFetch.mockResolvedValueOnce({ data: [draftAnnouncement] });
      mockAuthFetch.mockResolvedValueOnce(sent);

      const { wrapper, queryClient } = wrapperFactory();
      const spy = vi.spyOn(queryClient, "invalidateQueries");

      const { result: listResult } = renderHook(() => useAdminAnnouncements(), { wrapper });
      await waitFor(() => expect(listResult.current.isSuccess).toBe(true));

      const { result: sendResult } = renderHook(() => useSendAnnouncement(), { wrapper });
      await act(async () => {
        await sendResult.current.mutateAsync("ann-1");
      });

      expect(spy).toHaveBeenCalledWith({ queryKey: adminAnnouncementKeys.list() });
    });
  });
});
