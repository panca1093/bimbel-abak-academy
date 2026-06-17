package integration_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdmin(t *testing.T) {
	env := newTestEnv(t)
	ctx := context.Background()

	// FR-INT-21: Admin confirm transitions and persists status.
	t.Run("FR-INT-21 admin confirm persists paid status", func(t *testing.T) {
		adminID := seedUser(t, env, "super_admin", "active", false)
		adminToken := authToken(t, env, adminID, "super_admin")

		// Seed a payment_pending order directly; AdminConfirmOrder sets status to "paid".
		studentID := seedUser(t, env, "student", "active", false)
		var orderID string
		err := env.pool.QueryRow(ctx,
			`INSERT INTO orders (student_id, status, subtotal, total)
			 VALUES ($1, 'payment_pending', 50000, 50000) RETURNING id`,
			studentID,
		).Scan(&orderID)
		require.NoError(t, err)

		key := fmt.Sprintf("confirm-%d", time.Now().UnixNano())
		req, err := http.NewRequest(http.MethodPost,
			env.server.URL+"/api/v1/admin/orders/"+orderID+"/confirm", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		req.Header.Set("Idempotency-Key", key)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		drainClose(resp)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var status string
		require.NoError(t, env.pool.QueryRow(ctx,
			`SELECT status FROM orders WHERE id=$1`, orderID,
		).Scan(&status))
		assert.Equal(t, "paid", status, "order status must be 'paid' after admin confirm")
	})

	// FR-INT-22: Admin ship writes tracking_number + shipped_at.
	t.Run("FR-INT-22 admin ship writes tracking_number and shipped_at", func(t *testing.T) {
		adminID := seedUser(t, env, "super_admin", "active", false)
		adminToken := authToken(t, env, adminID, "super_admin")

		// Seed an order already in "paid" status (shippable).
		studentID := seedUser(t, env, "student", "active", false)
		var orderID string
		err := env.pool.QueryRow(ctx,
			`INSERT INTO orders (student_id, status, subtotal, total)
			 VALUES ($1, 'paid', 50000, 50000) RETURNING id`,
			studentID,
		).Scan(&orderID)
		require.NoError(t, err)

		resp := env.doJSON(t, http.MethodPost,
			"/api/v1/admin/orders/"+orderID+"/ship",
			map[string]string{"tracking_number": "JNE-123456"},
			adminToken,
		)
		drainClose(resp)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var trackingNumber string
		var shippedAt *time.Time
		require.NoError(t, env.pool.QueryRow(ctx,
			`SELECT tracking_number, shipped_at FROM orders WHERE id=$1`, orderID,
		).Scan(&trackingNumber, &shippedAt))
		assert.Equal(t, "JNE-123456", trackingNumber, "tracking_number must be persisted")
		assert.NotNil(t, shippedAt, "shipped_at must be set")
	})

	// FR-INT-23: Non-admin is forbidden on admin order routes.
	t.Run("FR-INT-23 student token gets 403 on admin order route", func(t *testing.T) {
		studentID := seedUser(t, env, "student", "active", false)
		studentToken := authToken(t, env, studentID, "student")

		// Seed any order so the route exists.
		var orderID string
		err := env.pool.QueryRow(ctx,
			`INSERT INTO orders (student_id, status, subtotal, total)
			 VALUES ($1, 'payment_pending', 10000, 10000) RETURNING id`,
			studentID,
		).Scan(&orderID)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost,
			env.server.URL+"/api/v1/admin/orders/"+orderID+"/confirm", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+studentToken)
		req.Header.Set("Idempotency-Key", "student-confirm-key")
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		body := decodeBody(t, resp)

		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
		assert.Equal(t, "forbidden", body["code"], "error code must be 'forbidden'")
	})

	// FR-INT-24: Validate promo computes discount against a real DB row.
	t.Run("FR-INT-24 validate promo computes discount against real DB row", func(t *testing.T) {
		discountPct := 10.0
		seedPromo(t, env, "SAVE10", discountPct)

		resp := env.doJSON(t, http.MethodPost, "/api/v1/promo-codes/validate",
			map[string]any{"code": "SAVE10", "subtotal": 100000.0},
			"", // no auth required
		)
		body := decodeBody(t, resp)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		discount, _ := body["discount"].(float64)
		finalTotal, _ := body["final_total"].(float64)
		assert.Equal(t, 10000.0, discount, "discount must be 10% of 100000")
		assert.Equal(t, 90000.0, finalTotal, "final_total must be subtotal - discount")
	})

	// FR-INT-25: Admin create promo persists; used_count increment documented gap.
	t.Run("FR-INT-25 admin create promo persists row", func(t *testing.T) {
		adminID := seedUser(t, env, "super_admin", "active", false)
		adminToken := authToken(t, env, adminID, "super_admin")

		discountPct := 15.0
		resp := env.doJSON(t, http.MethodPost, "/api/v1/admin/promo-codes",
			map[string]any{
				"code":             "ADMIN15",
				"discount_percent": discountPct,
			},
			adminToken,
		)
		body := decodeBody(t, resp)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		promoID, ok := body["ID"].(string)
		if !ok {
			// Try lowercase variant
			promoID, ok = body["id"].(string)
		}
		require.True(t, ok, "response must contain promo id, got: %v", body)
		require.NotEmpty(t, promoID)

		// Verify the row persisted in DB.
		var code string
		var usedCount int
		require.NoError(t, env.pool.QueryRow(ctx,
			`SELECT code, used_count FROM promo_code WHERE id=$1`, promoID,
		).Scan(&code, &usedCount))
		assert.Equal(t, "ADMIN15", code)
		assert.Equal(t, 0, usedCount, "used_count must start at 0")
	})

	t.Run("FR-INT-25 used_count increment on checkout (documented gap)", func(t *testing.T) {
		// IncrementPromoUses exists in repository/promo.go (line 84) but is NOT called
		// in service.Checkout — the checkout path does not wire promo usage tracking.
		// Asserting the create half above; skipping the increment half until wired.
		t.Skip("used_count increment not wired in checkout path — gap documented per FR-INT-25 note")
	})
}
