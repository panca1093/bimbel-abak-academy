import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import LoginPage from "./page";
import { ApiError } from "@/lib/api";

const pushMock = vi.fn();

vi.mock("next/navigation", () => ({
  useRouter: () => ({ push: pushMock }),
}));

const mutateAsyncMock = vi.fn();

vi.mock("@/lib/hooks/auth", () => ({
  useLogin: () => ({ mutateAsync: mutateAsyncMock, isPending: false }),
}));

describe("LoginPage", () => {
  beforeEach(() => {
    pushMock.mockClear();
    mutateAsyncMock.mockClear();
    sessionStorage.clear();
  });

  it("stores the pending token and navigates to /otp on verification_pending", async () => {
    mutateAsyncMock.mockRejectedValue(
      new ApiError("verification_pending", "verify your email", 403, {
        pending_token: "tok-abc",
      }),
    );

    render(<LoginPage />);

    fireEvent.change(
      screen.getByLabelText(/email atau username|email or username/i, { selector: "input" }),
      { target: { value: "budi@example.com" } },
    );
    fireEvent.change(screen.getByLabelText(/kata sandi|password/i, { selector: "input" }), {
      target: { value: "secret123" },
    });
    fireEvent.click(screen.getByRole("button", { name: /masuk|sign in|login/i }));

    await waitFor(() => {
      expect(pushMock).toHaveBeenCalledWith("/otp?id=budi%40example.com");
    });
    expect(sessionStorage.getItem("abak-pending-token")).toBe("tok-abc");
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });
});
