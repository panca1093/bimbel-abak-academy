import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import React from "react";
import { AppSidebar } from "./AppSidebar";

const replace = vi.fn();

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

let authStore = {
  user: null as {
    name?: string;
    email?: string;
    username?: string;
    role?: string;
    photo_url?: string;
  } | null,
};

vi.mock("@/stores/auth", () => ({
  useAuthStore: (selector: (s: typeof authStore) => unknown) => selector(authStore),
}));

vi.mock("./AbakMark", () => ({
  AbakMark: () => <span data-testid="abak-mark" />,
}));

// Radix AvatarImage only renders <img> after native image load, which
// never fires in jsdom.  We mock the avatar UI to behave like a real
// browser: render <img> when src is provided, fallback otherwise.
vi.mock("@/components/ui/avatar", () => ({
  Avatar: ({ children, className, ...props }: any) => (
    <div data-slot="avatar" className={className} {...props}>
      {children}
    </div>
  ),
  AvatarImage: ({ src, ...props }: { src?: string; [key: string]: any }) =>
    src ? <img src={src} alt="" data-slot="avatar-image" {...props} /> : null,
  AvatarFallback: ({ children, ...props }: any) => (
    <span data-slot="avatar-fallback" {...props}>
      {children}
    </span>
  ),
}));

describe("AppSidebar — avatar rendering", () => {
  beforeEach(() => {
    replace.mockClear();
    authStore = { user: null };
  });

  it("renders AvatarImage when user.photo_url is set", () => {
    authStore = {
      user: {
        name: "Budi Santoso",
        email: "budi@test.com",
        role: "student",
        photo_url: "https://example.com/photo.jpg",
      },
    };
    const { container } = render(<AppSidebar role="student" />);

    const imgs = container.querySelectorAll('img[data-slot="avatar-image"]');
    expect(imgs.length).toBeGreaterThan(0);
    const imgWithSrc = Array.from(imgs).find(
      (img) => img.getAttribute("src") === "https://example.com/photo.jpg"
    );
    expect(imgWithSrc).toBeTruthy();
  });

  it("falls back to AvatarFallback initials when photo_url is absent", () => {
    authStore = {
      user: {
        name: "Budi Santoso",
        email: "budi@test.com",
        role: "student",
        photo_url: undefined,
      },
    };
    render(<AppSidebar role="student" />);

    const imgs = document.querySelectorAll('img[data-slot="avatar-image"]');
    expect(imgs.length).toBe(0);
    expect(screen.getAllByText("B").length).toBeGreaterThan(0);
  });

  it("shows default initial 'A' when user has no name/email/username", () => {
    authStore = { user: { role: "student", photo_url: undefined } };
    render(<AppSidebar role="student" />);

    expect(screen.getByText("A")).toBeInTheDocument();
  });
});
