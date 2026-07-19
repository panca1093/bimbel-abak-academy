package integration_test

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConcurrency(t *testing.T) {
	env := newTestEnv(t)
	ctx := context.Background()

	t.Run("FR-INT-17 idempotent checkout retry returns same gateway_ref, no second stock decrement", func(t *testing.T) {
		userID := seedUser(t, env, "student", "active", false)
		token := authToken(t, env, userID, "student")
		productID := seedProduct(t, env, "book", "Buku Idempotency", 50000)

		resp := env.doJSON(t, http.MethodPost, "/api/v1/orders", nil, token)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		body := decodeBody(t, resp)
		orderID := body["id"].(string)

		drainClose(env.doJSON(t, http.MethodPost, "/api/v1/orders/"+orderID+"/items",
			map[string]any{"product_id": productID, "qty": 1}, token))

		provinceID, cityID, districtID := seedRegionIDs(t, env)
		drainClose(env.doJSON(t, http.MethodPatch, "/api/v1/orders/"+orderID,
			map[string]any{
				"courier":       "JNE",
				"shipping_cost": 15000.0,
				"province_id":   provinceID,
				"city_id":       cityID,
				"district_id":   districtID,
				"kode_pos":      "12345",
			}, token))

		idempKey := fmt.Sprintf("idemp-%d", time.Now().UnixNano())

		// First checkout.
		co1 := checkoutWithKey(t, env, orderID, token, idempKey)
		body1 := decodeBody(t, co1)
		require.Equal(t, http.StatusOK, co1.StatusCode, "first checkout failed: %v", body1)
		gatewayRef1, _ := body1["gateway_ref"].(string)
		require.NotEmpty(t, gatewayRef1)

		var stockAfterFirst int
		require.NoError(t, env.pool.QueryRow(ctx,
			`SELECT stock FROM product WHERE id=$1`, productID,
		).Scan(&stockAfterFirst))

		// Second checkout with same Idempotency-Key.
		co2 := checkoutWithKey(t, env, orderID, token, idempKey)
		body2 := decodeBody(t, co2)
		require.Equal(t, http.StatusOK, co2.StatusCode, "second checkout failed: %v", body2)
		gatewayRef2, _ := body2["gateway_ref"].(string)

		assert.Equal(t, gatewayRef1, gatewayRef2, "second call must return the same gateway_ref (Redis cache hit)")

		var stockAfterSecond int
		require.NoError(t, env.pool.QueryRow(ctx,
			`SELECT stock FROM product WHERE id=$1`, productID,
		).Scan(&stockAfterSecond))
		assert.Equal(t, stockAfterFirst, stockAfterSecond, "stock must not be decremented a second time on idempotent retry")
	})

	t.Run("FR-INT-18 concurrent checkout stock=1: one 200 one 409, final stock=0", func(t *testing.T) {
		// Seed a product with stock=1.
		productID := seedProduct(t, env, "book", "Buku Satu Stok", 50000)
		_, err := env.pool.Exec(ctx, `UPDATE product SET stock=1 WHERE id=$1`, productID)
		require.NoError(t, err)

		// Two separate students, each with their own cart holding the same product.
		userA := seedUser(t, env, "student", "active", false)
		tokenA := authToken(t, env, userA, "student")

		userB := seedUser(t, env, "student", "active", false)
		tokenB := authToken(t, env, userB, "student")

		setupCart := func(token string) string {
			resp := env.doJSON(t, http.MethodPost, "/api/v1/orders", nil, token)
			require.Equal(t, http.StatusCreated, resp.StatusCode)
			b := decodeBody(t, resp)
			orderID := b["id"].(string)
			drainClose(env.doJSON(t, http.MethodPost, "/api/v1/orders/"+orderID+"/items",
				map[string]any{"product_id": productID, "qty": 1}, token))
			provinceID, cityID, districtID := seedRegionIDs(t, env)
			drainClose(env.doJSON(t, http.MethodPatch, "/api/v1/orders/"+orderID,
				map[string]any{
					"courier":       "JNE",
					"shipping_cost": 15000.0,
					"province_id":   provinceID,
					"city_id":       cityID,
					"district_id":   districtID,
					"kode_pos":      "12345",
				}, token))
			return orderID
		}

		orderA := setupCart(tokenA)
		orderB := setupCart(tokenB)

		type result struct {
			status int
		}
		results := make([]result, 2)
		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			key := fmt.Sprintf("conc-a-%d", time.Now().UnixNano())
			resp := checkoutWithKey(t, env, orderA, tokenA, key)
			drainClose(resp)
			results[0] = result{status: resp.StatusCode}
		}()
		go func() {
			defer wg.Done()
			key := fmt.Sprintf("conc-b-%d", time.Now().UnixNano())
			resp := checkoutWithKey(t, env, orderB, tokenB, key)
			drainClose(resp)
			results[1] = result{status: resp.StatusCode}
		}()
		wg.Wait()

		statusA := results[0].status
		statusB := results[1].status

		// Exactly one 200 and one 409.
		assert.True(t,
			(statusA == http.StatusOK && statusB == http.StatusConflict) ||
				(statusA == http.StatusConflict && statusB == http.StatusOK),
			"one checkout must succeed (200) and the other must fail (409); got %d and %d", statusA, statusB,
		)

		var finalStock int
		require.NoError(t, env.pool.QueryRow(ctx,
			`SELECT stock FROM product WHERE id=$1`, productID,
		).Scan(&finalStock))
		assert.Equal(t, 0, finalStock, "final stock must be 0 after one successful checkout")
	})

	t.Run("FR-INT-19 duplicate gateway_ref rejected by DB UNIQUE constraint", func(t *testing.T) {
		userA := seedUser(t, env, "student", "active", false)
		userB := seedUser(t, env, "student", "active", false)

		var orderA, orderB string
		require.NoError(t, env.pool.QueryRow(ctx,
			`INSERT INTO orders (student_id, status, subtotal, total) VALUES ($1, 'cart', 50000, 50000) RETURNING id`, userA,
		).Scan(&orderA))
		require.NoError(t, env.pool.QueryRow(ctx,
			`INSERT INTO orders (student_id, status, subtotal, total) VALUES ($1, 'cart', 50000, 50000) RETURNING id`, userB,
		).Scan(&orderB))

		_, err := env.pool.Exec(ctx, `UPDATE orders SET gateway_ref='dup-ref-fr19' WHERE id=$1`, orderA)
		require.NoError(t, err)

		_, err = env.pool.Exec(ctx, `UPDATE orders SET gateway_ref='dup-ref-fr19' WHERE id=$1`, orderB)
		require.Error(t, err)
		require.Contains(t, err.Error(), "unique")
	})
}
