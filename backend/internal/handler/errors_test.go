package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"akademi-bimbel/internal/service"

	"github.com/labstack/echo/v4"
)

func TestMapServiceError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
		wantDetail any
	}{
		{name: "email taken", err: service.ErrEmailTaken, wantStatus: 409, wantCode: "email_taken"},
		{name: "invalid credentials", err: service.ErrInvalidCredentials, wantStatus: 401, wantCode: "invalid_credentials"},
		{name: "otp rate limit", err: service.ErrOTPRateLimit, wantStatus: 429, wantCode: "rate_limited"},
		{name: "otp expired", err: service.ErrOTPExpired, wantStatus: 422, wantCode: "otp_invalid"},
		{name: "invalid otp", err: service.ErrInvalidOTP, wantStatus: 422, wantCode: "otp_invalid"},
		{name: "invalid pending token", err: service.ErrInvalidPendingToken, wantStatus: 422, wantCode: "otp_invalid"},
		{name: "invalid refresh token", err: service.ErrInvalidRefreshToken, wantStatus: 401, wantCode: "invalid_refresh_token"},
		{name: "user not found", err: service.ErrUserNotFound, wantStatus: 404, wantCode: "user_not_found"},
		{name: "invalid reset token", err: service.ErrInvalidResetToken, wantStatus: 422, wantCode: "reset_invalid"},
		{name: "account deactivated", err: service.ErrAccountDeactivated, wantStatus: 403, wantCode: "account_deactivated"},
		{name: "invalid google token", err: service.ErrInvalidToken, wantStatus: 401, wantCode: "invalid_google_token"},
		{name: "invalid uuid", err: service.ErrInvalidUUID, wantStatus: 400, wantCode: "invalid_request"},
		{name: "weak password", err: service.ErrWeakPassword, wantStatus: 400, wantCode: "invalid_request", wantDetail: "password must be at least 8 characters"},
		{name: "product not found", err: service.ErrProductNotFound, wantStatus: 404, wantCode: "product_not_found"},
		{name: "course not found", err: service.ErrCourseNotFound, wantStatus: 404, wantCode: "course_not_found"},
		{name: "forbidden", err: service.ErrForbidden, wantStatus: 403, wantCode: "forbidden"},
		{name: "out of stock", err: service.ErrOutOfStock, wantStatus: 409, wantCode: "out_of_stock"},
		{name: "insufficient stock", err: service.ErrInsufficientStock, wantStatus: 409, wantCode: "insufficient_stock"},
		{name: "order not editable", err: service.ErrOrderNotEditable, wantStatus: 409, wantCode: "order_not_editable"},
		{name: "order not found", err: service.ErrOrderNotFound, wantStatus: 404, wantCode: "order_not_found"},
		{name: "invalid promo", err: service.ErrInvalidPromo, wantStatus: 422, wantCode: "invalid_promo"},
		{name: "promo min order", err: service.ErrPromoMinOrder, wantStatus: 422, wantCode: "promo_min_order"},
		{name: "invalid signature", err: service.ErrInvalidSignature, wantStatus: 401, wantCode: "invalid_signature"},
		{name: "course link required", err: service.ErrCourseLinkRequired, wantStatus: 422, wantCode: "course_required"},
		{name: "no course access", err: service.ErrNoCourseAccess, wantStatus: 422, wantCode: "no_course_access"},
		{name: "unknown error falls to 500", err: errors.New("something unexpected"), wantStatus: 500, wantCode: "internal_error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			_ = mapServiceError(c, tt.err)

			if rec.Code != tt.wantStatus {
				t.Errorf("status: want %d, got %d", tt.wantStatus, rec.Code)
			}

			var apiErr APIError
			if err := json.NewDecoder(rec.Body).Decode(&apiErr); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if apiErr.Code != tt.wantCode {
				t.Errorf("code: want %q, got %q", tt.wantCode, apiErr.Code)
			}
			if tt.wantDetail != nil {
				if apiErr.Details != tt.wantDetail {
					t.Errorf("details: want %v, got %v", tt.wantDetail, apiErr.Details)
				}
			}
		})
	}
}

func TestMapServiceError_OTPRateLimit_SetsRetryAfter(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_ = mapServiceError(c, service.ErrOTPRateLimit)

	if h := rec.Header().Get("Retry-After"); h != "60" {
		t.Errorf("Retry-After header: want 60, got %q", h)
	}
}

func TestBadRequest(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_ = badRequest(c, "something is wrong")

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status: want 400, got %d", rec.Code)
	}

	var apiErr APIError
	if err := json.NewDecoder(rec.Body).Decode(&apiErr); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if apiErr.Code != "invalid_request" {
		t.Errorf("code: want %q, got %q", "invalid_request", apiErr.Code)
	}
	if apiErr.Message != "something is wrong" {
		t.Errorf("message: want %q, got %q", "something is wrong", apiErr.Message)
	}
}
