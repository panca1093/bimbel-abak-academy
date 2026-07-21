package repository

import (
	"context"
	"testing"
	"time"

	"akademi-bimbel/internal/model"
)

// TestListProducts_AvailabilityWindow verifies the public catalog (VisibleOnly)
// hides products outside their availability window (P-A), while an admin listing
// (no VisibleOnly) still sees them.
func TestListProducts_AvailabilityWindow(t *testing.T) {
	ctx := context.Background()
	pool := newGradingTestPool(t)
	r := New(pool)

	now := time.Now()
	tomorrow := now.Add(24 * time.Hour)
	yesterday := now.Add(-24 * time.Hour)

	mk := func(name string, from, until *time.Time) {
		t.Helper()
		p := model.Product{
			Type: "book", Name: name, Price: 1000, Stock: 1, Status: "published",
			AvailableFrom: from, AvailableUntil: until,
		}
		if err := r.CreateProduct(ctx, &p); err != nil {
			t.Fatalf("create %s: %v", name, err)
		}
	}
	mk("always", nil, nil)
	mk("future", &tomorrow, nil)  // not yet available
	mk("expired", nil, &yesterday) // window has passed

	names := func(f ProductFilter) map[string]bool {
		t.Helper()
		got, _, err := r.ListProducts(ctx, f)
		if err != nil {
			t.Fatalf("list products: %v", err)
		}
		m := map[string]bool{}
		for _, p := range got {
			m[p.Name] = true
		}
		return m
	}

	public := names(ProductFilter{VisibleOnly: true, Limit: 100})
	if !public["always"] {
		t.Error("always-available product should be in the public catalog")
	}
	if public["future"] {
		t.Error("not-yet-available product must be hidden from the public catalog")
	}
	if public["expired"] {
		t.Error("expired product must be hidden from the public catalog")
	}

	admin := names(ProductFilter{Limit: 100})
	if !admin["always"] || !admin["future"] || !admin["expired"] {
		t.Errorf("admin listing should include all products regardless of window, got %v", admin)
	}
}
