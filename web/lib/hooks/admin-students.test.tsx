import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, waitFor, act } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import {
  useAdminStudents,
  useRegisterStudent,
  useChangeStudentStatus,
  useReissueStudentCredentials,
  adminStudentsKeys,
} from "./admin-students";
import type {
  AdminStudent,
  StudentRegistrationInput,
  StudentRegistrationResult,
  StudentCredentials,
} from "@/lib/types";

const mockAuthFetch = vi.fn();

vi.mock("@/lib/api", () => ({
  authFetch: (...args: Parameters<typeof mockAuthFetch>) =>
    mockAuthFetch(...args),
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

describe("admin-students hooks", () => {
  beforeEach(() => {
    mockAuthFetch.mockReset();
  });

  afterEach(() => {
    vi.clearAllTimers();
  });

  it("useAdminStudents fetches GET /admin/students and returns paginated response", async () => {
    const students: AdminStudent[] = [
      {
        id: "st1",
        name: "Budi Santoso",
        username: "budi",
        nis: "12345",
        status: "active",
        created_at: "2026-01-01T00:00:00Z",
      },
    ];
    const response = { data: students, next_cursor: undefined };
    mockAuthFetch.mockResolvedValueOnce(response);

    const { wrapper } = wrapperFactory();
    const { result } = renderHook(() => useAdminStudents(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(mockAuthFetch).toHaveBeenCalledWith("/admin/students");
    expect(result.current.data).toEqual(response);
  });

  it("useAdminStudents passes status, q, cursor, limit as query params", async () => {
    mockAuthFetch.mockResolvedValueOnce({ data: [], next_cursor: undefined });

    const { wrapper } = wrapperFactory();
    renderHook(
      () =>
        useAdminStudents({
          status: "active",
          q: "budi",
          cursor: "cursor-1",
          limit: 10,
        }),
      { wrapper },
    );

    await waitFor(() =>
      expect(mockAuthFetch).toHaveBeenCalledWith(
        "/admin/students?status=active&q=budi&cursor=cursor-1&limit=10",
      ),
    );
  });

  it("useRegisterStudent posts to /admin/students and invalidates list", async () => {
    const input: StudentRegistrationInput = {
      name: "Siti Aisyah",
      nis: "67890",
      email: "siti@test.com",
    };
    const result: StudentRegistrationResult = {
      id: "st2",
      name: "Siti Aisyah",
      username: "siti",
      nis: "67890",
      status: "active",
      created_at: "2026-02-01T00:00:00Z",
      temp_password: "temp123",
    };
    mockAuthFetch.mockResolvedValueOnce(result);

    const { wrapper, queryClient } = wrapperFactory();
    const spy = vi.spyOn(queryClient, "invalidateQueries");
    const { result: hookResult } = renderHook(() => useRegisterStudent(), {
      wrapper,
    });

    await act(async () => {
      await hookResult.current.mutateAsync(input);
    });

    expect(mockAuthFetch).toHaveBeenCalledWith("/admin/students", {
      method: "POST",
      body: JSON.stringify(input),
    });
    expect(spy).toHaveBeenCalledWith({ queryKey: adminStudentsKeys.all });
  });

  it("useChangeStudentStatus patches /admin/students/:id and invalidates list", async () => {
    mockAuthFetch.mockResolvedValueOnce({ message: "status updated" });

    const { wrapper, queryClient } = wrapperFactory();
    const spy = vi.spyOn(queryClient, "invalidateQueries");
    const { result } = renderHook(() => useChangeStudentStatus(), { wrapper });

    await act(async () => {
      await result.current.mutateAsync({ id: "st1", status: "deactivated" });
    });

    expect(mockAuthFetch).toHaveBeenCalledWith("/admin/students/st1", {
      method: "PATCH",
      body: JSON.stringify({ status: "deactivated" }),
    });
    expect(spy).toHaveBeenCalledWith({ queryKey: adminStudentsKeys.all });
  });

  it("useReissueStudentCredentials fetches /admin/students/:id/credentials and invalidates list", async () => {
    const creds: StudentCredentials = {
      username: "budi",
      temp_password: "newPass789",
    };
    mockAuthFetch.mockResolvedValueOnce(creds);

    const { wrapper, queryClient } = wrapperFactory();
    const spy = vi.spyOn(queryClient, "invalidateQueries");
    const { result } = renderHook(() => useReissueStudentCredentials(), {
      wrapper,
    });

    await act(async () => {
      await result.current.mutateAsync("st1");
    });

    expect(mockAuthFetch).toHaveBeenCalledWith(
      "/admin/students/st1/credentials",
    );
    expect(spy).toHaveBeenCalledWith({ queryKey: adminStudentsKeys.all });
  });
});

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
