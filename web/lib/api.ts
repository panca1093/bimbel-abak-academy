const API_BASE = process.env.NEXT_PUBLIC_API_BASE ?? "http://localhost:8080/api/v1";

export class ApiError extends Error {
  readonly code: string;
  readonly status: number;

  constructor(code: string, message: string, status: number) {
    super(message);
    this.name = "ApiError";
    this.code = code;
    this.status = status;
  }
}

interface ErrorBody {
  code?: string;
  message?: string;
  error?: string;
}

async function parseError(res: Response): Promise<ApiError> {
  const status = res.status;
  let body: ErrorBody | null = null;
  try {
    const text = await res.text();
    if (text) {
      body = JSON.parse(text) as ErrorBody;
    }
  } catch {
    // non-JSON body; fall through to status text
  }
  const message =
    body?.message ??
    body?.error ??
    res.statusText ??
    `Request failed with status ${status}`;
  const code = body?.code ?? `HTTP_${status}`;
  return new ApiError(code, message, status);
}

function withJsonHeaders(init?: RequestInit): HeadersInit {
  return { "Content-Type": "application/json", ...(init?.headers ?? {}) };
}

export { API_BASE };

export async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers: withJsonHeaders(init),
  });
  if (!res.ok) {
    throw await parseError(res);
  }
  return res.json() as Promise<T>;
}

async function tryRefresh(): Promise<string | null> {
  const { useAuthStore } = await import("@/stores/auth");
  const { refreshToken } = useAuthStore.getState();
  if (!refreshToken) return null;
  try {
    const res = await fetch(`${API_BASE}/auth/refresh`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });
    if (!res.ok) return null;
    const data = (await res.json()) as { access_token: string; refresh_token: string };
    const { useAuthStore: store } = await import("@/stores/auth");
    const user = store.getState().user!;
    store.getState().setSession(data.access_token, data.refresh_token, user);
    return data.access_token;
  } catch {
    return null;
  }
}

export async function authFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const { useAuthStore } = await import("@/stores/auth");
  const token = useAuthStore.getState().token;

  const buildHeaders = (t: string | null): HeadersInit => ({
    "Content-Type": "application/json",
    ...(init?.headers ?? {}),
    ...(t ? { Authorization: `Bearer ${t}` } : {}),
  });

  let res = await fetch(`${API_BASE}${path}`, { ...init, headers: buildHeaders(token) });

  if (res.status === 401) {
    const newToken = await tryRefresh();
    if (newToken) {
      res = await fetch(`${API_BASE}${path}`, { ...init, headers: buildHeaders(newToken) });
    }
  }

  if (!res.ok) {
    if (res.status === 401) {
      useAuthStore.getState().clear();
      if (typeof window !== "undefined") {
        window.location.href = "/login";
      }
    }
    throw await parseError(res);
  }
  if (res.status === 204 || res.status === 205) return undefined as T;
  return res.json() as Promise<T>;
}