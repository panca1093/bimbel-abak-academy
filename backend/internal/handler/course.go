package handler

import (
	"net/http"

	"akademi-bimbel/internal/platform"
	"github.com/labstack/echo/v4"
)

func (h *Handler) AdminListSections(c echo.Context) error {
	productID := c.Param("id")
	sections, err := h.svc.ListSections(c.Request().Context(), productID)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": sections,
	})
}

func (h *Handler) AdminCreateSection(c echo.Context) error {
	productID := c.Param("id")
	var req struct {
		Title string `json:"title"`
	}
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid request body")
	}
	if req.Title == "" {
		return badRequest(c, "title is required")
	}

	claims, _ := c.Get("claims").(*platform.Claims)
	role := ""
	if claims != nil {
		role = claims.Role
	}

	created, err := h.svc.CreateSection(c.Request().Context(), productID, req.Title, role)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusCreated, created)
}

func (h *Handler) AdminUpdateSection(c echo.Context) error {
	productID := c.Param("id")
	sectionID := c.Param("sId")
	var req struct {
		Title string `json:"title"`
	}
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid request body")
	}

	claims, _ := c.Get("claims").(*platform.Claims)
	role := ""
	if claims != nil {
		role = claims.Role
	}

	updated, err := h.svc.UpdateSection(c.Request().Context(), productID, sectionID, req.Title, role)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, updated)
}

func (h *Handler) AdminDeleteSection(c echo.Context) error {
	productID := c.Param("id")
	sectionID := c.Param("sId")
	claims, _ := c.Get("claims").(*platform.Claims)
	role := ""
	if claims != nil {
		role = claims.Role
	}

	err := h.svc.DeleteSection(c.Request().Context(), productID, sectionID, role)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) AdminReorderSections(c echo.Context) error {
	productID := c.Param("id")
	var req struct {
		SectionIDs []string `json:"section_ids"`
	}
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid request body")
	}

	claims, _ := c.Get("claims").(*platform.Claims)
	role := ""
	if claims != nil {
		role = claims.Role
	}

	err := h.svc.ReorderSections(c.Request().Context(), productID, req.SectionIDs, role)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "sections reordered",
	})
}

func (h *Handler) AdminCreateLesson(c echo.Context) error {
	productID := c.Param("id")
	sectionID := c.Param("sId")
	var req struct {
		Title    string `json:"title"`
		VideoURL string `json:"video_url"`
		Duration int    `json:"duration"`
	}
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid request body")
	}
	if req.Title == "" {
		return badRequest(c, "title is required")
	}

	claims, _ := c.Get("claims").(*platform.Claims)
	role := ""
	if claims != nil {
		role = claims.Role
	}

	created, err := h.svc.CreateLesson(c.Request().Context(), productID, sectionID, req.Title, req.VideoURL, req.Duration, role)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusCreated, created)
}

func (h *Handler) AdminUpdateLesson(c echo.Context) error {
	productID := c.Param("id")
	sectionID := c.Param("sId")
	lessonID := c.Param("lId")
	var req struct {
		Title    string `json:"title"`
		VideoURL string `json:"video_url"`
		Duration int    `json:"duration"`
	}
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid request body")
	}

	claims, _ := c.Get("claims").(*platform.Claims)
	role := ""
	if claims != nil {
		role = claims.Role
	}

	updated, err := h.svc.UpdateLesson(c.Request().Context(), productID, sectionID, lessonID, req.Title, req.VideoURL, req.Duration, role)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, updated)
}

func (h *Handler) AdminDeleteLesson(c echo.Context) error {
	productID := c.Param("id")
	sectionID := c.Param("sId")
	lessonID := c.Param("lId")
	claims, _ := c.Get("claims").(*platform.Claims)
	role := ""
	if claims != nil {
		role = claims.Role
	}

	err := h.svc.DeleteLesson(c.Request().Context(), productID, sectionID, lessonID, role)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) AdminReorderLessons(c echo.Context) error {
	productID := c.Param("id")
	sectionID := c.Param("sId")
	var req struct {
		LessonIDs []string `json:"lesson_ids"`
	}
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid request body")
	}

	claims, _ := c.Get("claims").(*platform.Claims)
	role := ""
	if claims != nil {
		role = claims.Role
	}

	err := h.svc.ReorderLessons(c.Request().Context(), productID, sectionID, req.LessonIDs, role)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "lessons reordered",
	})
}
