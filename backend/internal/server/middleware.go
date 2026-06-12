package server

import (
	"net/http"
	"strings"

	"akademi-bimbel/internal/platform"
	"akademi-bimbel/internal/service"

	"github.com/labstack/echo/v4"
)

const claimsKey = "claims"

func JWTMiddleware(svc *service.Service, signer *platform.JWTSigner) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			header := c.Request().Header.Get("Authorization")
			token, ok := strings.CutPrefix(header, "Bearer ")
			if !ok || token == "" {
				return errorJSON(c, http.StatusUnauthorized, "unauthorized", "missing or invalid token")
			}

			claims, err := signer.ParseAccess(token)
			if err != nil {
				return errorJSON(c, http.StatusUnauthorized, "unauthorized", "missing or invalid token")
			}

			if !svc.SessionActive(c.Request().Context(), claims.ID) {
				return errorJSON(c, http.StatusUnauthorized, "unauthorized", "missing or invalid token")
			}

			c.Set(claimsKey, claims)
			return next(c)
		}
	}
}

func RBACMiddleware(required string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims, _ := c.Get(claimsKey).(*platform.Claims)
			if claims == nil {
				return errorJSON(c, http.StatusUnauthorized, "unauthorized", "missing or invalid token")
			}
			if !service.HasCapability(claims.Role, required) {
				return errorJSON(c, http.StatusForbidden, "forbidden", "insufficient permissions")
			}
			return next(c)
		}
	}
}

func ClaimsFromContext(c echo.Context) *platform.Claims {
	claims, _ := c.Get(claimsKey).(*platform.Claims)
	return claims
}

func errorJSON(c echo.Context, status int, code, message string) error {
	return c.JSON(status, map[string]string{"code": code, "message": message})
}
