package server

import (
	"akademi-bimbel/internal/handler"
	"akademi-bimbel/internal/platform"
	"akademi-bimbel/internal/service"

	"github.com/labstack/echo/v4"
)

// JWTMiddleware delegates to handler.JWTMiddleware.
func JWTMiddleware(svc *service.Service, signer *platform.JWTSigner) echo.MiddlewareFunc {
	return handler.JWTMiddleware(svc, signer)
}

// RBACMiddleware delegates to handler.RBACMiddleware.
func RBACMiddleware(required string) echo.MiddlewareFunc {
	return handler.RBACMiddleware(required)
}

// ClaimsFromContext delegates to handler.ClaimsFromContext.
func ClaimsFromContext(c echo.Context) *platform.Claims {
	return handler.ClaimsFromContext(c)
}
