package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// AdminGrantExamAccess handles POST /admin/exam-grants.
// Creates direct exam registrations for the given students, bypassing the
// order pipeline (FR-GRANT-01/02/07). No school_id in the request —
// super_admin's grant is not school-scoped.
func (h *Handler) AdminGrantExamAccess(c echo.Context) error {
	claims := ClaimsFromContext(c)
	if claims == nil {
		return c.JSON(http.StatusUnauthorized, APIError{Code: "unauthorized", Message: "missing or invalid token"})
	}

	var req struct {
		ExamID     string      `json:"exam_id"`
		StudentIDs []uuid.UUID `json:"student_ids"`
	}
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return badRequest(c, "invalid request body")
	}
	if req.ExamID == "" {
		return badRequest(c, "exam_id is required")
	}
	if len(req.StudentIDs) == 0 {
		return badRequest(c, "student_ids is required")
	}

	registrations, err := h.svc.GrantExamAccess(c.Request().Context(), claims.Sub, req.ExamID, req.StudentIDs)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusCreated, map[string]any{
		"data": registrations,
	})
}
