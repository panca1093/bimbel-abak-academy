package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"akademi-bimbel/internal/model"
	"akademi-bimbel/internal/repository"
)

var (
	ErrTestNotFound     = errors.New("test not found")
	ErrQuestionNotFound = errors.New("question not found")
	ErrValidation       = errors.New("validation failed")
)

var validQuestionFormats = map[string]bool{
	"mcq":          true,
	"multi_answer": true,
	"short":        true,
	"fill_blank":   true,
	"essay":        true,
}

// validateQuestion enforces the format-validation matrix from spec.md §4.
// All error returns wrap ErrValidation with a sub-message so callers can
// use errors.Is(err, ErrValidation) AND err.Error() carries the WHY.
func validateQuestion(q model.Question, options []model.QuestionOption) error {
	if !validQuestionFormats[q.Format] {
		return fmt.Errorf("%w: unknown question format: %s", ErrValidation, q.Format)
	}

	seenKeys := map[string]bool{}
	for _, o := range options {
		if o.Key == "" {
			return fmt.Errorf("%w: option key cannot be empty", ErrValidation)
		}
		if strings.TrimSpace(o.Text) == "" {
			return fmt.Errorf("%w: option text cannot be empty", ErrValidation)
		}
		if seenKeys[o.Key] {
			return fmt.Errorf("%w: duplicate option key: %s", ErrValidation, o.Key)
		}
		seenKeys[o.Key] = true
	}

	hasOptions := len(options) > 0
	hasCorrectAnswer := q.CorrectAnswer != nil && strings.TrimSpace(*q.CorrectAnswer) != ""
	correctCount := 0
	for _, o := range options {
		if o.IsCorrect {
			correctCount++
		}
	}

	switch q.Format {
	case "mcq":
		if len(options) < 2 {
			return fmt.Errorf("%w: mcq requires at least 2 options", ErrValidation)
		}
		if correctCount != 1 {
			return fmt.Errorf("%w: mcq requires exactly 1 correct option", ErrValidation)
		}
		if hasCorrectAnswer {
			return fmt.Errorf("%w: mcq cannot have correct_answer", ErrValidation)
		}
	case "multi_answer":
		if len(options) < 2 {
			return fmt.Errorf("%w: multi_answer requires at least 2 options", ErrValidation)
		}
		if correctCount < 1 {
			return fmt.Errorf("%w: multi_answer requires at least 1 correct option", ErrValidation)
		}
		if hasCorrectAnswer {
			return fmt.Errorf("%w: multi_answer cannot have correct_answer", ErrValidation)
		}
	case "short":
		if hasOptions {
			return fmt.Errorf("%w: short cannot have options", ErrValidation)
		}
		if !hasCorrectAnswer {
			return fmt.Errorf("%w: short requires non-empty correct_answer", ErrValidation)
		}
	case "fill_blank":
		if hasOptions {
			return fmt.Errorf("%w: fill_blank cannot have options", ErrValidation)
		}
		if !hasCorrectAnswer {
			return fmt.Errorf("%w: fill_blank requires non-empty correct_answer", ErrValidation)
		}
	case "essay":
		if hasOptions {
			return fmt.Errorf("%w: essay cannot have options", ErrValidation)
		}
		if hasCorrectAnswer {
			return fmt.Errorf("%w: essay cannot have correct_answer", ErrValidation)
		}
	}

	return nil
}

// validateTest enforces metadata-only invariants for Test CRUD. Format-specific
// validation lives in validateQuestion; this is just title/subject/topic/duration/audio.
func validateTest(t model.Test) error {
	if strings.TrimSpace(t.Title) == "" {
		return fmt.Errorf("%w: test title required", ErrValidation)
	}
	if strings.TrimSpace(t.Subject) == "" || strings.TrimSpace(t.Topic) == "" {
		return fmt.Errorf("%w: test subject/topic required", ErrValidation)
	}
	if t.DurationMinutes <= 0 {
		return fmt.Errorf("%w: duration_minutes must be positive", ErrValidation)
	}
	if t.AudioPlayLimit != nil && *t.AudioPlayLimit <= 0 {
		return fmt.Errorf("%w: audio_play_limit must be positive when set", ErrValidation)
	}
	return nil
}

func (s *Service) CreateTest(ctx context.Context, t model.Test) (model.Test, error) {
	if err := validateTest(t); err != nil {
		return model.Test{}, err
	}
	if err := s.storeRepo.CreateTest(ctx, &t); err != nil {
		return model.Test{}, err
	}
	return t, nil
}

func (s *Service) UpdateTest(ctx context.Context, id uuid.UUID, t model.Test) (model.Test, error) {
	if err := validateTest(t); err != nil {
		return model.Test{}, err
	}
	if _, err := s.storeRepo.GetTestByID(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.Test{}, ErrTestNotFound
		}
		return model.Test{}, err
	}
	if err := s.storeRepo.UpdateTest(ctx, id, &t); err != nil {
		if errors.Is(err, repository.ErrSortOrderConflict) {
			return model.Test{}, fmt.Errorf("%w: sort order conflict", ErrValidation)
		}
		return model.Test{}, err
	}
	t.ID = id
	return t, nil
}

func (s *Service) DeleteTest(ctx context.Context, id uuid.UUID) error {
	if _, err := s.storeRepo.GetTestByID(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrTestNotFound
		}
		return err
	}
	return s.storeRepo.DeleteTest(ctx, id)
}

func (s *Service) ListTests(ctx context.Context, filter repository.TestFilter) ([]model.Test, string, error) {
	return s.storeRepo.ListTests(ctx, filter)
}

func (s *Service) GetTestDetail(ctx context.Context, id uuid.UUID) (model.TestDetail, error) {
	d, err := s.storeRepo.GetTestDetail(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.TestDetail{}, ErrTestNotFound
		}
		return model.TestDetail{}, err
	}
	return *d, nil
}

// SaveQuestion routes create vs update by q.ID == uuid.Nil.
func (s *Service) SaveQuestion(ctx context.Context, q model.Question, options []model.QuestionOption) (model.QuestionWithOptions, error) {
	if err := validateQuestion(q, options); err != nil {
		return model.QuestionWithOptions{}, err
	}

	tx, err := s.storeRepo.BeginTx(ctx)
	if err != nil {
		return model.QuestionWithOptions{}, err
	}
	defer tx.Rollback(ctx)

	if q.ID == uuid.Nil {
		if err := s.storeRepo.CreateQuestionTx(ctx, tx, &q, options); err != nil {
			if errors.Is(err, repository.ErrSortOrderConflict) {
				return model.QuestionWithOptions{}, fmt.Errorf("%w: sort order conflict", ErrValidation)
			}
			return model.QuestionWithOptions{}, err
		}
	} else {
		if err := s.storeRepo.UpdateQuestionTx(ctx, tx, &q, options); err != nil {
			if errors.Is(err, repository.ErrSortOrderConflict) {
				return model.QuestionWithOptions{}, fmt.Errorf("%w: sort order conflict", ErrValidation)
			}
			if errors.Is(err, pgx.ErrNoRows) {
				return model.QuestionWithOptions{}, ErrQuestionNotFound
			}
			return model.QuestionWithOptions{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return model.QuestionWithOptions{}, err
	}

	return model.QuestionWithOptions{Question: q, Options: options}, nil
}

func (s *Service) DeleteQuestion(ctx context.Context, id uuid.UUID) error {
	if err := s.storeRepo.DeleteQuestion(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrQuestionNotFound
		}
		return err
	}
	return nil
}