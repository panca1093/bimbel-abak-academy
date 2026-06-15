package handler

import (
	"net/http"

	"akademi-bimbel/internal/infra"
	"github.com/labstack/echo/v4"
)

// --- Admin Course CRUD ---

func (h *Handler) AdminCreateCourse(c echo.Context) error {
	var req struct {
		Title          string `json:"title"`
		Level          string `json:"level"`
		Subject        string `json:"subject"`
		InstructorName string `json:"instructor_name"`
	}
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid request body")
	}
	if req.Title == "" {
		return badRequest(c, "title is required")
	}

	claims, _ := c.Get("claims").(*infra.Claims)
	role := ""
	if claims != nil {
		role = claims.Role
	}

	course, err := h.svc.CreateCourse(c.Request().Context(), req.Title, req.Level, req.Subject, req.InstructorName, role)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusCreated, course)
}

func (h *Handler) AdminListCourses(c echo.Context) error {
	claims, _ := c.Get("claims").(*infra.Claims)
	role := ""
	if claims != nil {
		role = claims.Role
	}

	courses, err := h.svc.ListCourses(c.Request().Context(), role)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": courses,
	})
}

func (h *Handler) AdminUpdateCourse(c echo.Context) error {
	courseID := c.Param("id")
	var req struct {
		Title          string `json:"title"`
		Level          string `json:"level"`
		Subject        string `json:"subject"`
		InstructorName string `json:"instructor_name"`
	}
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid request body")
	}

	claims, _ := c.Get("claims").(*infra.Claims)
	role := ""
	if claims != nil {
		role = claims.Role
	}

	course, err := h.svc.UpdateCourse(c.Request().Context(), courseID, req.Title, req.Level, req.Subject, req.InstructorName, role)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, course)
}

// --- Admin Section CRUD (re-keyed to course_id) ---

func (h *Handler) AdminListSections(c echo.Context) error {
	courseID := c.Param("id")
	sections, err := h.svc.ListSections(c.Request().Context(), courseID)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": sections,
	})
}

func (h *Handler) AdminCreateSection(c echo.Context) error {
	courseID := c.Param("id")
	var req struct {
		Title string `json:"title"`
	}
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid request body")
	}
	if req.Title == "" {
		return badRequest(c, "title is required")
	}

	claims, _ := c.Get("claims").(*infra.Claims)
	role := ""
	if claims != nil {
		role = claims.Role
	}

	created, err := h.svc.CreateSection(c.Request().Context(), courseID, req.Title, role)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusCreated, created)
}

func (h *Handler) AdminUpdateSection(c echo.Context) error {
	courseID := c.Param("id")
	sectionID := c.Param("sId")
	var req struct {
		Title string `json:"title"`
	}
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid request body")
	}

	claims, _ := c.Get("claims").(*infra.Claims)
	role := ""
	if claims != nil {
		role = claims.Role
	}

	updated, err := h.svc.UpdateSection(c.Request().Context(), courseID, sectionID, req.Title, role)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, updated)
}

func (h *Handler) AdminDeleteSection(c echo.Context) error {
	courseID := c.Param("id")
	sectionID := c.Param("sId")
	claims, _ := c.Get("claims").(*infra.Claims)
	role := ""
	if claims != nil {
		role = claims.Role
	}

	err := h.svc.DeleteSection(c.Request().Context(), courseID, sectionID, role)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) AdminReorderSections(c echo.Context) error {
	courseID := c.Param("id")
	var req struct {
		SectionIDs []string `json:"section_ids"`
	}
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid request body")
	}

	claims, _ := c.Get("claims").(*infra.Claims)
	role := ""
	if claims != nil {
		role = claims.Role
	}

	err := h.svc.ReorderSections(c.Request().Context(), courseID, req.SectionIDs, role)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "sections reordered",
	})
}

// --- Admin Lesson CRUD (re-keyed to course_id) ---

func (h *Handler) AdminCreateLesson(c echo.Context) error {
	courseID := c.Param("id")
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

	claims, _ := c.Get("claims").(*infra.Claims)
	role := ""
	if claims != nil {
		role = claims.Role
	}

	created, err := h.svc.CreateLesson(c.Request().Context(), courseID, sectionID, req.Title, req.VideoURL, req.Duration, role)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusCreated, created)
}

func (h *Handler) AdminUpdateLesson(c echo.Context) error {
	courseID := c.Param("id")
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

	claims, _ := c.Get("claims").(*infra.Claims)
	role := ""
	if claims != nil {
		role = claims.Role
	}

	updated, err := h.svc.UpdateLesson(c.Request().Context(), courseID, sectionID, lessonID, req.Title, req.VideoURL, req.Duration, role)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, updated)
}

func (h *Handler) AdminDeleteLesson(c echo.Context) error {
	courseID := c.Param("id")
	sectionID := c.Param("sId")
	lessonID := c.Param("lId")
	claims, _ := c.Get("claims").(*infra.Claims)
	role := ""
	if claims != nil {
		role = claims.Role
	}

	err := h.svc.DeleteLesson(c.Request().Context(), courseID, sectionID, lessonID, role)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) AdminReorderLessons(c echo.Context) error {
	courseID := c.Param("id")
	sectionID := c.Param("sId")
	var req struct {
		LessonIDs []string `json:"lesson_ids"`
	}
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid request body")
	}

	claims, _ := c.Get("claims").(*infra.Claims)
	role := ""
	if claims != nil {
		role = claims.Role
	}

	err := h.svc.ReorderLessons(c.Request().Context(), courseID, sectionID, req.LessonIDs, role)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "lessons reordered",
	})
}

// --- Student-facing handlers ---

func (h *Handler) StudentListLibrary(c echo.Context) error {
	claims, _ := c.Get("claims").(*infra.Claims)
	studentID := ""
	if claims != nil {
		studentID = claims.Sub
	}

	sessions, err := h.svc.ListLibrary(c.Request().Context(), studentID)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": sessions,
	})
}

func (h *Handler) StudentMarkLessonComplete(c echo.Context) error {
	claims, _ := c.Get("claims").(*infra.Claims)
	var studentID string
	if claims != nil {
		studentID = claims.Sub
	}

	courseID := c.Param("id")
	lessonID := c.Param("lId")

	err := h.svc.MarkLessonComplete(c.Request().Context(), studentID, courseID, lessonID, "")
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "lesson marked complete",
	})
}

func (h *Handler) StudentCourseProgress(c echo.Context) error {
	claims, _ := c.Get("claims").(*infra.Claims)
	var studentID string
	if claims != nil {
		studentID = claims.Sub
	}

	courseID := c.Param("id")

	count, percent, err := h.svc.CourseProgress(c.Request().Context(), studentID, courseID)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"count":   count,
		"percent": percent,
	})
}
