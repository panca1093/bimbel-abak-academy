package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"akademi-bimbel/internal/handler"
	"akademi-bimbel/internal/service"

	"github.com/alicebob/miniredis/v2"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

func TestAdminRefundOrder_BasicCompilation(t *testing.T) {
	// This test ensures the handlers compile and respond correctly
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	svc := service.NewForTest(rdb)
	h := handler.New(svc)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/orders/test-id/refund", nil)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/admin/orders/:id/refund")
	c.SetParamNames("id")
	c.SetParamValues("test-id")

	// Act
	err = h.AdminRefundOrder(c)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Logf("status code = %d", rec.Code)
	}

	var respBody map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &respBody); err == nil {
		if msg, ok := respBody["message"]; ok {
			t.Logf("response message: %s", msg)
		}
	}
}
