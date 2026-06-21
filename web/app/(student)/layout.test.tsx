import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, waitFor } from "@testing-library/react";
import StudentLayout from "./layout";

const replace = vi.fn();

vi.mock("next/navigation", () => ({
  useRouter: () => ({ replace }),
}));

vi.mock("next/script", () => ({
  default: () => null,
}));

vi.mock("@/components/shell/AppShell", () => ({
  AppShell: ({ role, children }: { role: string; children: React.ReactNode }) => (
    <div data-testid="shell" data-role={role}>
      {children}
    </div>
  ),
}));

let authStore = {
  token: null as string | null,
  user: null as { role?: string } | null,
};

vi.mock("@/stores/auth", () => ({
  useAuthStore: (selector: (s: typeof authStore) => unknown) => selector(authStore),
}));

describe("StudentLayout", () => {
  beforeEach(() => {
    replace.mockClear();
    authStore = { token: null, user: null };
  });

  it("redirects to /login when there is no token", async () => {
    render(<StudentLayout>protected</StudentLayout>);
    await waitFor(() => expect(replace).toHaveBeenCalledWith("/login"));
  });

  it("renders the shell for student role", async () => {
    authStore = { token: "t", user: { role: "student" } };
    const { getByTestId } = render(<StudentLayout>protected</StudentLayout>);
    await waitFor(() => expect(getByTestId("shell")).toBeInTheDocument());
    expect(getByTestId("shell")).toHaveAttribute("data-role", "student");
  });

  it("redirects to /admin when the role is an admin role", async () => {
    authStore = { token: "t", user: { role: "admin_store" } };
    render(<StudentLayout>protected</StudentLayout>);
    await waitFor(() => expect(replace).toHaveBeenCalledWith("/admin"));
  });
});
