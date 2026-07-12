import { describe, it, expect, vi, afterEach } from "vitest";
import { apiFetch, ApiError } from "./api";

describe("apiFetch error body", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("retains the parsed error JSON on ApiError.body", async () => {
    const payload = {
      code: "verification_pending",
      message: "verify your email",
      otp_required: true,
      pending_token: "tok-123",
      id: "user@example.com",
    };
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response(JSON.stringify(payload), {
          status: 403,
          headers: { "Content-Type": "application/json" },
        }),
      ),
    );

    await expect(apiFetch("/auth/login", { method: "POST" })).rejects.toMatchObject({
      code: "verification_pending",
      status: 403,
      body: payload,
    });
  });

  it("leaves body undefined when the response has no JSON body", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(new Response("", { status: 500, statusText: "Server Error" })),
    );

    try {
      await apiFetch("/whatever");
      throw new Error("expected apiFetch to reject");
    } catch (err) {
      expect(err).toBeInstanceOf(ApiError);
      expect((err as ApiError).body).toBeUndefined();
    }
  });
});
