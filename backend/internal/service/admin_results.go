package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"akademi-bimbel/internal/model"
	"akademi-bimbel/internal/repository"
)

// ListSchoolResults returns cursor-paginated fully-graded submitted sessions for an
// exam, scoped to a single school (FR-SCHOOL-08-01..11). Exam-level gates (hidden,
// locked) return an empty result, not an error (FR-SCHOOL-08-05). Cursor errors are
// surfaced as 422 via mapCursorErr.
func (s *Service) ListSchoolResults(ctx context.Context, examID uuid.UUID, schoolID, q, cursor string, limit int) ([]model.AdminResultRow, string, error) {
	exam, err := s.storeRepo.GetExamByID(ctx, examID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, "", ErrExamNotFound
		}
		return nil, "", err
	}

	// Exam-level gates only (FR-SCHOOL-08-05): forcing both bools true collapses
	// the 3-gate check to gates 1 (hidden) and 3 (locked), skipping gate 2 (grading).
	if _, ok := resultGate(*exam, true, true); ok {
		return nil, "", nil
	}

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	filter := repository.AdminResultFilter{
		Q:      q,
		Cursor: cursor,
		Limit:  limit,
	}

	rows, next, err := s.storeRepo.ListSchoolResults(ctx, examID, schoolID, filter)
	return rows, next, mapCursorErr(err)
}

// GetSchoolResultDetail returns the full detail of a single school-scoped session
// result (FR-SCHOOL-08-12..16). Gate violations (hidden, grading, locked) and
// cross-school access all surface as ErrSessionNotFound (FR-SCHOOL-08-13). Never
// calls CountHigherScores or computeRank — no rank field in the response
// (FR-SCHOOL-08-16).
func (s *Service) GetSchoolResultDetail(ctx context.Context, sessionID uuid.UUID, schoolID string) (model.AdminResultDetail, error) {
	sess, err := s.storeRepo.GetSchoolResultSession(ctx, sessionID, schoolID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.AdminResultDetail{}, ErrSessionNotFound
		}
		return model.AdminResultDetail{}, err
	}

	exam, err := s.storeRepo.GetExamForSession(ctx, sess.ExamID)
	if err != nil {
		return model.AdminResultDetail{}, err
	}

	tests, err := s.storeRepo.GetSessionWithQuestions(ctx, sess.ExamID)
	if err != nil {
		return model.AdminResultDetail{}, err
	}
	answers, err := s.storeRepo.GetSessionAnswers(ctx, sessionID)
	if err != nil {
		return model.AdminResultDetail{}, err
	}

	var qs []model.QuestionWithOptions
	for _, td := range tests {
		qs = append(qs, td.Questions...)
	}

	// Gate check (FR-SCHOOL-08-13): all three gates map to ErrSessionNotFound.
	if _, ok := resultGate(*exam, sess.Status == "submitted", isFullyGraded(qs, answers)); ok {
		return model.AdminResultDetail{}, ErrSessionNotFound
	}

	score := 0.0
	if sess.Score != nil {
		score = *sess.Score
	}
	correct, wrong, empty := objectiveCounts(qs, answers)

	detail := model.AdminResultDetail{
		SessionID:    sess.SessionID,
		StudentName:  sess.StudentName,
		NIS:          sess.NIS,
		Score:        score,
		SubmittedAt:  sess.SubmittedAt,
		ResultConfig: exam.ResultConfig,
		CorrectCount: correct,
		WrongCount:   wrong,
		EmptyCount:   empty,
	}

	if exam.ResultConfig == "score_pembahasan" {
		detail.Breakdown = topicBreakdown(tests, answers)
		detail.Pembahasan = buildPembahasan(qs, answers)
	}

	return detail, nil
}
