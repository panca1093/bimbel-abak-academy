import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, waitFor, act } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import {
  useAdminProducts,
  useCreateProduct,
  useUpdateProduct,
  usePublishProduct,
  useDeleteProduct,
  adminProductsKeys,
} from "./admin-products";
import type { Product } from "@/lib/types";

const mockAuthFetch = vi.fn();

vi.mock("@/lib/api", () => ({
  authFetch: (...args: Parameters<typeof mockAuthFetch>) => mockAuthFetch(...args),
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

describe("admin-products hooks", () => {
  beforeEach(() => {
    mockAuthFetch.mockReset();
  });

  afterEach(() => {
    vi.clearAllTimers();
  });

  it("useAdminProducts fetches GET /admin/products and returns data", async () => {
    const products: Product[] = [
      { id: "p1", type: "book", name: "Buku A", price: 10000, stock: 5, status: "published" },
    ];
    mockAuthFetch.mockResolvedValueOnce({ data: products });

    const { wrapper } = wrapperFactory();
    const { result } = renderHook(() => useAdminProducts(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(mockAuthFetch).toHaveBeenCalledWith("/admin/products");
    expect(result.current.data).toEqual(products);
  });

  it("useCreateProduct posts to /admin/products and invalidates list", async () => {
    const product: Product = { id: "p2", type: "course", name: "Kursus A", price: 50000, status: "draft" };
    mockAuthFetch.mockResolvedValueOnce(product);

    const { wrapper, queryClient } = wrapperFactory();
    const spy = vi.spyOn(queryClient, "invalidateQueries");
    const { result } = renderHook(() => useCreateProduct(), { wrapper });

    await act(async () => {
      await result.current.mutateAsync({
        type: "course",
        name: "Kursus A",
        description: "desc",
        price: 50000,
        course_ids: ["c1"],
      });
    });

    expect(mockAuthFetch).toHaveBeenCalledWith("/admin/products", {
      method: "POST",
      body: JSON.stringify({
        type: "course",
        name: "Kursus A",
        description: "desc",
        price: 50000,
        course_ids: ["c1"],
      }),
    });
    expect(spy).toHaveBeenCalledWith({ queryKey: adminProductsKeys.list() });
  });

  it("useUpdateProduct patches /admin/products/:id and invalidates list", async () => {
    const product: Product = { id: "p1", type: "book", name: "Buku A v2", price: 12000, status: "published" };
    mockAuthFetch.mockResolvedValueOnce(product);

    const { wrapper, queryClient } = wrapperFactory();
    const spy = vi.spyOn(queryClient, "invalidateQueries");
    const { result } = renderHook(() => useUpdateProduct(), { wrapper });

    await act(async () => {
      await result.current.mutateAsync({ id: "p1", input: { name: "Buku A v2", price: 12000 } });
    });

    expect(mockAuthFetch).toHaveBeenCalledWith("/admin/products/p1", {
      method: "PATCH",
      body: JSON.stringify({ name: "Buku A v2", price: 12000 }),
    });
    expect(spy).toHaveBeenCalledWith({ queryKey: adminProductsKeys.list() });
  });

  it("usePublishProduct posts to /admin/products/:id/publish and invalidates list", async () => {
    mockAuthFetch.mockResolvedValueOnce({ message: "product published" });

    const { wrapper, queryClient } = wrapperFactory();
    const spy = vi.spyOn(queryClient, "invalidateQueries");
    const { result } = renderHook(() => usePublishProduct(), { wrapper });

    await act(async () => {
      await result.current.mutateAsync("p1");
    });

    expect(mockAuthFetch).toHaveBeenCalledWith("/admin/products/p1/publish", { method: "POST" });
    expect(spy).toHaveBeenCalledWith({ queryKey: adminProductsKeys.list() });
  });

  it("useDeleteProduct deletes /admin/products/:id and invalidates list", async () => {
    mockAuthFetch.mockResolvedValueOnce(undefined);

    const { wrapper, queryClient } = wrapperFactory();
    const spy = vi.spyOn(queryClient, "invalidateQueries");
    const { result } = renderHook(() => useDeleteProduct(), { wrapper });

    await act(async () => {
      await result.current.mutateAsync("p1");
    });

    expect(mockAuthFetch).toHaveBeenCalledWith("/admin/products/p1", { method: "DELETE" });
    expect(spy).toHaveBeenCalledWith({ queryKey: adminProductsKeys.list() });
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
