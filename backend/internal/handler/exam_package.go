package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"akademi-bimbel/internal/model"
	"akademi-bimbel/internal/repository"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (h *Handler) AdminListExams(c echo.Context) error {
	limit := 20
	if l := c.QueryParam("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}
	filter := repository.ExamFilter{
		Cursor: c.QueryParam("cursor"),
		Limit:  limit,
	}

	exams, nextCursor, err := h.svc.ListExams(c.Request().Context(), filter)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":        exams,
		"next_cursor": nextCursor,
	})
}

func (h *Handler) AdminCreateExam(c echo.Context) error {
	var req model.Exam
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid request body")
	}
	if req.ID == uuid.Nil {
		req.ID = uuid.New()
	}

	exam, product, err := h.svc.CreateExam(c.Request().Context(), req)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(http.StatusCreated, map[string]interface{}{
		"exam":    exam,
		"product": product,
	})
}

func (h *Handler) AdminGetExam(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return badRequest(c, "invalid id")
	}
	detail, err := h.svc.GetExam(c.Request().Context(), id)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(http.StatusOK, detail)
}

func (h *Handler) AdminUpdateExam(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return badRequest(c, "invalid id")
	}

	existing, err := h.svc.GetExam(c.Request().Context(), id)
	if err != nil {
		return mapServiceError(c, err)
	}

	var req model.Exam
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid request body")
	}

	overlay := existing.Exam
	if req.Title != "" {
		overlay.Title = req.Title
	}
	if req.ScheduledAt != nil {
		overlay.ScheduledAt = req.ScheduledAt
	}
	if req.TimerMode != "" {
		overlay.TimerMode = req.TimerMode
	}
	if req.DurationMinutes != nil {
		overlay.DurationMinutes = req.DurationMinutes
	}
	if req.ResultConfig != "" {
		overlay.ResultConfig = req.ResultConfig
	}
	if req.ResultReleaseAt != nil {
		overlay.ResultReleaseAt = req.ResultReleaseAt
	}
	if req.Status != "" {
		overlay.Status = req.Status
	}
	if req.CheckInWindowMinutes != nil {
		overlay.CheckInWindowMinutes = req.CheckInWindowMinutes
	}
	if req.GraceWindowMinutes != nil {
		overlay.GraceWindowMinutes = req.GraceWindowMinutes
	}
	if req.MaxAttempts != nil {
		overlay.MaxAttempts = req.MaxAttempts
	}
	if req.BundleURL != nil {
		overlay.BundleURL = req.BundleURL
	}
	if req.BundleGeneratedAt != nil {
		overlay.BundleGeneratedAt = req.BundleGeneratedAt
	}
	overlay.IsFree = req.IsFree
	overlay.RequiresCheckin = req.RequiresCheckin
	overlay.AllowLeaderboard = req.AllowLeaderboard
	overlay.CDNBundle = req.CDNBundle
	overlay.Randomize = req.Randomize

	out, err := h.svc.UpdateExam(c.Request().Context(), id, overlay)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(http.StatusOK, out)
}

type examTestReplaceItem struct {
	TestID    uuid.UUID `json:"test_id"`
	SortOrder *int      `json:"sort_order,omitempty"`
}

// AdminReplaceExamTests accepts either `[{test_id,sort_order},...]` or `[test_id,...]`
// and replaces the exam's attached-test list atomically. Service assigns sort_order
// from list position; sort_order in the body is accepted for forward-compat but ignored.
func (h *Handler) AdminReplaceExamTests(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return badRequest(c, "invalid id")
	}

	raw, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return badRequest(c, "invalid request body")
	}

	var testIDs []uuid.UUID
	if err := json.Unmarshal(raw, &testIDs); err != nil {
		var rich []examTestReplaceItem
		if err2 := json.Unmarshal(raw, &rich); err2 != nil {
			return badRequest(c, "invalid request body")
		}
		for _, it := range rich {
			if it.TestID == uuid.Nil {
				return badRequest(c, "test_id required")
			}
			testIDs = append(testIDs, it.TestID)
		}
	}

	if err := h.svc.ReplaceExamTests(c.Request().Context(), id, testIDs); err != nil {
		return mapServiceError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) AdminUpdateExamPrice(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return badRequest(c, "invalid id")
	}
	var req struct {
		Price int64 `json:"price"`
	}
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid request body")
	}
	if err := h.svc.UpdateExamPrice(c.Request().Context(), id, req.Price); err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"price": req.Price,
	})
}

func (h *Handler) AdminPublishExam(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return badRequest(c, "invalid id")
	}
	if err := h.svc.PublishExam(c.Request().Context(), id); err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]string{
		"message": "exam published",
	})
}