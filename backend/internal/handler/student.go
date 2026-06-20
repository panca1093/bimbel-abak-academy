package handler

import (
	"net/http"

	"akademi-bimbel/internal/infra"

	"github.com/labstack/echo/v4"
)

func (h *Handler) StudentDashboard(c echo.Context) error {
	claims, _ := c.Get("claims").(*infra.Claims)
	if claims == nil || claims.Sub == "" {
		return c.JSON(http.StatusUnauthorized, APIError{Code: "unauthorized", Message: "missing auth"})
	}
	dash, err := h.svc.GetDashboard(c.Request().Context(), claims.Sub)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(http.StatusOK, dash)
}

func (h *Handler) StudentProfile(c echo.Context) error {
	claims, _ := c.Get("claims").(*infra.Claims)
	if claims == nil || claims.Sub == "" {
		return c.JSON(http.StatusUnauthorized, APIError{Code: "unauthorized", Message: "missing auth"})
	}
	user, err := h.svc.Me(c.Request().Context(), claims.Sub)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(http.StatusOK, user)
}

func (h *Handler) StudentUpdateProfile(c echo.Context) error {
	claims, _ := c.Get("claims").(*infra.Claims)
	if claims == nil || claims.Sub == "" {
		return c.JSON(http.StatusUnauthorized, APIError{Code: "unauthorized", Message: "missing auth"})
	}
	var req struct {
		Name       *string `json:"name"`
		Email      *string `json:"email"`
		Username   *string `json:"username"`
		Phone      *string `json:"phone"`
		Address    *string `json:"address"`
		TargetExam *string `json:"target_exam"`
	}
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid request body")
	}
	user, err := h.svc.UpdateProfile(c.Request().Context(), claims.Sub, req.Name, req.Email, req.Username, req.Phone, req.Address, req.TargetExam)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(http.StatusOK, user)
}