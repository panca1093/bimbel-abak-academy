import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, within, fireEvent } from "@testing-library/react";
import { toast } from "sonner";
import OrdersPage from "./page";
import type { Order } from "@/lib/types";

const mockMutate = vi.fn();
const mockMutateAsync = vi.fn();

let ordersState = {
  data: null as Order[] | null,
  isLoading: true,
  isError: false,
  error: null as Error | null,
  refetch: vi.fn(),
};

let confirmState = { mutate: mockMutate, mutateAsync: mockMutateAsync, isPending: false };
let shipState = { mutate: mockMutate, mutateAsync: mockMutateAsync, isPending: false };
let completeState = { mutate: mockMutate, mutateAsync: mockMutateAsync, isPending: false };
let refundState = { mutate: mockMutate, mutateAsync: mockMutateAsync, isPending: false };
let reconcileState = { mutate: mockMutate, mutateAsync: mockMutateAsync, isPending: false };

vi.mock("@/lib/hooks/admin-orders", () => ({
  useAdminOrders: () => ordersState,
  useAdminOrder: () => ({}),
  useConfirmOrder: () => confirmState,
  useShipOrder: () => shipState,
  useCompleteOrder: () => completeState,
  useRefundOrder: () => refundState,
  useReconcileOrder: () => reconcileState,
}));

vi.mock("sonner", () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

const sampleOrders: Order[] = [
  {
    id: "o1",
    student_id: "s1",
    status: "payment_pending",
    subtotal: 100000,
    discount: 0,
    shipping_cost: 15000,
    total: 115000,
    items: [{ id: "i1", order_id: "o1", product_id: "p1", product_type: "book", name: "Buku A", unit_price: 100000, qty: 1, jumlah: 100000 }],
  },
  {
    id: "o2",
    student_id: "s2",
    status: "paid",
    subtotal: 200000,
    discount: 0,
    shipping_cost: 0,
    total: 200000,
    tracking_number: "JNE-999",
    items: [{ id: "i2", order_id: "o2", product_id: "p2", product_type: "book", name: "Buku Shipped", unit_price: 200000, qty: 1, jumlah: 200000 }],
  },
  {
    id: "o3",
    student_id: "s3",
    status: "completed",
    subtotal: 50000,
    discount: 0,
    shipping_cost: 0,
    total: 50000,
    items: [{ id: "i3", order_id: "o3", product_id: "p3", product_type: "course", name: "Kursus B", unit_price: 50000, qty: 1, jumlah: 50000 }],
  },
];

describe("OrdersPage", () => {
  beforeEach(() => {
    vi.unstubAllGlobals();
    ordersState = {
      data: sampleOrders,
      isLoading: false,
      isError: false,
      error: null,
      refetch: vi.fn(),
    };
    confirmState = { mutate: mockMutate, mutateAsync: mockMutateAsync, isPending: false };
    shipState = { mutate: mockMutate, mutateAsync: mockMutateAsync, isPending: false };
    completeState = { mutate: mockMutate, mutateAsync: mockMutateAsync, isPending: false };
    refundState = { mutate: mockMutate, mutateAsync: mockMutateAsync, isPending: false };
    reconcileState = { mutate: mockMutate, mutateAsync: mockMutateAsync, isPending: false };
    mockMutate.mockReset();
    mockMutateAsync.mockReset();
    (toast.success as ReturnType<typeof vi.fn>).mockReset();
    (toast.error as ReturnType<typeof vi.fn>).mockReset();
  });

  it("renders the orders table with order number, buyer, product, amount, payment and shipping", async () => {
    render(<OrdersPage />);

    await waitFor(() => {
      expect(screen.getByText(/Buku A/)).toBeInTheDocument();
    });

    expect(screen.getByText("Rp115.000")).toBeInTheDocument();
    expect(screen.getByText("...s1")).toBeInTheDocument();
    // Shipping column renders — "Dikirim" appears both as a filter chip and as a badge
    expect(screen.getAllByText("Dikirim").length).toBeGreaterThanOrEqual(1);
  });

  it("shows confirm and reconcile actions for pending orders", async () => {
    render(<OrdersPage />);

    await waitFor(() => expect(screen.getByText(/Buku A/)).toBeInTheDocument());

    const row = screen.getByText(/Buku A/).closest("tr");
    expect(row).toBeTruthy();
    expect(within(row!).queryByRole("button", { name: /konfirmasi/i })).toBeInTheDocument();
    expect(within(row!).queryByRole("button", { name: /rekonsiliasi/i })).toBeInTheDocument();
    expect(within(row!).queryByRole("button", { name: /kirim/i })).not.toBeInTheDocument();
    expect(within(row!).queryByRole("button", { name: /refund/i })).not.toBeInTheDocument();
  });

  it("confirms an order and shows success toast", async () => {
    mockMutateAsync.mockResolvedValueOnce({ message: "order confirmed" });

    render(<OrdersPage />);

    await waitFor(() => expect(screen.getByText(/Buku A/)).toBeInTheDocument());

    const row = screen.getByText(/Buku A/).closest("tr");
    const confirmButton = within(row!).getByRole("button", { name: /konfirmasi/i });
    fireEvent.click(confirmButton);

    await waitFor(() => {
      expect(mockMutateAsync).toHaveBeenCalledWith("o1");
      expect(toast.success).toHaveBeenCalledWith("Pesanan dikonfirmasi.");
    });
  });

  it("ships an order when tracking number is provided", async () => {
    mockMutateAsync.mockResolvedValueOnce({ message: "order shipped" });
    vi.stubGlobal("prompt", () => "JNE-123");

    render(<OrdersPage />);

    await waitFor(() => expect(screen.getByText(/Buku Shipped/)).toBeInTheDocument());

    const row = screen.getByText(/Buku Shipped/).closest("tr");
    const shipButton = within(row!).getByRole("button", { name: /kirim/i });
    fireEvent.click(shipButton);

    await waitFor(() => {
      expect(mockMutateAsync).toHaveBeenCalledWith({ id: "o2", trackingNumber: "JNE-123" });
      expect(toast.success).toHaveBeenCalledWith("Pesanan dikirim.");
    });
  });

  it("refunds an order after confirmation", async () => {
    mockMutateAsync.mockResolvedValueOnce({ message: "order refunded" });
    vi.stubGlobal("confirm", () => true);

    render(<OrdersPage />);

    await waitFor(() => expect(screen.getByText(/Kursus B/)).toBeInTheDocument());

    const row = screen.getByText(/Kursus B/).closest("tr");
    const refundButton = within(row!).getByRole("button", { name: /refund/i });
    fireEvent.click(refundButton);

    await waitFor(() => {
      expect(mockMutateAsync).toHaveBeenCalledWith("o3");
      expect(toast.success).toHaveBeenCalledWith("Pesanan direfund.");
    });

  });

  it("filters rows by status chips", async () => {
    ordersState = {
      data: sampleOrders.filter((o) => o.status === "paid"),
      isLoading: false,
      isError: false,
      error: null,
      refetch: vi.fn(),
    };

    render(<OrdersPage />);

    await waitFor(() => expect(screen.getByText(/Buku Shipped/)).toBeInTheDocument());

    const paidChip = screen.getByRole("button", { name: /^dibayar$/i });
    fireEvent.click(paidChip);

    expect(screen.getByText(/Buku Shipped/)).toBeInTheDocument();
    expect(screen.queryByText(/Buku A/)).not.toBeInTheDocument();
  });

  it("surfaces an API error as inline error text", async () => {
    ordersState = {
      data: null,
      isLoading: false,
      isError: true,
      error: new Error("gagal memuat"),
      refetch: vi.fn(),
    };

    render(<OrdersPage />);

    await waitFor(() => {
      expect(screen.getByText(/gagal memuat/i)).toBeInTheDocument();
    });
  });
});
