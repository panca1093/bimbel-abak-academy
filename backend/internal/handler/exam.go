package handler

import (
	"bytes"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"akademi-bimbel/internal/model"
	"akademi-bimbel/internal/repository"
)

func (h *Handler) AdminListTests(c echo.Context) error {
	limit := 20
	if l := c.QueryParam("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}
	filter := repository.TestFilter{
		Subject: c.QueryParam("subject"),
		Topic:   c.QueryParam("topic"),
		Cursor:  c.QueryParam("cursor"),
		Limit:   limit,
	}

	tests, nextCursor, err := h.svc.ListTests(c.Request().Context(), filter)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":        tests,
		"next_cursor": nextCursor,
	})
}

func (h *Handler) AdminCreateTest(c echo.Context) error {
	var req struct {
		Title           string  `json:"title"`
		Subject         string  `json:"subject"`
		Topic           string  `json:"topic"`
		DurationMinutes int     `json:"duration_minutes"`
		AudioURL        *string `json:"audio_url,omitempty"`
		AudioPlayLimit  *int    `json:"audio_play_limit,omitempty"`
	}
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid request body")
	}

	t := model.Test{
		Title:           req.Title,
		Subject:         req.Subject,
		Topic:           req.Topic,
		DurationMinutes: req.DurationMinutes,
		AudioURL:        req.AudioURL,
		AudioPlayLimit:  req.AudioPlayLimit,
	}

	out, err := h.svc.CreateTest(c.Request().Context(), t)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(http.StatusCreated, out)
}

func (h *Handler) AdminGetTest(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return badRequest(c, "invalid id")
	}
	detail, err := h.svc.GetTestDetail(c.Request().Context(), id)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(http.StatusOK, detail)
}

func (h *Handler) AdminUpdateTest(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return badRequest(c, "invalid id")
	}
	// PATCH is partial — read existing and overlay only fields supplied.
	// (Service validateTest enforces all required fields; merging here keeps
	// the service contract intact while supporting PATCH semantics.)
	existing, err := h.svc.GetTestDetail(c.Request().Context(), id)
	if err != nil {
		return mapServiceError(c, err)
	}
	var req struct {
		Title           string  `json:"title"`
		Subject         string  `json:"subject"`
		Topic           string  `json:"topic"`
		DurationMinutes int     `json:"duration_minutes"`
		AudioURL        *string `json:"audio_url,omitempty"`
		AudioPlayLimit  *int    `json:"audio_play_limit,omitempty"`
	}
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid request body")
	}

	t := existing.Test
	if req.Title != "" {
		t.Title = req.Title
	}
	if req.Subject != "" {
		t.Subject = req.Subject
	}
	if req.Topic != "" {
		t.Topic = req.Topic
	}
	if req.DurationMinutes > 0 {
		t.DurationMinutes = req.DurationMinutes
	}
	if req.AudioURL != nil {
		t.AudioURL = req.AudioURL
	}
	if req.AudioPlayLimit != nil {
		t.AudioPlayLimit = req.AudioPlayLimit
	}

	out, err := h.svc.UpdateTest(c.Request().Context(), id, t)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(http.StatusOK, out)
}

func (h *Handler) AdminDeleteTest(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return badRequest(c, "invalid id")
	}
	if err := h.svc.DeleteTest(c.Request().Context(), id); err != nil {
		return mapServiceError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) AdminListQuestions(c echo.Context) error {
	testID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return badRequest(c, "invalid id")
	}
	detail, err := h.svc.GetTestDetail(c.Request().Context(), testID)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":        detail.Questions,
		"next_cursor": "",
	})
}

func (h *Handler) AdminCreateQuestion(c echo.Context) error {
	testID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return badRequest(c, "invalid id")
	}

	var req questionRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid request body")
	}

	q := req.toQuestion()
	q.TestID = testID
	out, err := h.svc.SaveQuestion(c.Request().Context(), q, req.toOptions())
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(http.StatusCreated, out)
}

func (h *Handler) AdminUpdateQuestion(c echo.Context) error {
	qID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return badRequest(c, "invalid id")
	}

	var req questionRequest
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid request body")
	}

	q := req.toQuestion()
	q.ID = qID
	out, err := h.svc.SaveQuestion(c.Request().Context(), q, req.toOptions())
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(http.StatusOK, out)
}

func (h *Handler) AdminDeleteQuestion(c echo.Context) error {
	qID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return badRequest(c, "invalid id")
	}
	if err := h.svc.DeleteQuestion(c.Request().Context(), qID); err != nil {
		return mapServiceError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

// questionRequest is the shared body for AdminCreateQuestion / AdminUpdateQuestion.
type questionRequest struct {
	Format        string          `json:"format"`
	Body          string          `json:"body"`
	SortOrder     int             `json:"sort_order"`
	Difficulty    *string         `json:"difficulty,omitempty"`
	Explanation   *string         `json:"explanation,omitempty"`
	ImageURL      *string         `json:"image_url,omitempty"`
	CorrectAnswer *string         `json:"correct_answer,omitempty"`
	Options       []optionRequest `json:"options,omitempty"`
}

type optionRequest struct {
	Key       string  `json:"key"`
	Text      string  `json:"text"`
	ImageURL  *string `json:"image_url,omitempty"`
	IsCorrect bool    `json:"is_correct"`
	SortOrder int     `json:"sort_order"`
}

func (r questionRequest) toQuestion() model.Question {
	return model.Question{
		Format:        r.Format,
		Body:          r.Body,
		CorrectAnswer: r.CorrectAnswer,
		Explanation:   r.Explanation,
		Difficulty:    r.Difficulty,
		ImageURL:      r.ImageURL,
		SortOrder:     r.SortOrder,
	}
}

func (r questionRequest) toOptions() []model.QuestionOption {
	out := make([]model.QuestionOption, 0, len(r.Options))
	for _, o := range r.Options {
		out = append(out, model.QuestionOption{
			Key:       o.Key,
			Text:      o.Text,
			ImageURL:  o.ImageURL,
			IsCorrect: o.IsCorrect,
			SortOrder: o.SortOrder,
		})
	}
	return out
}

// Student registration handlers (moved from competition.go).
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