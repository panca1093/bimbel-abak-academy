import { describe, expect, it, vi } from "vitest";
import { act, render, screen } from "@testing-library/react";
import { Suspense } from "react";
import ProductDetailPage from "./page";

vi.mock("next/navigation", () => ({
  useRouter: () => ({ push: vi.fn() }),
}));

vi.mock("@/lib/hooks/products", () => ({
  useProduct: () => ({
    data: {
      id: "merch-1",
      type: "merchandise",
      name: "Kaos Akademi",
      price: 75000,
      stock: 12,
      status: "published",
    },
    isLoading: false,
    isError: false,
    error: null,
    refetch: vi.fn(),
  }),
}));

vi.mock("@/lib/hooks/orders", () => ({
  useAddToCart: () => ({ mutate: vi.fn(), isPending: false }),
  useCart: () => ({ data: { items: [] } }),
}));

vi.mock("@/stores/cart", () => ({
  useCartStore: (selector: (state: { setCount: ReturnType<typeof vi.fn> }) => unknown) => selector({ setCount: vi.fn() }),
  default: { getState: () => ({ count: 0 }) },
}));

vi.mock("@/stores/auth", () => ({
  useAuthStore: (selector: (state: { token: string }) => unknown) => selector({ token: "token" }),
}));

describe("ProductDetailPage", () => {
  it("shows stock and delivery information for merchandise", async () => {
    await act(async () => {
      render(
        <Suspense fallback={null}>
          <ProductDetailPage params={Promise.resolve({ id: "merch-1" })} />
        </Suspense>,
      );
    });

    expect(await screen.findAllByText(/stok: 12/i)).toHaveLength(2);
    expect(screen.getByText(/dikirim ke alamat/i)).toBeInTheDocument();
  });
});
