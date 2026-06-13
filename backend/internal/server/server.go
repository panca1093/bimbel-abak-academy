package server

import (
	"akademi-bimbel/config"
	"log/slog"

	"akademi-bimbel/internal/handler"
	"akademi-bimbel/internal/platform"
	"akademi-bimbel/internal/service"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func New(h *handler.Handler, svc *service.Service, jwtSigner *platform.JWTSigner, cfg config.Config) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: cfg.CORSOrigins,
	}))
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus: true,
		LogURI:    true,
		LogMethod: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			slog.Info("request", "method", v.Method, "uri", v.URI, "status", v.Status)
			return nil
		},
	}))

	registerRoutes(e, h, svc, jwtSigner)
	return e
}
