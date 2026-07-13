"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { useGoogleLogin } from "@/lib/hooks/auth";
import { isProfileComplete } from "@/lib/profile";
import { redirectForRole } from "@/lib/auth-redirect";
import { useRouter } from "next/navigation";

declare global {
  interface Window {
    google?: {
      accounts: {
        id: {
          initialize: (config: {
            client_id: string;
            callback: (response: { credential: string }) => void;
          }) => void;
          renderButton: (element: HTMLElement, options: { theme: string; size: string; text?: string }) => void;
        };
      };
    };
  }
}

const GSI_SRC = "https://accounts.google.com/gsi/client";
const GOOGLE_CLIENT_ID = process.env.NEXT_PUBLIC_GOOGLE_CLIENT_ID ?? "";

interface Props {
  /** Button text passed to GSI renderButton. "signin_with" | "signup_with" */
  text?: "signin_with" | "signup_with";
}

export function GoogleSignInButton({ text }: Props) {
  const containerRef = useRef<HTMLDivElement>(null);
  const googleLogin = useGoogleLogin();
  const router = useRouter();
  const [scriptLoaded, setScriptLoaded] = useState(false);

  // Inject GSI script once.
  useEffect(() => {
    if (!GOOGLE_CLIENT_ID) return;
    if (document.querySelector('script[src*="gsi/client"]')) {
      setScriptLoaded(true);
      return;
    }
    const script = document.createElement("script");
    script.src = GSI_SRC;
    script.async = true;
    script.onload = () => setScriptLoaded(true);
    document.head.appendChild(script);
  }, []);

  // Initialize + render after script is loaded and container is mounted.
  const handleCredential = useCallback(
    (response: { credential: string }) => {
      googleLogin.mutate(response.credential, {
        onSuccess: (data) => {
          if (!data.user) return;
          // Route: incomplete Google user → /complete-profile, else role home.
          if (
            data.user.auth_provider === "google" &&
            !isProfileComplete(data.user)
          ) {
            router.replace("/complete-profile");
          } else {
            router.replace(redirectForRole(data.user.role));
          }
        },
      });
    },
    [googleLogin, router]
  );

  useEffect(() => {
    if (!scriptLoaded || !containerRef.current || !GOOGLE_CLIENT_ID) return;
    const w = window as Window;
    if (!w.google?.accounts?.id) return;

    w.google.accounts.id.initialize({
      client_id: GOOGLE_CLIENT_ID,
      callback: handleCredential,
    });
    w.google.accounts.id.renderButton(containerRef.current, {
      theme: "outline",
      size: "large",
      text,
    });
  }, [scriptLoaded, handleCredential, text]);

  if (!GOOGLE_CLIENT_ID) return null;

  return (
    <div className="flex flex-col items-center gap-3">
      <div ref={containerRef} className="flex justify-center" />
      {googleLogin.isPending && (
        <span className="text-xs text-ink-400">Menghubungkan…</span>
      )}
    </div>
  );
}
