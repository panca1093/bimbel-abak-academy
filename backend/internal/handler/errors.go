package handler

import (
	"errors"
	"net/http"

	"akademi-bimbel/internal/service"
	"github.com/labstack/echo/v4"
)

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func mapServiceError(c echo.Context, err error) error {
	var status int
	var apiErr APIError
	switch {
	case errors.Is(err, service.ErrEmailTaken):
		status, apiErr = http.StatusConflict, APIError{Code: "email_taken", Message: err.Error()}
	case errors.Is(err, service.ErrInvalidCredentials):
		status, apiErr = http.StatusUnauthorized, APIError{Code: "invalid_credentials", Message: err.Error()}
	case errors.Is(err, service.ErrOTPRateLimit):
		c.Response().Header().Set("Retry-After", "60")
		status, apiErr = http.StatusTooManyRequests, APIError{Code: "rate_limited", Message: err.Error()}
	case errors.Is(err, service.ErrOTPExpired):
		status, apiErr = http.StatusUnprocessableEntity, APIError{Code: "otp_invalid", Message: err.Error()}
	case errors.Is(err, service.ErrInvalidOTP):
		status, apiErr = http.StatusUnprocessableEntity, APIError{Code: "otp_invalid", Message: err.Error()}
	case errors.Is(err, service.ErrInvalidPendingToken):
		status, apiErr = http.StatusUnprocessableEntity, APIError{Code: "otp_invalid", Message: err.Error()}
	case errors.Is(err, service.ErrInvalidRefreshToken):
		status, apiErr = http.StatusUnauthorized, APIError{Code: "invalid_refresh_token", Message: err.Error()}
	case errors.Is(err, service.ErrUserNotFound):
		status, apiErr = http.StatusNotFound, APIError{Code: "user_not_found", Message: err.Error()}
	case errors.Is(err, service.ErrInvalidResetToken):
		status, apiErr = http.StatusUnprocessableEntity, APIError{Code: "reset_invalid", Message: err.Error()}
	case errors.Is(err, service.ErrAccountDeactivated):
		status, apiErr = http.StatusForbidden, APIError{Code: "account_deactivated", Message: err.Error()}
	case errors.Is(err, service.ErrInvalidToken):
		status, apiErr = http.StatusUnauthorized, APIError{Code: "invalid_google_token", Message: "invalid or expired Google token"}
	case errors.Is(err, service.ErrWeakPassword):
		status, apiErr = http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error(), Details: "password must be at least 8 characters"}
	case errors.Is(err, service.ErrProductNotFound):
		status, apiErr = http.StatusNotFound, APIError{Code: "product_not_found", Message: err.Error()}
	case errors.Is(err, service.ErrForbidden):
		status, apiErr = http.StatusForbidden, APIError{Code: "forbidden", Message: err.Error()}
	case errors.Is(err, service.ErrOutOfStock):
		status, apiErr = http.StatusConflict, APIError{Code: "out_of_stock", Message: err.Error()}
	case errors.Is(err, service.ErrInsufficientStock):
		status, apiErr = http.StatusConflict, APIError{Code: "insufficient_stock", Message: err.Error()}
	case errors.Is(err, service.ErrOrderNotEditable):
		status, apiErr = http.StatusConflict, APIError{Code: "order_not_editable", Message: err.Error()}
	case errors.Is(err, service.ErrOrderNotFound):
		status, apiErr = http.StatusNotFound, APIError{Code: "order_not_found", Message: err.Error()}
	case errors.Is(err, service.ErrInvalidPromo):
		status, apiErr = http.StatusUnprocessableEntity, APIError{Code: "invalid_promo", Message: err.Error()}
	case errors.Is(err, service.ErrPromoMinOrder):
		status, apiErr = http.StatusUnprocessableEntity, APIError{Code: "promo_min_order", Message: err.Error()}
	case errors.Is(err, service.ErrInvalidSignature):
		status, apiErr = http.StatusUnauthorized, APIError{Code: "invalid_signature", Message: err.Error()}
	default:
		status, apiErr = http.StatusInternalServerError, APIError{Code: "internal_error", Message: "internal server error"}
	}
	return c.JSON(status, apiErr)
}

func badRequest(c echo.Context, msg string) error {
	return c.JSON(http.StatusBadRequest, APIError{Code: "invalid_request", Message: msg})
}
