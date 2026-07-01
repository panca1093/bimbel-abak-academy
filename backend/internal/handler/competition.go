package handler

import (
	"bytes"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *Handler) StudentListRegistrations(c echo.Context) error {
	claims := claimsFromContext(c)
	if claims == nil || claims.Sub == "" {
		return c.JSON(http.StatusUnauthorized, APIError{Code: "unauthorized", Message: "missing auth"})
	}
	items, err := h.svc.GetExamRegistrations(c.Request().Context(), claims.Sub)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": items})
}

func (h *Handler) StudentGetRegistration(c echo.Context) error {
	claims := claimsFromContext(c)
	if claims == nil || claims.Sub == "" {
		return c.JSON(http.StatusUnauthorized, APIError{Code: "unauthorized", Message: "missing auth"})
	}
	id := c.Param("id")
	detail, err := h.svc.GetExamRegistration(c.Request().Context(), id, claims.Sub)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(http.StatusOK, detail)
}

func (h *Handler) StudentGetExamCard(c echo.Context) error {
	claims := claimsFromContext(c)
	if claims == nil || claims.Sub == "" {
		return c.JSON(http.StatusUnauthorized, APIError{Code: "unauthorized", Message: "missing auth"})
	}
	id := c.Param("id")
	pdf, filename, err := h.svc.GetExamCard(c.Request().Context(), id, claims.Sub)
	if err != nil {
		return mapServiceError(c, err)
	}
	c.Response().Header().Set("Content-Type", "application/pdf")
	c.Response().Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	return c.Stream(http.StatusOK, "application/pdf", bytes.NewReader(pdf))
}