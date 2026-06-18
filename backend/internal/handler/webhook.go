package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
)

type midtransWebhookBody struct {
	SignatureKey string `json:"signature_key"`
}

func (h *Handler) HandlePaymentWebhook(c echo.Context) error {
	key := c.Request().Header.Get("Idempotency-Key")
	if key == "" {
		return badRequest(c, "Idempotency-Key header is required")
	}

	payload, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return badRequest(c, "invalid request body")
	}

	var body midtransWebhookBody
	if err := json.Unmarshal(payload, &body); err != nil {
		return badRequest(c, "invalid request body")
	}

	err = h.svc.HandlePaymentWebhook(c.Request().Context(), payload, body.SignatureKey, key)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "webhook processed",
	})
}
