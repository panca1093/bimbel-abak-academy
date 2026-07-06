import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import NotificationsPage from "./page";

vi.mock("@/components/admin/PurchaseNotificationFeed", () => ({
  PurchaseNotificationFeed: () => <div data-testid="purchase-feed">Notifikasi Pembelian</div>,
}));

vi.mock("@/components/admin/AnnouncementTable", () => ({
  AnnouncementTable: ({ onCreateClick, onEdit }: { onCreateClick: () => void; onEdit: () => void }) => (
    <div data-testid="announcement-table">
      <button onClick={onCreateClick} data-testid="create-btn">Buat</button>
      <button onClick={() => onEdit()} data-testid="edit-btn">Edit</button>
    </div>
  ),
}));

vi.mock("@/components/admin/AnnouncementComposer", () => ({
  AnnouncementComposer: () => <div data-testid="announcement-composer" />,
}));

describe("NotificationsPage", () => {
  it("renders the page-level AdminPageHeader with title", () => {
    render(<NotificationsPage />);
    expect(screen.getByRole("heading", { level: 1, name: "Notifikasi" })).toBeInTheDocument();
  });

  it("renders the purchase notification feed section", () => {
    render(<NotificationsPage />);
    expect(screen.getByTestId("purchase-feed")).toHaveTextContent("Notifikasi Pembelian");
  });

  it("renders the announcement table section", () => {
    render(<NotificationsPage />);
    expect(screen.getByTestId("announcement-table")).toBeInTheDocument();
  });

  it("renders the announcement composer", () => {
    render(<NotificationsPage />);
    expect(screen.getByTestId("announcement-composer")).toBeInTheDocument();
  });
});
