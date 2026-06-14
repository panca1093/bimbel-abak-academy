package handler

import (
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *Handler) HandlePaymentWebhook(c echo.Context) error {
	key := c.Request().Header.Get("Idempotency-Key")
	if key == "" {
		return badRequest(c, "Idempotency-Key header is required")
	}

	signature := c.Request().Header.Get("X-Signature")
	if signature == "" {
		return c.JSON(http.StatusUnauthorized, APIError{Code: "invalid_signature", Message: "missing signature"})
	}

	payload, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return badRequest(c, "invalid request body")
	}

	err = h.svc.HandlePaymentWebhook(c.Request().Context(), payload, signature, key)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "webhook processed",
	})
}
