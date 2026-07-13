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
  useGoogleLogin: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock("@/components/auth/GoogleSignInButton", () => ({
  GoogleSignInButton: ({ text }: { text: string }) => (
    <div data-testid="google-sign-in" data-text={text} />
  ),
}));

describe("LoginPage", () => {
  beforeEach(() => {
    pushMock.mockClear();
    mutateAsyncMock.mockClear();
    sessionStorage.clear();
  });

  it("renders the Google sign-in button above the password form", () => {
    const { container } = render(<LoginPage />);
    const form = container.querySelector("form");

    expect(screen.getByTestId("google-sign-in")).toHaveAttribute("data-text", "signin_with");
    expect(screen.getByTestId("google-sign-in").compareDocumentPosition(form!))
      .toBe(Node.DOCUMENT_POSITION_FOLLOWING);
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
