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

export async function authFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const { useAuthStore } = await import("@/stores/auth");
  const token = useAuthStore.getState().token;

  const headers: HeadersInit = {
    "Content-Type": "application/json",
    ...(init?.headers ?? {}),
  };
  if (token) {
    (headers as Record<string, string>)["Authorization"] = `Bearer ${token}`;
  }

  const res = await fetch(`${API_BASE}${path}`, { ...init, headers });
  if (!res.ok) {
    if (res.status === 401) {
      useAuthStore.getState().clear();
      if (typeof window !== "undefined") {
        window.location.href = "/login";
      }
    }
    throw await parseError(res);
  }
  return res.json() as Promise<T>;
}