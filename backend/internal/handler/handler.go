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
