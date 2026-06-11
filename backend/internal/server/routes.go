package server

import (
	"akademi-bimbel/internal/handler"
	"github.com/labstack/echo/v4"
)

func registerRoutes(e *echo.Echo, h *handler.Handler) {
	v1 := e.Group("/api/v1")
	v1.GET("/health", h.Health)

	// Feature handlers (auth, exam, course, store, admin, webhooks/midtrans)
	// mount here as the TRD component diagram is built out.
}
