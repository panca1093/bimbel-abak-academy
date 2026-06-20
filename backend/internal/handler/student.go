package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *Handler) ListSchools(c echo.Context) error {
	schools, err := h.svc.ListSchools(c.Request().Context())
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(http.StatusOK, schools)
}

func (h *Handler) StudentProfile(c echo.Context) error {
	claims := claimsFromContext(c)
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
	claims := claimsFromContext(c)
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
		Grade      *int    `json:"grade"`
		SchoolID   *string `json:"school_id"`
	}
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid request body")
	}
	user, err := h.svc.UpdateProfile(
		c.Request().Context(),
		claims.Sub,
		req.Name,
		req.Email,
		req.Username,
		req.Phone,
		req.Address,
		req.TargetExam,
		req.Grade,
		req.SchoolID,
	)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(http.StatusOK, user)
}

func (h *Handler) GeneratePresignUploadURL(c echo.Context) error {
	claims := claimsFromContext(c)
	if claims == nil || claims.Sub == "" {
		return c.JSON(http.StatusUnauthorized, APIError{Code: "unauthorized", Message: "missing auth"})
	}
	filename := c.QueryParam("filename")
	contentType := c.QueryParam("content_type")
	if filename == "" {
		return badRequest(c, "filename is required")
	}
	resp, err := h.svc.GeneratePresignedUploadURL(c.Request().Context(), claims.Sub, filename, contentType)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) UpdatePhoto(c echo.Context) error {
	claims := claimsFromContext(c)
	if claims == nil || claims.Sub == "" {
		return c.JSON(http.StatusUnauthorized, APIError{Code: "unauthorized", Message: "missing auth"})
	}
	var req struct {
		PhotoURL string `json:"photo_url"`
	}
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid request body")
	}
	if req.PhotoURL == "" {
		return badRequest(c, "photo_url is required")
	}
	user, err := h.svc.UpdatePhoto(c.Request().Context(), claims.Sub, req.PhotoURL)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(http.StatusOK, user)
}
