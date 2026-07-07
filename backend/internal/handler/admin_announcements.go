package handler

import (
	"net/http"
	"time"

	"akademi-bimbel/internal/infra"
	"github.com/labstack/echo/v4"
)

type createAnnouncementRequest struct {
	Title       string     `json:"title"`
	Message     string     `json:"message"`
	Type        string     `json:"type"`
	Recipients  string     `json:"recipients"`
	Status      string     `json:"status"`
	ScheduledAt *time.Time `json:"scheduled_at"`
}

type updateAnnouncementRequest struct {
	Title       string     `json:"title"`
	Message     string     `json:"message"`
	Type        string     `json:"type"`
	Recipients  string     `json:"recipients"`
	ScheduledAt *time.Time `json:"scheduled_at"`
}

func (h *Handler) AdminCreateAnnouncement(c echo.Context) error {
	claims, ok := c.Get("claims").(*infra.Claims)
	if !ok || claims == nil || claims.Sub == "" {
		return c.JSON(http.StatusUnauthorized, APIError{Code: "unauthorized", Message: "missing auth"})
	}

	var req createAnnouncementRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid request body")
	}
	if req.Title == "" {
		return badRequest(c, "title is required")
	}
	if req.Message == "" {
		return badRequest(c, "message is required")
	}

	a, err := h.svc.CreateAnnouncement(
		c.Request().Context(),
		claims.Sub, req.Title, req.Message, req.Type, req.Recipients, req.Status, req.ScheduledAt,
	)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(http.StatusCreated, a)
}

func (h *Handler) AdminListAnnouncements(c echo.Context) error {
	announcements, err := h.svc.ListAnnouncements(c.Request().Context())
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": announcements,
	})
}

func (h *Handler) AdminUpdateAnnouncement(c echo.Context) error {
	id := c.Param("id")

	var req updateAnnouncementRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid request body")
	}

	a, err := h.svc.UpdateAnnouncement(
		c.Request().Context(), id, req.Title, req.Message, req.Type, req.Recipients, req.ScheduledAt,
	)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(http.StatusOK, a)
}

func (h *Handler) AdminDeleteAnnouncement(c echo.Context) error {
	id := c.Param("id")
	err := h.svc.DeleteAnnouncement(c.Request().Context(), id)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "announcement deleted"})
}

func (h *Handler) AdminSendAnnouncement(c echo.Context) error {
	id := c.Param("id")
	a, err := h.svc.SendAnnouncementNow(c.Request().Context(), id)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(http.StatusOK, a)
}
