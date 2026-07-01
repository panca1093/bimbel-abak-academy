package handler

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"akademi-bimbel/internal/model"
	"akademi-bimbel/internal/repository"
	"akademi-bimbel/internal/service"
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
	PointCorrect  *int            `json:"point_correct,omitempty"`
	PointWrong    *int            `json:"point_wrong,omitempty"`
}

type optionRequest struct {
	Key       string  `json:"key"`
	Text      string  `json:"text"`
	ImageURL  *string `json:"image_url,omitempty"`
	IsCorrect bool    `json:"is_correct"`
	SortOrder int     `json:"sort_order"`
}

func (r questionRequest) toQuestion() model.Question {
	pointCorrect := 1
	if r.PointCorrect != nil {
		pointCorrect = *r.PointCorrect
	}
	pointWrong := 0
	if r.PointWrong != nil {
		pointWrong = *r.PointWrong
	}
	return model.Question{
		Format:        r.Format,
		Body:          r.Body,
		CorrectAnswer: r.CorrectAnswer,
		Explanation:   r.Explanation,
		Difficulty:    r.Difficulty,
		ImageURL:      r.ImageURL,
		SortOrder:     r.SortOrder,
		PointCorrect:  pointCorrect,
		PointWrong:    pointWrong,
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

// fingerprint derives a device fingerprint from IP and User-Agent.
func fingerprint(ip, ua string) string {
	h := sha256.Sum256([]byte(ip + "|" + ua))
	return hex.EncodeToString(h[:])
}

// StudentCheckIn validates the registration token and stamps check-in. FR2-FR5.
func (h *Handler) StudentCheckIn(c echo.Context) error {
	claims := claimsFromContext(c)
	if claims == nil || claims.Sub == "" {
		return c.JSON(http.StatusUnauthorized, APIError{Code: "unauthorized", Message: "missing auth"})
	}
	var req struct {
		Token string `json:"token"`
	}
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid request body")
	}
	fp := fingerprint(c.RealIP(), c.Request().UserAgent())
	result, err := h.svc.CheckIn(c.Request().Context(), claims.Sub, req.Token, fp)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

// StudentStartSession creates a new exam session. FR6-FR12.
func (h *Handler) StudentStartSession(c echo.Context) error {
	claims := claimsFromContext(c)
	if claims == nil || claims.Sub == "" {
		return c.JSON(http.StatusUnauthorized, APIError{Code: "unauthorized", Message: "missing auth"})
	}
	var req struct {
		RegistrationID string `json:"registration_id"`
	}
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid request body")
	}
	fp := fingerprint(c.RealIP(), c.Request().UserAgent())
	result, err := h.svc.StartSession(c.Request().Context(), claims.Sub, req.RegistrationID, fp)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

// StudentReconnectSession returns current session state. FR13-FR14.
func (h *Handler) StudentReconnectSession(c echo.Context) error {
	claims := claimsFromContext(c)
	if claims == nil || claims.Sub == "" {
		return c.JSON(http.StatusUnauthorized, APIError{Code: "unauthorized", Message: "missing auth"})
	}
	sessionID := c.Param("id")
	result, err := h.svc.ReconnectSession(c.Request().Context(), claims.Sub, sessionID)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

// StudentSaveAnswers upserts answers for a session. FR15-FR16.
func (h *Handler) StudentSaveAnswers(c echo.Context) error {
	claims := claimsFromContext(c)
	if claims == nil || claims.Sub == "" {
		return c.JSON(http.StatusUnauthorized, APIError{Code: "unauthorized", Message: "missing auth"})
	}
	sessionID := c.Param("id")
	var req struct {
		Answers []service.AnswerInput `json:"answers"`
	}
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid request body")
	}
	if err := h.svc.SaveAnswers(c.Request().Context(), claims.Sub, sessionID, req.Answers); err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

// StudentSubmitSession grades and submits the session. FR17-FR20.
func (h *Handler) StudentSubmitSession(c echo.Context) error {
	claims := claimsFromContext(c)
	if claims == nil || claims.Sub == "" {
		return c.JSON(http.StatusUnauthorized, APIError{Code: "unauthorized", Message: "missing auth"})
	}
	sessionID := c.Param("id")
	result, err := h.svc.SubmitSession(c.Request().Context(), claims.Sub, sessionID)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(http.StatusOK, result)
}

// StudentLogViolation records an integrity event. FR21-FR22.
func (h *Handler) StudentLogViolation(c echo.Context) error {
	claims := claimsFromContext(c)
	if claims == nil || claims.Sub == "" {
		return c.JSON(http.StatusUnauthorized, APIError{Code: "unauthorized", Message: "missing auth"})
	}
	sessionID := c.Param("id")
	var req struct {
		ViolationType string `json:"violation_type"`
	}
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid request body")
	}
	if err := h.svc.LogViolation(c.Request().Context(), claims.Sub, sessionID, req.ViolationType); err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

// AdminReopenSession extends a session's deadline. FR23.
func (h *Handler) AdminReopenSession(c echo.Context) error {
	sessionID := c.Param("id")
	var req struct {
		ExtendMinutes int `json:"extend_minutes"`
	}
	if err := c.Bind(&req); err != nil {
		return badRequest(c, "invalid request body")
	}
	if err := h.svc.ReopenSession(c.Request().Context(), sessionID, req.ExtendMinutes); err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

// AdminForceSubmitSession grades and submits an in-progress session. FR24.
func (h *Handler) AdminForceSubmitSession(c echo.Context) error {
	sessionID := c.Param("id")
	result, err := h.svc.ForceSubmitSession(c.Request().Context(), sessionID)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(http.StatusOK, result)
}