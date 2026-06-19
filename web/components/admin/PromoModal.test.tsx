import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, fireEvent } from "@testing-library/react";
import { PromoModal } from "./PromoModal";
import type { PromoCode } from "@/lib/types";

const mockOnSubmit = vi.fn();
const mockOnOpenChange = vi.fn();

describe("PromoModal", () => {
  beforeEach(() => {
    mockOnSubmit.mockReset();
    mockOnOpenChange.mockReset();
  });

  it("renders create modal with empty fields", () => {
    render(
      <PromoModal
        open={true}
        onOpenChange={mockOnOpenChange}
        onSubmit={mockOnSubmit}
        isPending={false}
      />
    );

    expect(screen.getByRole("dialog", { name: /create promo code/i })).toBeInTheDocument();
    expect(screen.getByPlaceholderText(/e\.g\. diskon10/i)).toHaveValue("");
    expect(screen.getByLabelText(/discount type/i)).toHaveValue("percent");
  });

  it("renders edit modal prefilled with promo data", () => {
    const promo: PromoCode = {
      id: "promo-1",
      code: "DISKON10",
      discount_percent: 10,
      max_discount_amount: 50000,
      min_order_amount: 100000,
      max_uses: 100,
      used_count: 3,
      expires_at: "2026-12-31T00:00:00Z",
    };

    render(
      <PromoModal
        open={true}
        onOpenChange={mockOnOpenChange}
        promo={promo}
        onSubmit={mockOnSubmit}
        isPending={false}
      />
    );

    expect(screen.getByRole("dialog", { name: /edit promo code/i })).toBeInTheDocument();
    expect(screen.getByDisplayValue("DISKON10")).toBeInTheDocument();
    expect(screen.getByDisplayValue("10")).toBeInTheDocument();
    expect(screen.getByDisplayValue("100")).toBeInTheDocument();
  });

  it("submits create input with percent discount", async () => {
    render(
      <PromoModal
        open={true}
        onOpenChange={mockOnOpenChange}
        onSubmit={mockOnSubmit}
        isPending={false}
      />
    );

    fireEvent.input(screen.getByPlaceholderText(/e\.g\. diskon10/i), { target: { value: "DISKON15" } });
    fireEvent.input(screen.getByLabelText(/discount value/i), { target: { value: "15" } });
    fireEvent.input(screen.getByLabelText(/max uses/i), { target: { value: "50" } });

    fireEvent.click(screen.getByRole("button", { name: /^save$/i }));

    await waitFor(() => {
      expect(mockOnSubmit).toHaveBeenCalledWith(
        expect.objectContaining({
          code: "DISKON15",
          discount_percent: 15,
          max_uses: 50,
        })
      );
    });
  });

  it("submits create input with fixed discount", async () => {
    render(
      <PromoModal
        open={true}
        onOpenChange={mockOnOpenChange}
        onSubmit={mockOnSubmit}
        isPending={false}
      />
    );

    fireEvent.input(screen.getByPlaceholderText(/e\.g\. diskon10/i), { target: { value: "CASHBACK20K" } });

    const typeSelect = screen.getByLabelText(/discount type/i);
    fireEvent.change(typeSelect, { target: { value: "fixed" } });

    fireEvent.input(screen.getByLabelText(/discount value/i), { target: { value: "20000" } });
    fireEvent.input(screen.getByLabelText(/min order amount/i), { target: { value: "50000" } });

    fireEvent.click(screen.getByRole("button", { name: /^save$/i }));

    await waitFor(() => {
      expect(mockOnSubmit).toHaveBeenCalledWith(
        expect.objectContaining({
          code: "CASHBACK20K",
          discount_amount: 20000,
          min_order_amount: 50000,
        })
      );
    });
  });

  it("submits edit input only with allowed fields", async () => {
    const promo: PromoCode = {
      id: "promo-1",
      code: "DISKON10",
      discount_percent: 10,
      max_uses: 100,
      used_count: 3,
    };

    render(
      <PromoModal
        open={true}
        onOpenChange={mockOnOpenChange}
        promo={promo}
        onSubmit={mockOnSubmit}
        isPending={false}
      />
    );

    fireEvent.input(screen.getByLabelText(/max uses/i), { target: { value: "200" } });
    fireEvent.click(screen.getByRole("button", { name: /^save$/i }));

    await waitFor(() => {
      expect(mockOnSubmit).toHaveBeenCalledWith({ max_uses: 200 });
    });
  });

  it("disables save when code or discount value is empty", () => {
    render(
      <PromoModal
        open={true}
        onOpenChange={mockOnOpenChange}
        onSubmit={mockOnSubmit}
        isPending={false}
      />
    );

    const saveButton = screen.getByRole("button", { name: /^save$/i });
    expect(saveButton).toBeDisabled();
  });
});
