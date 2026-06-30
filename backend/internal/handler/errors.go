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
	case errors.Is(err, service.ErrInvalidUUID):
		status, apiErr = http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()}
	case errors.Is(err, service.ErrWeakPassword):
		status, apiErr = http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error(), Details: "password must be at least 8 characters"}
	case errors.Is(err, service.ErrProductNotFound):
		status, apiErr = http.StatusNotFound, APIError{Code: "product_not_found", Message: err.Error()}
	case errors.Is(err, service.ErrCourseNotFound):
		status, apiErr = http.StatusNotFound, APIError{Code: "course_not_found", Message: err.Error()}
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
	case errors.Is(err, service.ErrTestNotFound):
		status, apiErr = http.StatusNotFound, APIError{Code: "test_not_found", Message: err.Error()}
	case errors.Is(err, service.ErrQuestionNotFound):
		status, apiErr = http.StatusNotFound, APIError{Code: "question_not_found", Message: err.Error()}
	case errors.Is(err, service.ErrExamNotFound):
		status, apiErr = http.StatusNotFound, APIError{Code: "exam_not_found", Message: err.Error()}
	case errors.Is(err, service.ErrRegistrationNotFound):
		status, apiErr = http.StatusNotFound, APIError{Code: "registration_not_found", Message: err.Error()}
	case errors.Is(err, service.ErrValidation):
		status, apiErr = http.StatusUnprocessableEntity, APIError{Code: "validation_failed", Message: err.Error()}
	case errors.Is(err, service.ErrInvalidPromo):
		status, apiErr = http.StatusUnprocessableEntity, APIError{Code: "invalid_promo", Message: err.Error()}
	case errors.Is(err, service.ErrPromoMinOrder):
		status, apiErr = http.StatusUnprocessableEntity, APIError{Code: "promo_min_order", Message: err.Error()}
	case errors.Is(err, service.ErrInvalidSignature):
		status, apiErr = http.StatusUnauthorized, APIError{Code: "invalid_signature", Message: err.Error()}
	case errors.Is(err, service.ErrCourseLinkRequired):
		status, apiErr = http.StatusUnprocessableEntity, APIError{Code: "course_required", Message: err.Error()}
	case errors.Is(err, service.ErrNoCourseAccess):
		status, apiErr = http.StatusUnprocessableEntity, APIError{Code: "no_course_access", Message: err.Error()}
	case errors.Is(err, service.ErrUnknownConfigKey):
		status, apiErr = http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()}
	case errors.Is(err, service.ErrConfigEncryption):
		status, apiErr = http.StatusInternalServerError, APIError{Code: "internal_error", Message: "config encryption failed"}
	case errors.Is(err, service.ErrCannotDeactivateSelf):
		status, apiErr = http.StatusForbidden, APIError{Code: "cannot_deactivate_self", Message: err.Error()}
	case errors.Is(err, service.ErrInvalidAdminRole):
		status, apiErr = http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()}
	case errors.Is(err, service.ErrInvalidRoleFilter):
		status, apiErr = http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()}
	case errors.Is(err, service.ErrInvalidStatusFilter):
		status, apiErr = http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()}
	case errors.Is(err, service.ErrAccountNoEmail):
		status, apiErr = http.StatusUnprocessableEntity, APIError{Code: "reset_invalid", Message: err.Error()}
	case errors.Is(err, service.ErrMissingField):
		status, apiErr = http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()}
	case errors.Is(err, service.ErrInvalidDateFormat):
		status, apiErr = http.StatusBadRequest, APIError{Code: "invalid_request", Message: err.Error()}
	case errors.Is(err, service.ErrAlreadyAttempted):
		status, apiErr = http.StatusConflict, APIError{Code: "already_attempted", Message: err.Error()}
	case errors.Is(err, service.ErrExamNotStarted):
		status, apiErr = http.StatusConflict, APIError{Code: "exam_not_started", Message: err.Error()}
	case errors.Is(err, service.ErrDeviceMismatch):
		status, apiErr = http.StatusForbidden, APIError{Code: "device_mismatch", Message: err.Error()}
	case errors.Is(err, service.ErrCheckinWindowClosed):
		status, apiErr = http.StatusConflict, APIError{Code: "checkin_window_closed", Message: err.Error()}
	case errors.Is(err, service.ErrNotCheckedIn):
		status, apiErr = http.StatusConflict, APIError{Code: "not_checked_in", Message: err.Error()}
	case errors.Is(err, service.ErrAlreadySubmitted):
		status, apiErr = http.StatusConflict, APIError{Code: "already_submitted", Message: err.Error()}
	case errors.Is(err, service.ErrSessionNotFound):
		status, apiErr = http.StatusNotFound, APIError{Code: "session_not_found", Message: err.Error()}
	case errors.Is(err, service.ErrInvalidViolationType):
		status, apiErr = http.StatusBadRequest, APIError{Code: "invalid_violation_type", Message: err.Error()}

	default:
		status, apiErr = http.StatusInternalServerError, APIError{Code: "internal_error", Message: "internal server error"}
	}
	return c.JSON(status, apiErr)
}

func badRequest(c echo.Context, msg string) error {
	return c.JSON(http.StatusBadRequest, APIError{Code: "invalid_request", Message: msg})
}
