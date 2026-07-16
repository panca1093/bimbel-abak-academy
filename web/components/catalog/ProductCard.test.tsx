import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { ProductCard } from "./ProductCard";

describe("ProductCard", () => {
  it("resolves object keys, preserves legacy URLs, and falls back for merchandise", () => {
    const { rerender } = render(
      <ProductCard
        product={{
          id: "merch-key",
          type: "merchandise",
          name: "Kaos Akademi",
          price: 75000,
          image_url: "avatars/store/tee.png",
        }}
      />,
    );

    let cover = screen.getByText("Kaos Akademi").closest("a")?.firstElementChild as HTMLElement;
    expect(cover.style.backgroundImage).toContain("http://localhost:8080/api/v1/files/avatars/store/tee.png");

    rerender(
      <ProductCard
        product={{
          id: "merch-legacy",
          type: "merchandise",
          name: "Tote Akademi",
          price: 50000,
          image_url: "https://cdn.example.com/tote.png",
        }}
      />,
    );

    cover = screen.getByText("Tote Akademi").closest("a")?.firstElementChild as HTMLElement;
    expect(cover.style.backgroundImage).toContain("https://cdn.example.com/tote.png");

    rerender(
      <ProductCard
        product={{ id: "medal-fallback", type: "medal", name: "Medali", price: 10000 }}
      />,
    );

    cover = screen.getAllByText("Medali")[0].closest("a")?.firstElementChild as HTMLElement;
    expect(cover.style.background).toContain("linear-gradient");
  });
});
