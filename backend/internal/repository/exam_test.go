package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"akademi-bimbel/internal/model"
)

// Compile-time check: *Repository must implement all exam methods declared by Task 4.
var _ interface {
	CreateTest(context.Context, *model.Test) error
	GetTestByID(context.Context, uuid.UUID) (*model.Test, error)
	GetTestDetail(context.Context, uuid.UUID) (*model.TestDetail, error)
	ListTests(context.Context, TestFilter) ([]model.Test, string, error)
	UpdateTest(context.Context, uuid.UUID, *model.Test) error
	DeleteTest(context.Context, uuid.UUID) error
	ListQuestions(context.Context, uuid.UUID) ([]model.QuestionWithOptions, error)
	CreateQuestionTx(context.Context, pgx.Tx, *model.Question, []model.QuestionOption) error
	UpdateQuestionTx(context.Context, pgx.Tx, *model.Question, []model.QuestionOption) error
	DeleteQuestion(context.Context, uuid.UUID) error
} = (*Repository)(nil)

// Sentinel error from the package: uq_question_order SQLSTATE 23505 — surfaced for service-layer mapping.
var _ error = ErrSortOrderConflict

func TestErrSortOrderConflict_isExported(t *testing.T) {
	if ErrSortOrderConflict == nil {
		t.Fatal("ErrSortOrderConflict must be a non-nil sentinel")
	}
	if ErrSortOrderConflict.Error() == "" {
		t.Error("ErrSortOrderConflict must have a non-empty message")
	}
}

func TestErrSortOrderConflict_isDistinctFromErrNotFound(t *testing.T) {
	if errors.Is(ErrSortOrderConflict, ErrNotFound) {
		t.Error("ErrSortOrderConflict must NOT be wrapped by/equal to ErrNotFound")
	}
}

func TestTestFilterShape(t *testing.T) {
	f := TestFilter{
		Subject: "math",
		Topic:   "algebra",
		Cursor:  uuid.NewString(),
		Limit:   10,
	}
	if f.Subject != "math" || f.Topic != "algebra" || f.Limit != 10 {
		t.Errorf("TestFilter fields not round-tripping: %+v", f)
	}
}

func TestTestDetailShape(t *testing.T) {
	now := time.Now()
	td := model.TestDetail{
		Test: model.Test{
			ID:              uuid.New(),
			Title:           "Sample",
			Subject:         "math",
			Topic:           "algebra",
			DurationMinutes: 60,
			CreatedAt:       now,
		},
		Questions: []model.QuestionWithOptions{
			{
				Question: model.Question{
					ID:        uuid.New(),
					TestID:    uuid.New(),
					Format:    "mcq",
					Body:      "2+2",
					SortOrder: 1,
				},
				Options: []model.QuestionOption{
					{QuestionID: uuid.New(), Key: "a", Text: "4", IsCorrect: true, SortOrder: 1},
				},
			},
		},
	}
	if len(td.Questions) != 1 || len(td.Questions[0].Options) != 1 {
		t.Errorf("TestDetail not assembling as expected: %+v", td)
	}
}

type recordingScanner struct {
	dests []any
	err   error
}

func (r *recordingScanner) Scan(dest ...any) error {
	r.dests = append(r.dests, dest...)
	return r.err
}

func TestScanTest_passes_expected_destinations(t *testing.T) {
	var t0 model.Test
	t0.ID = uuid.Nil

	rec := &recordingScanner{}
	if err := scanTest(rec, &t0); err != nil {
		t.Fatalf("scanTest returned error: %v", err)
	}

	if got := len(rec.dests); got != 8 {
		t.Fatalf("scanTest passed %d destinations, want 8 (id, title, subject, topic, duration_minutes, audio_url, audio_play_limit, created_at)", got)
	}

	if _, ok := rec.dests[0].(*uuid.UUID); !ok {
		t.Errorf("dest[0] = %T, want *uuid.UUID (id)", rec.dests[0])
	}
	if _, ok := rec.dests[1].(*string); !ok {
		t.Errorf("dest[1] = %T, want *string (title)", rec.dests[1])
	}
	if _, ok := rec.dests[2].(*string); !ok {
		t.Errorf("dest[2] = %T, want *string (subject)", rec.dests[2])
	}
	if _, ok := rec.dests[3].(*string); !ok {
		t.Errorf("dest[3] = %T, want *string (topic)", rec.dests[3])
	}
	if _, ok := rec.dests[4].(*int); !ok {
		t.Errorf("dest[4] = %T, want *int (duration_minutes)", rec.dests[4])
	}
	if _, ok := rec.dests[5].(**string); !ok {
		t.Errorf("dest[5] = %T, want **string (audio_url, nullable local scanned into struct)", rec.dests[5])
	}
	if _, ok := rec.dests[6].(**int); !ok {
		t.Errorf("dest[6] = %T, want **int (audio_play_limit, nullable local scanned into struct)", rec.dests[6])
	}
	if _, ok := rec.dests[7].(*time.Time); !ok {
		t.Errorf("dest[7] = %T, want *time.Time (created_at)", rec.dests[7])
	}
}

func TestScanQuestion_passes_expected_destinations(t *testing.T) {
	var q model.Question
	q.ID = uuid.Nil
	q.TestID = uuid.Nil

	rec := &recordingScanner{}
	if err := scanQuestion(rec, &q); err != nil {
		t.Fatalf("scanQuestion returned error: %v", err)
	}

	if got := len(rec.dests); got != 9 {
		t.Fatalf("scanQuestion passed %d destinations, want 9 (id, test_id, format, body, correct_answer, explanation, difficulty, image_url, sort_order)", got)
	}

	if _, ok := rec.dests[0].(*uuid.UUID); !ok {
		t.Errorf("dest[0] = %T, want *uuid.UUID (id)", rec.dests[0])
	}
	if _, ok := rec.dests[1].(*uuid.UUID); !ok {
		t.Errorf("dest[1] = %T, want *uuid.UUID (test_id)", rec.dests[1])
	}
	if _, ok := rec.dests[2].(*string); !ok {
		t.Errorf("dest[2] = %T, want *string (format)", rec.dests[2])
	}
	if _, ok := rec.dests[3].(*string); !ok {
		t.Errorf("dest[3] = %T, want *string (body)", rec.dests[3])
	}
	if _, ok := rec.dests[4].(**string); !ok {
		t.Errorf("dest[4] = %T, want **string (correct_answer, nullable local)", rec.dests[4])
	}
	if _, ok := rec.dests[5].(**string); !ok {
		t.Errorf("dest[5] = %T, want **string (explanation, nullable local)", rec.dests[5])
	}
	if _, ok := rec.dests[6].(**string); !ok {
		t.Errorf("dest[6] = %T, want **string (difficulty, nullable local)", rec.dests[6])
	}
	if _, ok := rec.dests[7].(**string); !ok {
		t.Errorf("dest[7] = %T, want **string (image_url, nullable local)", rec.dests[7])
	}
	if _, ok := rec.dests[8].(*int); !ok {
		t.Errorf("dest[8] = %T, want *int (sort_order)", rec.dests[8])
	}
}

func TestScanQuestionOption_passes_expected_destinations(t *testing.T) {
	var o model.QuestionOption
	o.QuestionID = uuid.Nil

	rec := &recordingScanner{}
	if err := scanQuestionOption(rec, &o); err != nil {
		t.Fatalf("scanQuestionOption returned error: %v", err)
	}

	if got := len(rec.dests); got != 6 {
		t.Fatalf("scanQuestionOption passed %d destinations, want 6 (question_id, key, text, image_url, is_correct, sort_order)", got)
	}

	if _, ok := rec.dests[0].(*uuid.UUID); !ok {
		t.Errorf("dest[0] = %T, want *uuid.UUID (question_id)", rec.dests[0])
	}
	if _, ok := rec.dests[1].(*string); !ok {
		t.Errorf("dest[1] = %T, want *string (key)", rec.dests[1])
	}
	if _, ok := rec.dests[2].(*string); !ok {
		t.Errorf("dest[2] = %T, want *string (text)", rec.dests[2])
	}
	if _, ok := rec.dests[3].(**string); !ok {
		t.Errorf("dest[3] = %T, want **string (image_url, nullable local)", rec.dests[3])
	}
	if _, ok := rec.dests[4].(*bool); !ok {
		t.Errorf("dest[4] = %T, want *bool (is_correct)", rec.dests[4])
	}
	if _, ok := rec.dests[5].(*int); !ok {
		t.Errorf("dest[5] = %T, want *int (sort_order)", rec.dests[5])
	}
}