package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"akademi-bimbel/internal/model"
	"akademi-bimbel/internal/repository"
)

// ---------- Session result types ----------

type CheckInResult struct {
	RegistrationID uuid.UUID `json:"registration_id"`
	ExamTitle      string    `json:"exam_title"`
	ScheduledAt    *time.Time `json:"scheduled_at"`
}

type SessionTestPayload struct {
	ID        uuid.UUID         `json:"id"`
	Title     string            `json:"title"`
	Subject   string            `json:"subject"`
	Questions []SessionQuestion `json:"questions"`
}

type SessionStartPayload struct {
	SessionID        uuid.UUID            `json:"session_id"`
	RemainingSeconds int64                `json:"remaining_seconds"`
	TimerMode        string               `json:"timer_mode"`
	DurationMinutes  *int                 `json:"duration_minutes"`
	Tests            []SessionTestPayload `json:"tests"`
}

type SessionQuestion struct {
	ID        uuid.UUID        `json:"id"`
	TestID    uuid.UUID        `json:"test_id"`
	Format    string           `json:"format"`
	Body      string           `json:"body"`
	Options   []SessionOption  `json:"options"`
	SortOrder int              `json:"sort_order"`
}

type SessionOption struct {
	Key       string  `json:"key"`
	Text      string  `json:"text"`
	ImageURL  *string `json:"image_url"`
	SortOrder int     `json:"sort_order"`
}

type SessionStatePayload struct {
	SessionID        uuid.UUID                `json:"session_id"`
	Status           string                   `json:"status"`
	RemainingSeconds int64                    `json:"remaining_seconds"`
	TimerMode        string                   `json:"timer_mode"`
	DurationMinutes  *int                     `json:"duration_minutes"`
	Tests            []SessionTestPayload     `json:"tests"`
	Answers          []model.ExamSessionAnswer `json:"answers"`
}

type SubmitResult struct {
	Status string   `json:"status"`
	Score  *float64 `json:"score"`
}

type AnswerInput struct {
	QuestionID       uuid.UUID `json:"question_id"`
	Answer           *string   `json:"answer"`
	FlaggedForReview bool      `json:"flagged_for_review"`
}

// fingerprint derives a device fingerprint from IP and User-Agent.
func fingerprint(ip, ua string) string {
	h := sha256.Sum256([]byte(ip + "|" + ua))
	return hex.EncodeToString(h[:])
}

// validViolationTypes is the set of allowed violation_type values.
var validViolationTypes = map[string]bool{
	"fullscreen_exit": true,
	"tab_switch":      true,
	"copy_attempt":    true,
}

// ---------- Shared helpers ----------

// groupQuestionsByTest groups questions by their parent test and strips
// correct_answer / is_correct from the student-facing payload.
func groupQuestionsByTest(tests []model.TestDetail) []SessionTestPayload {
	var out []SessionTestPayload
	for _, td := range tests {
		st := SessionTestPayload{
			ID:        td.Test.ID,
			Title:     td.Test.Title,
			Subject:   td.Test.Subject,
			Questions: make([]SessionQuestion, 0, len(td.Questions)),
		}
		for _, q := range td.Questions {
			sq := SessionQuestion{
				ID:        q.Question.ID,
				TestID:    q.Question.TestID,
				Format:    q.Question.Format,
				Body:      q.Question.Body,
				SortOrder: q.Question.SortOrder,
			}
			for _, o := range q.Options {
				sq.Options = append(sq.Options, SessionOption{
					Key:       o.Key,
					Text:      o.Text,
					ImageURL:  o.ImageURL,
					SortOrder: o.SortOrder,
				})
			}
			st.Questions = append(st.Questions, sq)
		}
		out = append(out, st)
	}
	return out
}

// computeRemainingSeconds calculates remaining time based on effective deadline.
func computeRemainingSeconds(startedAt time.Time, durationMinutes *int, extendedUntil *time.Time) int64 {
	if durationMinutes == nil || *durationMinutes <= 0 {
		return 0
	}
	deadline := startedAt.Add(time.Duration(*durationMinutes) * time.Minute)
	if extendedUntil != nil && extendedUntil.After(deadline) {
		deadline = *extendedUntil
	}
	return int64(math.Max(0, time.Until(deadline).Seconds()))
}

// ---------- CheckIn ----------

// CheckIn validates the registration token, the check-in window, and stamps
// checked_in_at in a transaction. A device-lock key is written to Redis.
// FR2–FR5.
func (s *Service) CheckIn(ctx context.Context, studentID, token, fp string) (CheckInResult, error) {
	sid, err := uuid.Parse(studentID)
	if err != nil {
		return CheckInResult{}, fmt.Errorf("%w: invalid student id", ErrValidation)
	}

	reg, err := s.storeRepo.GetExamRegistrationByToken(ctx, sid, token)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return CheckInResult{}, ErrRegistrationNotFound
		}
		return CheckInResult{}, err
	}

	exam, err := s.storeRepo.GetExamForSession(ctx, reg.ExamID)
	if err != nil {
		return CheckInResult{}, err
	}

	if !exam.RequiresCheckin {
		return CheckInResult{}, ErrNotCheckedIn
	}

	// Window check: now ∈ [scheduled_at − window, scheduled_at)
	if exam.ScheduledAt != nil && exam.CheckInWindowMinutes != nil {
		now := time.Now()
		windowStart := exam.ScheduledAt.Add(-time.Duration(*exam.CheckInWindowMinutes) * time.Minute)
		if now.Before(windowStart) || !now.Before(*exam.ScheduledAt) {
			return CheckInResult{}, ErrCheckinWindowClosed
		}
	}

	// Stamped checked_in_at in a transaction
	tx, err := s.storeRepo.BeginTx(ctx)
	if err != nil {
		return CheckInResult{}, err
	}
	defer tx.Rollback(ctx)

	if err := s.storeRepo.CheckInExamTx(ctx, tx, reg.ID); err != nil {
		return CheckInResult{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return CheckInResult{}, err
	}

	// Redis device lock
	key := "exam:device:" + reg.ID.String()
	ttl := 24 * time.Hour
	if exam.DurationMinutes != nil && *exam.DurationMinutes > 0 {
		ttl = time.Duration(*exam.DurationMinutes) * time.Minute
	}
	if err := s.rdb.Set(ctx, key, fp, ttl).Err(); err != nil {
		return CheckInResult{}, err
	}

	return CheckInResult{
		RegistrationID: reg.ID,
		ExamTitle:      exam.Title,
		ScheduledAt:    exam.ScheduledAt,
	}, nil
}

// ---------- StartSession ----------

// StartSession creates a new exam session. For requires_checkin=true exams it
// validates the window, check-in status, and device fingerprint. FR6–FR12.
func (s *Service) StartSession(ctx context.Context, studentID, registrationID, fp string) (SessionStartPayload, error) {
	sid, err := uuid.Parse(studentID)
	if err != nil {
		return SessionStartPayload{}, fmt.Errorf("%w: invalid student id", ErrValidation)
	}
	rid, err := uuid.Parse(registrationID)
	if err != nil {
		return SessionStartPayload{}, fmt.Errorf("%w: invalid registration id", ErrValidation)
	}

	detail, err := s.storeRepo.GetExamRegistrationByID(ctx, rid, sid)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return SessionStartPayload{}, ErrRegistrationNotFound
		}
		return SessionStartPayload{}, err
	}

	exam, err := s.storeRepo.GetExamForSession(ctx, detail.ExamID)
	if err != nil {
		return SessionStartPayload{}, err
	}

	// requires_checkin branch
	if exam.RequiresCheckin {
		if exam.ScheduledAt != nil && time.Now().Before(*exam.ScheduledAt) {
			return SessionStartPayload{}, ErrExamNotStarted
		}
		if detail.CheckedInAt == nil {
			return SessionStartPayload{}, ErrNotCheckedIn
		}

		key := "exam:device:" + rid.String()
		deviceFP, err := s.rdb.Get(ctx, key).Result()
		if err == redis.Nil {
			return SessionStartPayload{}, ErrDeviceMismatch
		}
		if err != nil {
			return SessionStartPayload{}, err
		}
		if deviceFP != fp {
			return SessionStartPayload{}, ErrDeviceMismatch
		}
	}

	if detail.AttemptsUsed >= 1 {
		return SessionStartPayload{}, ErrAlreadyAttempted
	}

	tx, err := s.storeRepo.BeginTx(ctx)
	if err != nil {
		return SessionStartPayload{}, err
	}
	defer tx.Rollback(ctx)

	sess, err := s.storeRepo.CreateExamSessionTx(ctx, tx, detail.ExamRegistration)
	if err != nil {
		return SessionStartPayload{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return SessionStartPayload{}, err
	}

	// Load questions
	tests, _ := s.storeRepo.GetSessionWithQuestions(ctx, detail.ExamID)
	grouped := groupQuestionsByTest(tests)

	remaining := computeRemainingSeconds(sess.StartedAt, exam.DurationMinutes, nil)

	return SessionStartPayload{
		SessionID:        sess.ID,
		RemainingSeconds: remaining,
		TimerMode:        exam.TimerMode,
		DurationMinutes:  exam.DurationMinutes,
		Tests:            grouped,
	}, nil
}

// ---------- ReconnectSession ----------

// ReconnectSession returns the current session state, questions, saved answers,
// and remaining time. FR13–FR14.
func (s *Service) ReconnectSession(ctx context.Context, studentID, sessionID string) (SessionStatePayload, error) {
	sid, err := uuid.Parse(studentID)
	if err != nil {
		return SessionStatePayload{}, fmt.Errorf("%w: invalid student id", ErrValidation)
	}
	sessID, err := uuid.Parse(sessionID)
	if err != nil {
		return SessionStatePayload{}, fmt.Errorf("%w: invalid session id", ErrValidation)
	}

	sess, err := s.storeRepo.GetExamSessionForStudent(ctx, sessID, sid)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return SessionStatePayload{}, ErrSessionNotFound
		}
		return SessionStatePayload{}, err
	}

	exam, err := s.storeRepo.GetExamForSession(ctx, sess.ExamID)
	if err != nil {
		return SessionStatePayload{}, err
	}

	remaining := computeRemainingSeconds(sess.StartedAt, exam.DurationMinutes, sess.ExtendedUntil)

	tests, _ := s.storeRepo.GetSessionWithQuestions(ctx, sess.ExamID)
	grouped := groupQuestionsByTest(tests)

	answers, _ := s.storeRepo.GetSessionAnswers(ctx, sessID)

	return SessionStatePayload{
		SessionID:        sess.ID,
		Status:           sess.Status,
		RemainingSeconds: remaining,
		TimerMode:        exam.TimerMode,
		DurationMinutes:  exam.DurationMinutes,
		Tests:            grouped,
		Answers:          answers,
	}, nil
}

// ---------- SaveAnswers ----------

// SaveAnswers upserts the student's answers for a session. FR15–FR16.
func (s *Service) SaveAnswers(ctx context.Context, studentID, sessionID string, inputs []AnswerInput) error {
	sid, err := uuid.Parse(studentID)
	if err != nil {
		return fmt.Errorf("%w: invalid student id", ErrValidation)
	}
	sessID, err := uuid.Parse(sessionID)
	if err != nil {
		return fmt.Errorf("%w: invalid session id", ErrValidation)
	}

	sess, err := s.storeRepo.GetExamSessionForStudent(ctx, sessID, sid)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrSessionNotFound
		}
		return err
	}

	if sess.Status != "in_progress" {
		return ErrAlreadySubmitted
	}

	answers := make([]model.ExamSessionAnswer, len(inputs))
	for i, in := range inputs {
		answers[i] = model.ExamSessionAnswer{
			SessionID:        sessID,
			QuestionID:       in.QuestionID,
			Answer:           in.Answer,
			FlaggedForReview: in.FlaggedForReview,
		}
	}

	return s.storeRepo.SaveAnswersTx(ctx, sessID, answers)
}

// ---------- SubmitSession ----------

// SubmitSession grades objective answers and marks the session as submitted.
// FR17–FR20.
func (s *Service) SubmitSession(ctx context.Context, studentID, sessionID string) (SubmitResult, error) {
	sid, err := uuid.Parse(studentID)
	if err != nil {
		return SubmitResult{}, fmt.Errorf("%w: invalid student id", ErrValidation)
	}
	sessID, err := uuid.Parse(sessionID)
	if err != nil {
		return SubmitResult{}, fmt.Errorf("%w: invalid session id", ErrValidation)
	}

	sess, err := s.storeRepo.GetExamSessionForStudent(ctx, sessID, sid)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return SubmitResult{}, ErrSessionNotFound
		}
		return SubmitResult{}, err
	}

	questions, _ := s.storeRepo.GetSessionWithQuestions(ctx, sess.ExamID)
	answers, _ := s.storeRepo.GetSessionAnswers(ctx, sessID)

	answerMap := make(map[uuid.UUID]*string)
	for _, a := range answers {
		answerMap[a.QuestionID] = a.Answer
	}

	var qs []model.QuestionWithOptions
	for _, td := range questions {
		qs = append(qs, td.Questions...)
	}

	graded, score := gradeObjective(qs, answerMap)

	tx, err := s.storeRepo.BeginTx(ctx)
	if err != nil {
		return SubmitResult{}, err
	}
	defer tx.Rollback(ctx)

	rows, err := s.storeRepo.SubmitSessionTx(ctx, tx, sessID, graded, score, false)
	if err != nil {
		return SubmitResult{}, err
	}
	if rows == 0 {
		return SubmitResult{}, ErrAlreadySubmitted
	}

	if err := tx.Commit(ctx); err != nil {
		return SubmitResult{}, err
	}

	return SubmitResult{
		Status: "submitted",
		Score:  &score,
	}, nil
}

// ---------- LogViolation ----------

// LogViolation records an integrity event. FR21–FR22.
func (s *Service) LogViolation(ctx context.Context, studentID, sessionID, violationType string) error {
	if !validViolationTypes[violationType] {
		return ErrInvalidViolationType
	}

	sid, err := uuid.Parse(studentID)
	if err != nil {
		return fmt.Errorf("%w: invalid student id", ErrValidation)
	}
	sessID, err := uuid.Parse(sessionID)
	if err != nil {
		return fmt.Errorf("%w: invalid session id", ErrValidation)
	}

	sess, err := s.storeRepo.GetExamSessionForStudent(ctx, sessID, sid)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrSessionNotFound
		}
		return err
	}

	if sess.Status != "in_progress" {
		return ErrAlreadySubmitted
	}

	return s.storeRepo.LogViolation(ctx, model.SessionViolationLog{
		SessionID:     sessID,
		StudentID:     sid,
		ViolationType: violationType,
		OccurredAt:    time.Now(),
	})
}

// ---------- ReopenSession (admin) ----------

// ReopenSession extends a session's deadline. FR23.
func (s *Service) ReopenSession(ctx context.Context, sessionID string, minutes int) error {
	sessID, err := uuid.Parse(sessionID)
	if err != nil {
		return fmt.Errorf("%w: invalid session id", ErrValidation)
	}

	if err := s.storeRepo.ReopenSession(ctx, sessID, minutes); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrSessionNotFound
		}
		return err
	}
	return nil
}

// ---------- ForceSubmitSession (admin) ----------

// ForceSubmitSession grades and submits an in-progress session as admin. FR24.
func (s *Service) ForceSubmitSession(ctx context.Context, sessionID string) (SubmitResult, error) {
	sessID, err := uuid.Parse(sessionID)
	if err != nil {
		return SubmitResult{}, fmt.Errorf("%w: invalid session id", ErrValidation)
	}

	sess, err := s.storeRepo.GetExamSessionByID(ctx, sessID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return SubmitResult{}, ErrSessionNotFound
		}
		return SubmitResult{}, err
	}

	questions, _ := s.storeRepo.GetSessionWithQuestions(ctx, sess.ExamID)
	answers, _ := s.storeRepo.GetSessionAnswers(ctx, sessID)

	answerMap := make(map[uuid.UUID]*string)
	for _, a := range answers {
		answerMap[a.QuestionID] = a.Answer
	}

	var qs []model.QuestionWithOptions
	for _, td := range questions {
		qs = append(qs, td.Questions...)
	}

	graded, score := gradeObjective(qs, answerMap)

	tx, err := s.storeRepo.BeginTx(ctx)
	if err != nil {
		return SubmitResult{}, err
	}
	defer tx.Rollback(ctx)

	rows, err := s.storeRepo.SubmitSessionTx(ctx, tx, sessID, graded, score, true)
	if err != nil {
		return SubmitResult{}, err
	}
	if rows == 0 {
		return SubmitResult{}, ErrAlreadySubmitted
	}

	if err := tx.Commit(ctx); err != nil {
		return SubmitResult{}, err
	}

	return SubmitResult{
		Status: "submitted",
		Score:  &score,
	}, nil
}
// ---------- Session monitor helpers ----------

// effectiveDeadline computes the overdue threshold for a session.
// When durationMinutes is nil (per_test), no duration-based deadline applies
// (FR-6a) — only extended_until can push the deadline forward. The zero
// time.Time return signals "no deadline" to the caller.
func effectiveDeadline(startedAt time.Time, durationMinutes *int, graceMinutes *int, extendedUntil *time.Time) time.Time {
	var deadline time.Time

	if durationMinutes != nil && *durationMinutes > 0 {
		deadline = startedAt.Add(time.Duration(*durationMinutes) * time.Minute)
		if graceMinutes != nil {
			deadline = deadline.Add(time.Duration(*graceMinutes) * time.Minute)
		}
	}

	if extendedUntil != nil && !extendedUntil.IsZero() {
		if deadline.IsZero() || extendedUntil.After(deadline) {
			deadline = *extendedUntil
		}
	}

	return deadline
}

// deriveStatus sets the monitor row's derived status per FR-3.
func deriveStatus(row model.SessionMonitorRow, now time.Time, durationMinutes *int, graceMinutes *int) string {
	if row.SessionID == nil {
		if row.CheckedInAt != nil {
			return "checked_in"
		}
		return "registered"
	}

	if row.SessionStatus != nil && *row.SessionStatus == "submitted" {
		return "submitted"
	}

	// Session is in_progress — check deadline.
	deadline := effectiveDeadline(*row.StartedAt, durationMinutes, graceMinutes, row.ExtendedUntil)
	if deadline.IsZero() {
		return "in_progress"
	}
	if now.After(deadline) {
		return "overdue"
	}
	return "in_progress"
}

// ---------- GetSessionMonitor ----------

// GetSessionMonitor returns the monitor payload for an exam: the exam summary,
// one row per registrant with derived status, and recent violations. FR-1–FR-7.
func (s *Service) GetSessionMonitor(ctx context.Context, examID string) (model.SessionMonitorResponse, error) {
	eid, err := uuid.Parse(examID)
	if err != nil {
		return model.SessionMonitorResponse{}, fmt.Errorf("%w: invalid exam id", ErrValidation)
	}

	exam, err := s.storeRepo.GetExamForSession(ctx, eid)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.SessionMonitorResponse{}, ErrExamNotFound
		}
		return model.SessionMonitorResponse{}, err
	}

	rows, err := s.storeRepo.GetSessionMonitorRows(ctx, eid)
	if err != nil {
		return model.SessionMonitorResponse{}, err
	}

	totalQ, err := s.storeRepo.GetExamQuestionTotal(ctx, eid)
	if err != nil {
		return model.SessionMonitorResponse{}, err
	}

	recentV, err := s.storeRepo.GetRecentViolations(ctx, eid, 20)
	if err != nil {
		return model.SessionMonitorResponse{}, err
	}

	now := time.Now()
	for i := range rows {
		rows[i].TotalQuestions = totalQ
		rows[i].Status = deriveStatus(rows[i], now, exam.DurationMinutes, exam.GraceWindowMinutes)
	}

	return model.SessionMonitorResponse{
		Exam: model.SessionMonitorExam{
			ID:                 exam.ID,
			Title:              exam.Title,
			ScheduledAt:        exam.ScheduledAt,
			DurationMinutes:    exam.DurationMinutes,
			GraceWindowMinutes: exam.GraceWindowMinutes,
			Status:             exam.Status,
		},
		Rows:             rows,
		ViolationsRecent: recentV,
	}, nil
}

// ---------- GetSessionViolations ----------

// GetSessionViolations returns the violation log for a session, newest-first. FR-8.
func (s *Service) GetSessionViolations(ctx context.Context, sessionID string) ([]model.SessionViolationLog, error) {
	sid, err := uuid.Parse(sessionID)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid session id", ErrValidation)
	}
	return s.storeRepo.ListSessionViolations(ctx, sid)
}
