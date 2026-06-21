import { describe, it, expect, vi, beforeAll } from "vitest";
import { render } from "@testing-library/react";
import React from "react";
import { AppShell } from "./AppShell";

const replace = vi.fn();

beforeAll(() => {
  Object.defineProperty(window, "matchMedia", {
    writable: true,
    value: vi.fn().mockImplementation((query: string) => ({
      matches: false,
      media: query,
      onchange: null,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
    })),
  });
});

vi.mock("next/navigation", () => ({
  useRouter: () => ({ replace }),
  usePathname: () => "/dashboard",
}));

vi.mock("@/lib/hooks/auth", () => ({
  useLogout: () => ({ mutate: vi.fn() }),
}));

vi.mock("@/lib/i18n", () => ({
  useTranslation: () => ({ t: (k: string) => k, lang: "id" }),
}));

vi.mock("./AppSidebar", () => ({
  AppSidebar: () => <div data-testid="sidebar" />,
}));

vi.mock("./AppHeader", () => ({
  AppHeader: ({ onMenuClick }: { onMenuClick: () => void }) => (
    <div data-testid="header" onClick={onMenuClick} />
  ),
}));

describe("AppShell — admin-shell class", () => {
  it("adds admin-shell class when role is super_admin", () => {
    const { container } = render(
      <AppShell role="super_admin">
        <div>content</div>
      </AppShell>
    );
    const root = container.firstElementChild;
    expect(root?.className).toContain("admin-shell");
  });

  it("adds admin-shell class when role is admin_store", () => {
    const { container } = render(
      <AppShell role="admin_store">
        <div>content</div>
      </AppShell>
    );
    const root = container.firstElementChild;
    expect(root?.className).toContain("admin-shell");
  });

  it("does NOT add admin-shell class when role is student", () => {
    const { container } = render(
      <AppShell role="student">
        <div>content</div>
      </AppShell>
    );
    const root = container.firstElementChild;
    expect(root?.className).not.toContain("admin-shell");
  });
});
