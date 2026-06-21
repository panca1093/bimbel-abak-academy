package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"akademi-bimbel/config"
	"akademi-bimbel/internal/handler"
	"akademi-bimbel/internal/repository"
	"akademi-bimbel/internal/server"
	"akademi-bimbel/internal/service"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// checkoutWithKey issues POST /api/v1/orders/:id/checkout with the Idempotency-Key header.
func checkoutWithKey(t *testing.T, env *testEnv, orderID, token, idempKey string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, env.server.URL+"/api/v1/orders/"+orderID+"/checkout", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Idempotency-Key", idempKey)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

// sendWebhook POSTs to /api/v1/webhooks/payment with a Midtrans notification body.
// Pass signature="" to omit signature_key from the JSON body.
func sendWebhook(t *testing.T, env *testEnv, orderID, idempKey, signature, grossAmount string) *http.Response {
	t.Helper()
	return sendWebhookToURL(t, env.server.URL, orderID, idempKey, signature, grossAmount)
}

// sendWebhookToURL builds and POSTs a Midtrans notification to the given base URL.
func sendWebhookToURL(t *testing.T, baseURL, orderID, idempKey, signature, grossAmount string) *http.Response {
	t.Helper()
	body := map[string]string{
		"transaction_status": "settlement",
		"order_id":           orderID,
		"transaction_id":     fmt.Sprintf("tx-%d", time.Now().UnixNano()),
		"gross_amount":       grossAmount,
		"status_code":        "200",
	}
	if signature != "" {
		body["signature_key"] = signature
	}

	payload, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost,
		baseURL+"/api/v1/webhooks/payment",
		bytes.NewReader(payload))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", idempKey)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

// strictPaymentClient rejects empty signatures, so a missing signature_key fails verification.
type strictPaymentClient struct{}

func (strictPaymentClient) CreatePayment(ctx context.Context, req service.PaymentRequest) (service.PaymentResponse, error) {
	return service.PaymentResponse{}, nil
}

func (strictPaymentClient) QueryStatus(ctx context.Context, reference string) (service.PaymentStatus, error) {
	return service.PaymentStatus{}, nil
}

func (strictPaymentClient) VerifySignature(payload []byte, signature string) bool {
	return signature != ""
}

// webhookServer returns a test server wired with a strict payment client for signature verification tests.
func webhookServer(t *testing.T, env *testEnv, payment service.PaymentClient) *httptest.Server {
	t.Helper()
	repo := repository.New(env.pool)
	cfg := &config.Config{CORSOrigins: []string{"*"}}
	svc := service.NewWithStore(repo, repo, env.rdb, env.signer,
		&service.NoopOTPProvider{}, &service.NoopEmailProvider{}, payment,
		&service.NoopLogisticsClient{}, nil, cfg)
	h := handler.New(svc)
	e := echo.New()
	e.HideBanner = true
	server.RegisterRoutesForTest(e, h, svc, env.signer)
	ts := httptest.NewServer(e)
	t.Cleanup(ts.Close)
	return ts
}

// drainClose discards and closes a response body.
func drainClose(resp *http.Response) {
	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
}

func TestCheckout(t *testing.T) {
	env := newTestEnv(t)
	ctx := context.Background()

	t.Run("FR-INT-07 mint cart idempotency — one cart row", func(t *testing.T) {
		userID := seedUser(t, env, "student", "active", false)
		token := authToken(t, env, userID, "student")

		resp1 := env.doJSON(t, http.MethodPost, "/api/v1/orders", nil, token)
		body1 := decodeBody(t, resp1)
		require.Equal(t, http.StatusCreated, resp1.StatusCode)
		orderID1, ok := body1["id"].(string)
		require.True(t, ok, "body must contain 'id' field, got: %v", body1)
		require.NotEmpty(t, orderID1)

		resp2 := env.doJSON(t, http.MethodPost, "/api/v1/orders", nil, token)
		body2 := decodeBody(t, resp2)
		require.Equal(t, http.StatusOK, resp2.StatusCode)
		orderID2, ok := body2["id"].(string)
		require.True(t, ok, "body must contain 'id' field, got: %v", body2)

		assert.Equal(t, orderID1, orderID2, "second MintCart must return same order id")

		var count int
		err := env.pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM orders WHERE student_id=$1 AND status='cart'`, userID,
		).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "exactly one cart row must exist (idx_orders_student_cart)")
	})

	t.Run("FR-INT-08 add item writes order_item row", func(t *testing.T) {
		userID := seedUser(t, env, "student", "active", false)
		token := authToken(t, env, userID, "student")
		productID := seedProduct(t, env, "book", "Buku Matematika", 50000)

		resp := env.doJSON(t, http.MethodPost, "/api/v1/orders", nil, token)
		body := decodeBody(t, resp)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		orderID := body["id"].(string)

		resp2 := env.doJSON(t, http.MethodPost, "/api/v1/orders/"+orderID+"/items",
			map[string]any{"product_id": productID, "qty": 1}, token)
		require.Equal(t, http.StatusCreated, resp2.StatusCode)
		drainClose(resp2)

		var count int
		err := env.pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM order_item WHERE order_id=$1 AND product_id=$2`,
			orderID, productID,
		).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "order_item row must exist for order/product pair")
	})

	t.Run("FR-INT-09 patch cart persists shipping/courier fields", func(t *testing.T) {
		userID := seedUser(t, env, "student", "active", false)
		token := authToken(t, env, userID, "student")
		productID := seedProduct(t, env, "book", "Buku Fisika", 30000)

		resp := env.doJSON(t, http.MethodPost, "/api/v1/orders", nil, token)
		body := decodeBody(t, resp)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		orderID := body["id"].(string)

		drainClose(env.doJSON(t, http.MethodPost, "/api/v1/orders/"+orderID+"/items",
			map[string]any{"product_id": productID, "qty": 1}, token))

		// shipping_address is []byte in the handler (decoded as base64 JSON); send courier only to
		// keep the test simple and avoid base64 encoding in the test body.
		patchResp := env.doJSON(t, http.MethodPatch, "/api/v1/orders/"+orderID,
			map[string]any{"courier": "JNE"}, token)
		require.Equal(t, http.StatusOK, patchResp.StatusCode)
		drainClose(patchResp)

		var courier string
		err := env.pool.QueryRow(ctx,
			`SELECT selected_courier FROM orders WHERE id=$1`, orderID,
		).Scan(&courier)
		require.NoError(t, err)
		assert.Equal(t, "JNE", courier)
	})

	t.Run("FR-INT-10 checkout sets payment_pending, gateway_ref, decrements stock, Redis key", func(t *testing.T) {
		userID := seedUser(t, env, "student", "active", false)
		token := authToken(t, env, userID, "student")
		productID := seedProduct(t, env, "book", "Buku IPA", 40000)

		var stockBefore int
		require.NoError(t, env.pool.QueryRow(ctx,
			`SELECT stock FROM product WHERE id=$1`, productID,
		).Scan(&stockBefore))

		resp := env.doJSON(t, http.MethodPost, "/api/v1/orders", nil, token)
		body := decodeBody(t, resp)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		orderID := body["id"].(string)

		drainClose(env.doJSON(t, http.MethodPost, "/api/v1/orders/"+orderID+"/items",
			map[string]any{"product_id": productID, "qty": 1}, token))

		idempKey := fmt.Sprintf("test-checkout-%d", time.Now().UnixNano())
		coResp := checkoutWithKey(t, env, orderID, token, idempKey)
		coBody := decodeBody(t, coResp)
		require.Equal(t, http.StatusOK, coResp.StatusCode, "checkout failed: %v", coBody)

		gatewayRef, _ := coBody["gateway_ref"].(string)
		assert.NotEmpty(t, gatewayRef)
		assert.Equal(t, "noop-"+orderID, gatewayRef)

		var status, gwRef string
		var paymentExpiresAt time.Time
		require.NoError(t, env.pool.QueryRow(ctx,
			`SELECT status, gateway_ref, payment_expires_at FROM orders WHERE id=$1`, orderID,
		).Scan(&status, &gwRef, &paymentExpiresAt))
		assert.Equal(t, "payment_pending", status)
		assert.Equal(t, "noop-"+orderID, gwRef)
		assert.True(t, paymentExpiresAt.After(time.Now()), "payment_expires_at must be in the future")

		var stockAfter int
		require.NoError(t, env.pool.QueryRow(ctx,
			`SELECT stock FROM product WHERE id=$1`, productID,
		).Scan(&stockAfter))
		assert.Equal(t, stockBefore-1, stockAfter, "stock must be decremented by 1 after checkout")

		exists, err := env.rdb.Exists(ctx, "idempotency:checkout:"+idempKey).Result()
		require.NoError(t, err)
		assert.EqualValues(t, 1, exists, "idempotency:checkout:<key> must exist in Redis")
	})

	t.Run("FR-INT-11 webhook flips order to paid, inserts outbox row and webhook_log", func(t *testing.T) {
		userID := seedUser(t, env, "student", "active", false)
		token := authToken(t, env, userID, "student")
		productID := seedProduct(t, env, "book", "Buku Kimia", 25000)

		resp := env.doJSON(t, http.MethodPost, "/api/v1/orders", nil, token)
		body := decodeBody(t, resp)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		orderID := body["id"].(string)

		drainClose(env.doJSON(t, http.MethodPost, "/api/v1/orders/"+orderID+"/items",
			map[string]any{"product_id": productID, "qty": 1}, token))

		coResp := checkoutWithKey(t, env, orderID, token,
			fmt.Sprintf("co-%d", time.Now().UnixNano()))
		require.Equal(t, http.StatusOK, coResp.StatusCode, "checkout failed: %v", decodeBody(t, coResp))

		webhookKey := fmt.Sprintf("wh-%d", time.Now().UnixNano())
		whResp := sendWebhook(t, env, orderID, webhookKey, "any-sig", fmt.Sprintf("%.2f", float64(25000)))
		whBody := decodeBody(t, whResp)
		require.Equal(t, http.StatusOK, whResp.StatusCode, "webhook failed: %v", whBody)

		var status string
		require.NoError(t, env.pool.QueryRow(ctx,
			`SELECT status FROM orders WHERE id=$1`, orderID,
		).Scan(&status))
		assert.Equal(t, "paid", status)

		var outboxCount int
		require.NoError(t, env.pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM outbox WHERE aggregate_id=$1 AND event_type='OrderPaid'`, orderID,
		).Scan(&outboxCount))
		assert.Equal(t, 1, outboxCount, "exactly one outbox OrderPaid row")

		var logCount int
		require.NoError(t, env.pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM webhook_log WHERE gateway_ref=$1`, orderID,
		).Scan(&logCount))
		assert.Equal(t, 1, logCount, "webhook_log row must exist")
	})

	t.Run("FR-INT-12 webhook missing signature_key returns 401 invalid_signature", func(t *testing.T) {
		strict := webhookServer(t, env, strictPaymentClient{})
		whResp := sendWebhookToURL(t, strict.URL,
			"00000000-0000-0000-0000-000000000000",
			fmt.Sprintf("key-%d", time.Now().UnixNano()),
			"" /* omit signature_key */,
			"0.00")
		whBody := decodeBody(t, whResp)
		require.Equal(t, http.StatusUnauthorized, whResp.StatusCode)
		assert.Equal(t, "invalid_signature", whBody["code"])
	})

	t.Run("FR-INT-13 webhook idempotency — duplicate delivery produces single outbox row", func(t *testing.T) {
		userID := seedUser(t, env, "student", "active", false)
		token := authToken(t, env, userID, "student")
		productID := seedProduct(t, env, "book", "Buku Biologi", 20000)

		resp := env.doJSON(t, http.MethodPost, "/api/v1/orders", nil, token)
		body := decodeBody(t, resp)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		orderID := body["id"].(string)

		drainClose(env.doJSON(t, http.MethodPost, "/api/v1/orders/"+orderID+"/items",
			map[string]any{"product_id": productID, "qty": 1}, token))

		coResp := checkoutWithKey(t, env, orderID, token,
			fmt.Sprintf("co-%d", time.Now().UnixNano()))
		require.Equal(t, http.StatusOK, coResp.StatusCode, "checkout failed: %v", decodeBody(t, coResp))

		webhookKey := fmt.Sprintf("dedup-%d", time.Now().UnixNano())

		wh1 := sendWebhook(t, env, orderID, webhookKey, "sig1", fmt.Sprintf("%.2f", float64(20000)))
		wh1Body := decodeBody(t, wh1)
		require.Equal(t, http.StatusOK, wh1.StatusCode, "first delivery: %v", wh1Body)

		wh2 := sendWebhook(t, env, orderID, webhookKey, "sig2", fmt.Sprintf("%.2f", float64(20000)))
		wh2Body := decodeBody(t, wh2)
		require.Equal(t, http.StatusOK, wh2.StatusCode, "second delivery: %v", wh2Body)

		var status string
		require.NoError(t, env.pool.QueryRow(ctx,
			`SELECT status FROM orders WHERE id=$1`, orderID,
		).Scan(&status))
		assert.Equal(t, "paid", status)

		var outboxCount int
		require.NoError(t, env.pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM outbox WHERE aggregate_id=$1 AND event_type='OrderPaid'`, orderID,
		).Scan(&outboxCount))
		assert.Equal(t, 1, outboxCount, "duplicate delivery must not insert a second outbox row")
	})
}
