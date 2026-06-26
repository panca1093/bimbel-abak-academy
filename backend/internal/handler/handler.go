package handler

import (
	"net/http"

	"akademi-bimbel/internal/service"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	svc *service.Service
}

func New(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

// Health threads handler -> service -> repository/cache to prove the layering.
func (h *Handler) Health(c echo.Context) error {
	hc := h.svc.Health(c.Request().Context())
	status := http.StatusOK
	if hc.Status != "ok" {
		status = http.StatusServiceUnavailable
	}
	return c.JSON(status, hc)
}

// GetPaymentClientKey returns the Midtrans client key from DB config.
// Public endpoint — client key is safe to expose (Midtrans docs confirm it is public).
func (h *Handler) GetPaymentClientKey(c echo.Context) error {
	clientKey, err := h.svc.GetPaymentClientKey(c.Request().Context())
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]string{
		"client_key": clientKey,
	})
}
