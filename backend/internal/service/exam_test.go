package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"akademi-bimbel/internal/model"
	"akademi-bimbel/internal/repository"
)

func strPtr(s string) *string { return &s }

func TestValidateQuestion_mcq_accepts_exactly_one_correct(t *testing.T) {
	q := model.Question{Format: "mcq", Body: "2+2", PointCorrect: 1}
	options := []model.QuestionOption{
		{Key: "a", Text: "4", IsCorrect: true, SortOrder: 1},
		{Key: "b", Text: "5", SortOrder: 2},
	}
	if err := validateQuestion(q, options); err != nil {
		t.Errorf("mcq with 1 correct + 2 options should pass, got %v", err)
	}
}

func TestValidateQuestion_mcq_rejects_zero_correct(t *testing.T) {
	q := model.Question{Format: "mcq", Body: "2+2"}
	options := []model.QuestionOption{
		{Key: "a", Text: "4", SortOrder: 1},
		{Key: "b", Text: "5", SortOrder: 2},
	}
	err := validateQuestion(q, options)
	if !errors.Is(err, ErrValidation) {
		t.Errorf("mcq with 0 correct should return ErrValidation, got %v", err)
	}
	if !strings.Contains(err.Error(), "exactly 1 correct option") {
		t.Errorf("mcq 0-correct msg should mention 'exactly 1 correct option', got %q", err.Error())
	}
}

func TestValidateQuestion_mcq_rejects_two_correct(t *testing.T) {
	q := model.Question{Format: "mcq", Body: "2+2"}
	options := []model.QuestionOption{
		{Key: "a", Text: "4", IsCorrect: true, SortOrder: 1},
		{Key: "b", Text: "5", IsCorrect: true, SortOrder: 2},
	}
	err := validateQuestion(q, options)
	if !errors.Is(err, ErrValidation) {
		t.Errorf("mcq with 2 correct should return ErrValidation, got %v", err)
	}
	if !strings.Contains(err.Error(), "exactly 1 correct option") {
		t.Errorf("mcq 2-correct msg should mention 'exactly 1 correct option', got %q", err.Error())
	}
}

func TestValidateQuestion_mcq_rejects_fewer_than_2_options(t *testing.T) {
	q := model.Question{Format: "mcq", Body: "2+2"}
	options := []model.QuestionOption{
		{Key: "a", Text: "4", IsCorrect: true, SortOrder: 1},
	}
	err := validateQuestion(q, options)
	if !errors.Is(err, ErrValidation) {
		t.Errorf("mcq with 1 option should return ErrValidation, got %v", err)
	}
	if !strings.Contains(err.Error(), "at least 2 options") {
		t.Errorf("mcq 1-option msg should mention 'at least 2 options', got %q", err.Error())
	}
}

func TestValidateQuestion_multi_answer_accepts_one_or_more_correct(t *testing.T) {
	q := model.Question{Format: "multi_answer", Body: "primes", PointCorrect: 1}
	// one correct
	opts1 := []model.QuestionOption{
		{Key: "a", Text: "2", IsCorrect: true, SortOrder: 1},
		{Key: "b", Text: "4", SortOrder: 2},
		{Key: "c", Text: "6", SortOrder: 3},
	}
	if err := validateQuestion(q, opts1); err != nil {
		t.Errorf("multi_answer with 1 correct + 3 options should pass, got %v", err)
	}
	// two correct
	opts2 := []model.QuestionOption{
		{Key: "a", Text: "2", IsCorrect: true, SortOrder: 1},
		{Key: "b", Text: "4", IsCorrect: true, SortOrder: 2},
		{Key: "c", Text: "6", SortOrder: 3},
	}
	if err := validateQuestion(q, opts2); err != nil {
		t.Errorf("multi_answer with 2 correct + 3 options should pass, got %v", err)
	}
}

func TestValidateQuestion_multi_answer_rejects_zero_correct(t *testing.T) {
	q := model.Question{Format: "multi_answer", Body: "primes"}
	options := []model.QuestionOption{
		{Key: "a", Text: "2", SortOrder: 1},
		{Key: "b", Text: "4", SortOrder: 2},
	}
	err := validateQuestion(q, options)
	if !errors.Is(err, ErrValidation) {
		t.Errorf("multi_answer with 0 correct should return ErrValidation, got %v", err)
	}
	if !strings.Contains(err.Error(), "at least 1 correct option") {
		t.Errorf("multi_answer 0-correct msg should mention 'at least 1 correct option', got %q", err.Error())
	}
}

func TestValidateQuestion_short_requires_correct_answer(t *testing.T) {
	q := model.Question{Format: "short", Body: "capital of France"}
	err := validateQuestion(q, nil)
	if !errors.Is(err, ErrValidation) {
		t.Errorf("short with empty correct_answer should return ErrValidation, got %v", err)
	}
	if !strings.Contains(err.Error(), "non-empty correct_answer") {
		t.Errorf("short empty-answer msg should mention 'non-empty correct_answer', got %q", err.Error())
	}
}

func TestValidateQuestion_short_rejects_options(t *testing.T) {
	q := model.Question{Format: "short", Body: "x", CorrectAnswer: strPtr("y")}
	options := []model.QuestionOption{
		{Key: "a", Text: "y", SortOrder: 1},
	}
	err := validateQuestion(q, options)
	if !errors.Is(err, ErrValidation) {
		t.Errorf("short with options should return ErrValidation, got %v", err)
	}
	if !strings.Contains(err.Error(), "cannot have options") {
		t.Errorf("short options msg should mention 'cannot have options', got %q", err.Error())
	}
}

func TestValidateQuestion_fill_blank_requires_correct_answer(t *testing.T) {
	q := model.Question{Format: "fill_blank", Body: "the ___ is blue"}
	err := validateQuestion(q, nil)
	if !errors.Is(err, ErrValidation) {
		t.Errorf("fill_blank with empty correct_answer should return ErrValidation, got %v", err)
	}
	if !strings.Contains(err.Error(), "non-empty correct_answer") {
		t.Errorf("fill_blank empty-answer msg should mention 'non-empty correct_answer', got %q", err.Error())
	}
}

func TestValidateQuestion_essay_accepts_no_options_no_correct_answer(t *testing.T) {
	q := model.Question{Format: "essay", Body: "explain gravity", PointCorrect: 1}
	if err := validateQuestion(q, nil); err != nil {
		t.Errorf("essay with no options + no correct_answer should pass, got %v", err)
	}
}

func TestValidateQuestion_essay_rejects_options(t *testing.T) {
	q := model.Question{Format: "essay", Body: "explain"}
	options := []model.QuestionOption{
		{Key: "a", Text: "x", SortOrder: 1},
	}
	err := validateQuestion(q, options)
	if !errors.Is(err, ErrValidation) {
		t.Errorf("essay with options should return ErrValidation, got %v", err)
	}
	if !strings.Contains(err.Error(), "cannot have options") {
		t.Errorf("essay options msg should mention 'cannot have options', got %q", err.Error())
	}
}

func TestValidateQuestion_essay_rejects_correct_answer(t *testing.T) {
	q := model.Question{Format: "essay", Body: "explain", CorrectAnswer: strPtr("something")}
	err := validateQuestion(q, nil)
	if !errors.Is(err, ErrValidation) {
		t.Errorf("essay with correct_answer should return ErrValidation, got %v", err)
	}
	if !strings.Contains(err.Error(), "cannot have correct_answer") {
		t.Errorf("essay correct_answer msg should mention 'cannot have correct_answer', got %q", err.Error())
	}
}

func TestValidateQuestion_rejects_unknown_format(t *testing.T) {
	q := model.Question{Format: "matching", Body: "x"}
	err := validateQuestion(q, nil)
	if !errors.Is(err, ErrValidation) {
		t.Errorf("unknown format should return ErrValidation, got %v", err)
	}
	if !strings.Contains(err.Error(), "unknown question format") {
		t.Errorf("unknown format msg should mention 'unknown question format', got %q", err.Error())
	}
}

func TestValidateQuestion_rejects_duplicate_option_keys(t *testing.T) {
	q := model.Question{Format: "mcq", Body: "x"}
	options := []model.QuestionOption{
		{Key: "a", Text: "1", IsCorrect: true, SortOrder: 1},
		{Key: "a", Text: "2", SortOrder: 2},
	}
	err := validateQuestion(q, options)
	if !errors.Is(err, ErrValidation) {
		t.Errorf("duplicate option key should return ErrValidation, got %v", err)
	}
	if !strings.Contains(err.Error(), "duplicate option key") {
		t.Errorf("duplicate key msg should mention 'duplicate option key', got %q", err.Error())
	}
}

func TestValidateQuestion_rejects_empty_option_text(t *testing.T) {
	q := model.Question{Format: "mcq", Body: "x"}
	options := []model.QuestionOption{
		{Key: "a", Text: "   ", IsCorrect: true, SortOrder: 1},
		{Key: "b", Text: "y", SortOrder: 2},
	}
	err := validateQuestion(q, options)
	if !errors.Is(err, ErrValidation) {
		t.Errorf("empty (whitespace) option text should return ErrValidation, got %v", err)
	}
	if !strings.Contains(err.Error(), "option text cannot be empty") {
		t.Errorf("empty option text msg should mention 'option text cannot be empty', got %q", err.Error())
	}
}

func TestValidateQuestion_mcq_rejects_correct_answer_set(t *testing.T) {
	q := model.Question{Format: "mcq", Body: "x", CorrectAnswer: strPtr("a")}
	options := []model.QuestionOption{
		{Key: "a", Text: "1", IsCorrect: true, SortOrder: 1},
		{Key: "b", Text: "2", SortOrder: 2},
	}
	err := validateQuestion(q, options)
	if !errors.Is(err, ErrValidation) {
		t.Errorf("mcq with correct_answer set should return ErrValidation, got %v", err)
	}
	if !strings.Contains(err.Error(), "mcq cannot have correct_answer") {
		t.Errorf("mcq correct_answer msg should mention 'mcq cannot have correct_answer', got %q", err.Error())
	}
}

func TestValidateTest_rejects_empty_title(t *testing.T) {
	tst := model.Test{Title: "   ", Subject: "math", Topic: "algebra", DurationMinutes: 60}
	err := validateTest(tst)
	if !errors.Is(err, ErrValidation) {
		t.Errorf("empty title should return ErrValidation, got %v", err)
	}
	if !strings.Contains(err.Error(), "test title required") {
		t.Errorf("empty title msg should mention 'test title required', got %q", err.Error())
	}
}

func TestValidateTest_rejects_zero_duration(t *testing.T) {
	tst := model.Test{Title: "x", Subject: "math", Topic: "algebra", DurationMinutes: 0}
	err := validateTest(tst)
	if !errors.Is(err, ErrValidation) {
		t.Errorf("zero duration should return ErrValidation, got %v", err)
	}
	if !strings.Contains(err.Error(), "duration_minutes must be positive") {
		t.Errorf("zero duration msg should mention 'duration_minutes must be positive', got %q", err.Error())
	}
}

func TestValidateTest_rejects_empty_subject_topic(t *testing.T) {
	tst := model.Test{Title: "x", Subject: "", Topic: "", DurationMinutes: 60}
	err := validateTest(tst)
	if !errors.Is(err, ErrValidation) {
		t.Errorf("empty subject should return ErrValidation, got %v", err)
	}
	if !strings.Contains(err.Error(), "test subject/topic required") {
		t.Errorf("empty subject msg should mention 'test subject/topic required', got %q", err.Error())
	}
}

func TestValidateTest_rejects_negative_audio_play_limit(t *testing.T) {
	tst := model.Test{Title: "x", Subject: "math", Topic: "algebra", DurationMinutes: 60, AudioPlayLimit: intptr(-1)}
	err := validateTest(tst)
	if !errors.Is(err, ErrValidation) {
		t.Errorf("negative audio_play_limit should return ErrValidation, got %v", err)
	}
	if !strings.Contains(err.Error(), "audio_play_limit must be positive") {
		t.Errorf("negative audio_play_limit msg should mention 'audio_play_limit must be positive', got %q", err.Error())
	}
}

func TestValidateTest_accepts_valid(t *testing.T) {
	tst := model.Test{Title: "Algebra 1", Subject: "math", Topic: "algebra", DurationMinutes: 60}
	if err := validateTest(tst); err != nil {
		t.Errorf("valid test should pass, got %v", err)
	}
}

// sanity: validateQuestion for a short question with non-empty correct_answer passes
func TestValidateQuestion_short_accepts_valid(t *testing.T) {
	q := model.Question{Format: "short", Body: "capital of France", CorrectAnswer: strPtr("Paris"), PointCorrect: 1}
	if err := validateQuestion(q, nil); err != nil {
		t.Errorf("valid short should pass, got %v", err)
	}
}

func TestValidateQuestion_short_rejects_whitespace_only_correct_answer(t *testing.T) {
	q := model.Question{Format: "short", Body: "x", CorrectAnswer: strPtr("   ")}
	err := validateQuestion(q, nil)
	if !errors.Is(err, ErrValidation) {
		t.Errorf("whitespace-only correct_answer should return ErrValidation, got %v", err)
	}
}

func TestValidateQuestion_empty_option_key(t *testing.T) {
	q := model.Question{Format: "mcq", Body: "x"}
	options := []model.QuestionOption{
		{Key: "", Text: "1", IsCorrect: true, SortOrder: 1},
		{Key: "b", Text: "2", SortOrder: 2},
	}
	err := validateQuestion(q, options)
	if !errors.Is(err, ErrValidation) {
		t.Errorf("empty option key should return ErrValidation, got %v", err)
	}
	if !strings.Contains(err.Error(), "option key cannot be empty") {
		t.Errorf("empty key msg should mention 'option key cannot be empty', got %q", err.Error())
	}
}

func TestValidateQuestion_multi_answer_rejects_fewer_than_2_options(t *testing.T) {
	q := model.Question{Format: "multi_answer", Body: "x"}
	options := []model.QuestionOption{
		{Key: "a", Text: "1", IsCorrect: true, SortOrder: 1},
	}
	err := validateQuestion(q, options)
	if !errors.Is(err, ErrValidation) {
		t.Errorf("multi_answer with 1 option should return ErrValidation, got %v", err)
	}
	if !strings.Contains(err.Error(), "at least 2 options") {
		t.Errorf("multi_answer 1-option msg should mention 'at least 2 options', got %q", err.Error())
	}
}

func TestValidateQuestion_fill_blank_rejects_options(t *testing.T) {
	q := model.Question{Format: "fill_blank", Body: "x", CorrectAnswer: strPtr("y")}
	options := []model.QuestionOption{
		{Key: "a", Text: "y", SortOrder: 1},
	}
	err := validateQuestion(q, options)
	if !errors.Is(err, ErrValidation) {
		t.Errorf("fill_blank with options should return ErrValidation, got %v", err)
	}
}

func TestValidateQuestion_multi_answer_rejects_correct_answer_set(t *testing.T) {
	q := model.Question{Format: "multi_answer", Body: "x", CorrectAnswer: strPtr("a")}
	options := []model.QuestionOption{
		{Key: "a", Text: "1", IsCorrect: true, SortOrder: 1},
		{Key: "b", Text: "2", SortOrder: 2},
	}
	err := validateQuestion(q, options)
	if !errors.Is(err, ErrValidation) {
		t.Errorf("multi_answer with correct_answer set should return ErrValidation, got %v", err)
	}
	if !strings.Contains(err.Error(), "multi_answer cannot have correct_answer") {
		t.Errorf("multi_answer correct_answer msg should mention 'multi_answer cannot have correct_answer', got %q", err.Error())
	}
}

func TestValidateQuestion_rejects_point_correct_below_1(t *testing.T) {
	q := model.Question{Format: "essay", Body: "explain gravity", PointCorrect: 0, PointWrong: 0}
	err := validateQuestion(q, nil)
	if !errors.Is(err, ErrValidation) {
		t.Errorf("point_correct=0 should return ErrValidation, got %v", err)
	}
	if !strings.Contains(err.Error(), "point_correct must be >= 1") {
		t.Errorf("point_correct=0 msg should mention 'point_correct must be >= 1', got %q", err.Error())
	}
}

func TestValidateQuestion_rejects_negative_point_wrong(t *testing.T) {
	q := model.Question{Format: "essay", Body: "explain gravity", PointCorrect: 1, PointWrong: -1}
	err := validateQuestion(q, nil)
	if !errors.Is(err, ErrValidation) {
		t.Errorf("point_wrong=-1 should return ErrValidation, got %v", err)
	}
	if !strings.Contains(err.Error(), "point_wrong must be >= 0") {
		t.Errorf("point_wrong=-1 msg should mention 'point_wrong must be >= 0', got %q", err.Error())
	}
}

func TestValidateQuestion_accepts_valid_points(t *testing.T) {
	q := model.Question{Format: "essay", Body: "explain gravity", PointCorrect: 2, PointWrong: 1}
	if err := validateQuestion(q, nil); err != nil {
		t.Errorf("point_correct=2, point_wrong=1 should pass, got %v", err)
	}
}

func TestValidateQuestion_rejects_body_empty_after_sanitization(t *testing.T) {
	// Simulates what every write path does: sanitize, then validate. <br> has
	// no allowlisted tag and no text content, so it sanitizes to "".
	q := model.Question{Format: "essay", Body: sanitizeQuestionBody("<br>"), PointCorrect: 1}
	err := validateQuestion(q, nil)
	if !errors.Is(err, ErrValidation) {
		t.Errorf("body that sanitizes to empty should return ErrValidation, got %v", err)
	}
	if !strings.Contains(err.Error(), "body cannot be empty") {
		t.Errorf("empty-body msg should mention 'body cannot be empty', got %q", err.Error())
	}
}

func TestValidateExam_rejects_empty_title(t *testing.T) {
	e := model.Exam{Title: "   "}
	err := validateExam(e)
	if !errors.Is(err, ErrValidation) {
		t.Errorf("empty title should return ErrValidation, got %v", err)
	}
	if !strings.Contains(err.Error(), "exam title required") {
		t.Errorf("empty title msg should mention 'exam title required', got %q", err.Error())
	}
}

func TestValidateExam_rejects_invalid_timer_mode(t *testing.T) {
	e := model.Exam{Title: "Finals", TimerMode: "freeform"}
	err := validateExam(e)
	if !errors.Is(err, ErrValidation) {
		t.Errorf("invalid timer_mode should return ErrValidation, got %v", err)
	}
	if !strings.Contains(err.Error(), "timer_mode must be overall or per_test") {
		t.Errorf("invalid timer_mode msg should mention 'timer_mode must be overall or per_test', got %q", err.Error())
	}
}

func TestValidateExam_requires_duration_when_overall(t *testing.T) {
	e := model.Exam{Title: "Finals", TimerMode: "overall"}
	err := validateExam(e)
	if !errors.Is(err, ErrValidation) {
		t.Errorf("overall with nil duration should return ErrValidation, got %v", err)
	}
	if !strings.Contains(err.Error(), "duration_minutes required and positive when timer_mode=overall") {
		t.Errorf("overall nil-duration msg should mention duration requirement, got %q", err.Error())
	}

	zero := 0
	e2 := model.Exam{Title: "Finals", TimerMode: "overall", DurationMinutes: &zero}
	err = validateExam(e2)
	if !errors.Is(err, ErrValidation) {
		t.Errorf("overall with zero duration should return ErrValidation, got %v", err)
	}
}

func TestValidateExam_accepts_valid_overall(t *testing.T) {
	e := model.Exam{Title: "Finals", TimerMode: "overall", DurationMinutes: intptr(120)}
	if err := validateExam(e); err != nil {
		t.Errorf("valid overall should pass, got %v", err)
	}
}

func TestValidateExam_accepts_valid_per_test(t *testing.T) {
	e := model.Exam{Title: "Finals", TimerMode: "per_test"}
	if err := validateExam(e); err != nil {
		t.Errorf("valid per_test should pass, got %v", err)
	}
}

func TestValidateExam_accepts_empty_timer_mode_legacy(t *testing.T) {
	e := model.Exam{Title: "Legacy", TimerMode: ""}
	if err := validateExam(e); err != nil {
		t.Errorf("empty timer_mode (legacy) should pass, got %v", err)
	}
}

func TestValidateExam_rejects_invalid_result_config(t *testing.T) {
	e := model.Exam{Title: "Finals", ResultConfig: "walkthrough"}
	err := validateExam(e)
	if !errors.Is(err, ErrValidation) {
		t.Errorf("invalid result_config should return ErrValidation, got %v", err)
	}
	if !strings.Contains(err.Error(), "result_config must be hidden, score_only, or score_pembahasan") {
		t.Errorf("invalid result_config msg should mention allowed values, got %q", err.Error())
	}
}

func TestValidateExam_accepts_empty_result_config(t *testing.T) {
	e := model.Exam{Title: "Finals", ResultConfig: ""}
	if err := validateExam(e); err != nil {
		t.Errorf("empty result_config should pass validateExam (defaulting happens in CreateExam), got %v", err)
	}
}

func TestValidateExam_accepts_each_valid_result_config(t *testing.T) {
	for _, rc := range []string{"hidden", "score_only", "score_pembahasan"} {
		e := model.Exam{Title: "Finals", ResultConfig: rc}
		if err := validateExam(e); err != nil {
			t.Errorf("result_config=%q should pass, got %v", rc, err)
		}
	}
}

// --- FR-18: mode authoring validation ---

func TestValidateExam_rejects_invalid_mode(t *testing.T) {
	e := model.Exam{Title: "Finals", Mode: "foo"}
	err := validateExam(e)
	if !errors.Is(err, ErrValidation) {
		t.Errorf("invalid mode should return ErrValidation, got %v", err)
	}
	if !strings.Contains(err.Error(), "mode must be") {
		t.Errorf("invalid mode msg should mention allowed modes, got %q", err.Error())
	}
}

func TestValidateExam_accepts_each_valid_mode(t *testing.T) {
	for _, m := range []string{"standard", "utbk", "ielts"} {
		e := model.Exam{Title: "Finals", Mode: m}
		if err := validateExam(e); err != nil {
			t.Errorf("mode=%q should pass, got %v", m, err)
		}
	}
}

func TestValidateExam_accepts_empty_mode(t *testing.T) {
	// empty on PATCH preserves; on CREATE, CreateExam defaults to standard before
	// validateExam runs. Either way validateExam must accept empty.
	e := model.Exam{Title: "Finals", Mode: ""}
	if err := validateExam(e); err != nil {
		t.Errorf("empty mode should pass validateExam (default/overlay happens in CreateExam/handler), got %v", err)
	}
}

// --- FR-18: section_type authoring validation ---

func TestValidateTest_rejects_invalid_section_type(t *testing.T) {
	invalid := "speaking"
	tst := model.Test{Title: "x", Subject: "math", Topic: "algebra", DurationMinutes: 60, SectionType: &invalid}
	err := validateTest(tst)
	if !errors.Is(err, ErrValidation) {
		t.Errorf("invalid section_type should return ErrValidation, got %v", err)
	}
	if !strings.Contains(err.Error(), "section_type must be") {
		t.Errorf("invalid section_type msg should mention allowed values, got %q", err.Error())
	}
}

func TestValidateTest_rejects_listening_without_audio_url(t *testing.T) {
	listening := "listening"
	tst := model.Test{Title: "x", Subject: "math", Topic: "algebra", DurationMinutes: 60, SectionType: &listening}
	err := validateTest(tst)
	if !errors.Is(err, ErrValidation) {
		t.Errorf("listening without audio_url should return ErrValidation, got %v", err)
	}
	if !strings.Contains(err.Error(), "audio_url required when section_type=listening") {
		t.Errorf("listening-no-audio msg should mention audio_url requirement, got %q", err.Error())
	}
}

func TestValidateTest_accepts_listening_with_audio_url(t *testing.T) {
	listening := "listening"
	audio := "https://cdn.example.com/track.mp3"
	tst := model.Test{Title: "x", Subject: "math", Topic: "algebra", DurationMinutes: 60, SectionType: &listening, AudioURL: &audio}
	if err := validateTest(tst); err != nil {
		t.Errorf("listening with audio_url should pass, got %v", err)
	}
}

func TestValidateTest_accepts_reading_section(t *testing.T) {
	reading := "reading"
	tst := model.Test{Title: "x", Subject: "math", Topic: "algebra", DurationMinutes: 60, SectionType: &reading}
	if err := validateTest(tst); err != nil {
		t.Errorf("reading section (no audio required) should pass, got %v", err)
	}
}

func TestValidateTest_accepts_null_section_type(t *testing.T) {
	// standard/utbk tests may be untyped; SectionType nil must pass.
	tst := model.Test{Title: "x", Subject: "math", Topic: "algebra", DurationMinutes: 60}
	if err := validateTest(tst); err != nil {
		t.Errorf("null section_type should pass, got %v", err)
	}
}

func TestValidateTest_accepts_writing_section(t *testing.T) {
	writing := "writing"
	tst := model.Test{Title: "x", Subject: "math", Topic: "algebra", DurationMinutes: 60, SectionType: &writing}
	if err := validateTest(tst); err != nil {
		t.Errorf("writing section should pass, got %v", err)
	}
}

// --- FR-19: publish-time completeness gate for sectioned modes ---

func entryTitled(title string, sectionType *string) model.ExamTestEntry {
	return model.ExamTestEntry{Test: struct {
		ID              uuid.UUID `json:"id"`
		Title           string    `json:"title"`
		Subject         string    `json:"subject"`
		Topic           *string   `json:"topic"`
		DurationMinutes *int      `json:"duration_minutes"`
		SectionType     *string   `json:"section_type,omitempty"`
		QuestionCount   int       `json:"question_count"`
	}{Title: title, SectionType: sectionType}}
}

func TestValidatePublishSections_rejects_sectioned_exam_with_zero_tests(t *testing.T) {
	for _, mode := range []string{"utbk", "ielts"} {
		exam := model.Exam{Mode: mode}
		err := validatePublishSections(exam, nil)
		if !errors.Is(err, ErrValidation) {
			t.Errorf("mode=%s with 0 tests should return ErrValidation, got %v", mode, err)
		}
		if !strings.Contains(err.Error(), "at least one test") {
			t.Errorf("zero-tests msg should mention 'at least one test', got %q", err.Error())
		}
		err = validatePublishSections(exam, []model.ExamTestEntry{})
		if !errors.Is(err, ErrValidation) {
			t.Errorf("mode=%s with empty tests slice should return ErrValidation, got %v", mode, err)
		}
	}
}

func TestValidatePublishSections_rejects_ielts_with_untyped_section(t *testing.T) {
	exam := model.Exam{Mode: "ielts"}
	tests := []model.ExamTestEntry{
		entryTitled("Listening", strPtr("listening")),
		entryTitled("Untyped Section", nil),
	}
	err := validatePublishSections(exam, tests)
	if !errors.Is(err, ErrValidation) {
		t.Errorf("ielts with an untyped attached section should return ErrValidation, got %v", err)
	}
	if !strings.Contains(err.Error(), "Untyped Section") {
		t.Errorf("ielts untyped-section msg should name the offending section, got %q", err.Error())
	}
}

func TestValidatePublishSections_allows_fully_typed_ielts(t *testing.T) {
	exam := model.Exam{Mode: "ielts"}
	tests := []model.ExamTestEntry{
		entryTitled("Listening", strPtr("listening")),
		entryTitled("Reading", strPtr("reading")),
		entryTitled("Writing", strPtr("writing")),
	}
	if err := validatePublishSections(exam, tests); err != nil {
		t.Errorf("fully-typed ielts should pass, got %v", err)
	}
}

func TestValidatePublishSections_allows_utbk_with_untyped_tests(t *testing.T) {
	// utbk may have untyped tests per spec (FR-19); only duration_minutes>0 is
	// enforced, and that already lives in validateTest.
	exam := model.Exam{Mode: "utbk"}
	tests := []model.ExamTestEntry{
		entryTitled("Subtest 1", nil),
		entryTitled("Subtest 2", strPtr("reading")),
	}
	if err := validatePublishSections(exam, tests); err != nil {
		t.Errorf("utbk with a mix of untyped/typed tests should pass, got %v", err)
	}
}

func TestValidatePublishSections_allows_standard_with_any_tests(t *testing.T) {
	// standard publish is unchanged; the gate is skipped entirely.
	exam := model.Exam{Mode: "standard"}
	if err := validatePublishSections(exam, nil); err != nil {
		t.Errorf("standard with no tests should pass (gate skipped), got %v", err)
	}
	if err := validatePublishSections(exam, []model.ExamTestEntry{entryTitled("Any", nil)}); err != nil {
		t.Errorf("standard with untyped test should pass (gate skipped), got %v", err)
	}
}

func TestValidatePublishSections_allows_empty_mode(t *testing.T) {
	// empty mode (legacy rows / pre-default) must not trigger the gate.
	exam := model.Exam{Mode: ""}
	if err := validatePublishSections(exam, nil); err != nil {
		t.Errorf("empty mode should pass (gate skipped), got %v", err)
	}
}

func TestCheckTypeRBAC_admin_exam_allows_exam(t *testing.T) {
	if err := checkTypeRBAC(RoleAdminExam, "exam"); err != nil {
		t.Errorf("admin_exam on exam type should be allowed, got %v", err)
	}
}

func TestCheckTypeRBAC_admin_exam_blocks_book(t *testing.T) {
	err := checkTypeRBAC(RoleAdminExam, "book")
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("admin_exam on book type should return ErrForbidden, got %v", err)
	}
}

func TestCheckTypeRBAC_admin_exam_blocks_course(t *testing.T) {
	err := checkTypeRBAC(RoleAdminExam, "course")
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("admin_exam on course type should return ErrForbidden, got %v", err)
	}
}

// --- FR-9..FR-15: bank question CRUD + delete-guard + list-bank ---

func seedTopicDirect(t *testing.T, ctx context.Context, repo *repository.Repository, name, subject string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	err := repo.Pool().QueryRow(ctx,
		`INSERT INTO exam_topic (name, subject) VALUES ($1, $2) RETURNING id`,
		name, subject,
	).Scan(&id)
	require.NoError(t, err)
	return id
}

func seedTestDirect(t *testing.T, ctx context.Context, repo *repository.Repository, title, subject, topic string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	err := repo.Pool().QueryRow(ctx,
		`INSERT INTO test (title, subject, topic, duration_minutes) VALUES ($1, $2, $3, $4) RETURNING id`,
		title, subject, topic, 60,
	).Scan(&id)
	require.NoError(t, err)
	return id
}

func seedBankQuestionDirect(t *testing.T, ctx context.Context, repo *repository.Repository, format, body string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	err := repo.Pool().QueryRow(ctx,
		`INSERT INTO question (format, body, point_correct, point_wrong) VALUES ($1, $2, 1, 0) RETURNING id`,
		format, body,
	).Scan(&id)
	require.NoError(t, err)
	return id
}

// seedExamWithTestsDirect creates an exam and attaches the given tests to it via
// exam_test, in the order given.
func seedExamWithTestsDirect(t *testing.T, ctx context.Context, repo *repository.Repository, testIDs ...uuid.UUID) uuid.UUID {
	t.Helper()
	var examID uuid.UUID
	err := repo.Pool().QueryRow(ctx,
		`INSERT INTO exam (title, status) VALUES ($1, 'draft') RETURNING id`,
		"Exam "+uniqueSuffix(),
	).Scan(&examID)
	require.NoError(t, err)
	for i, tid := range testIDs {
		_, err := repo.Pool().Exec(ctx,
			`INSERT INTO exam_test (exam_id, test_id, sort_order) VALUES ($1, $2, $3)`,
			examID, tid, i+1,
		)
		require.NoError(t, err)
	}
	return examID
}

func attachQuestionDirect(t *testing.T, ctx context.Context, repo *repository.Repository, testID, questionID uuid.UUID, sortOrder int) {
	t.Helper()
	_, err := repo.Pool().Exec(ctx,
		`INSERT INTO test_question (test_id, question_id, sort_order) VALUES ($1, $2, $3)`,
		testID, questionID, sortOrder,
	)
	require.NoError(t, err)
}

func answerQuestionDirect(t *testing.T, ctx context.Context, repo *repository.Repository, questionID uuid.UUID) {
	t.Helper()
	// exam_session_answer requires a session; create the minimal session row.
	var studentID uuid.UUID
	err := repo.Pool().QueryRow(ctx,
		`INSERT INTO users (email, name, role, status) VALUES ($1, $2, 'student', 'active') RETURNING id`,
		"student-"+uniqueSuffix()+"@example.com", "Student",
	).Scan(&studentID)
	require.NoError(t, err)
	var examID uuid.UUID
	err = repo.Pool().QueryRow(ctx,
		`INSERT INTO exam (title, status) VALUES ($1, 'draft') RETURNING id`,
		"Exam "+uniqueSuffix(),
	).Scan(&examID)
	require.NoError(t, err)
	var regID uuid.UUID
	err = repo.Pool().QueryRow(ctx,
		`INSERT INTO exam_registration (student_id, exam_id, token, status) VALUES ($1, $2, $3, 'registered') RETURNING id`,
		studentID, examID, "TOKEN"+uniqueSuffix(),
	).Scan(&regID)
	require.NoError(t, err)
	var sessionID uuid.UUID
	err = repo.Pool().QueryRow(ctx,
		`INSERT INTO exam_session (registration_id, student_id, exam_id, attempt_number, started_at, status) VALUES ($1, $2, $3, 1, now(), 'submitted') RETURNING id`,
		regID, studentID, examID,
	).Scan(&sessionID)
	require.NoError(t, err)
	_, err = repo.Pool().Exec(ctx,
		`INSERT INTO exam_session_answer (session_id, question_id, answer, saved_at) VALUES ($1, $2, $3, now())`,
		sessionID, questionID, "answer",
	)
	require.NoError(t, err)
}

func countQuestionAttachments(t *testing.T, ctx context.Context, repo *repository.Repository, id uuid.UUID) int {
	t.Helper()
	var count int
	err := repo.Pool().QueryRow(ctx,
		`SELECT COUNT(*) FROM test_question WHERE question_id = $1`, id,
	).Scan(&count)
	require.NoError(t, err)
	return count
}

func listTestQuestions(t *testing.T, ctx context.Context, svc *Service, testID uuid.UUID) []model.QuestionWithOptions {
	t.Helper()
	detail, err := svc.GetTestDetail(ctx, testID)
	require.NoError(t, err)
	return detail.Questions
}

func TestCreateBankQuestion_creates_no_attachment(t *testing.T) {
	svc, repo := newRealDBService(t)
	ctx := context.Background()

	q := model.Question{Format: "essay", Body: "explain gravity", PointCorrect: 1, PointWrong: 0}
	out, err := svc.CreateBankQuestion(ctx, q, nil)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, out.Question.ID)
	assert.Equal(t, "essay", out.Question.Format)
	assert.Equal(t, 0, countQuestionAttachments(t, ctx, repo, out.Question.ID))
}

func TestCreateBankQuestion_rejects_body_that_sanitizes_to_empty(t *testing.T) {
	svc, _ := newRealDBService(t)
	ctx := context.Background()

	// <br> is not in questionBodyPolicy's allowlist and carries no text
	// content, so sanitizeQuestionBody reduces it to "" before validateQuestion
	// runs — a blank question must not be persisted.
	q := model.Question{Format: "essay", Body: "<br>", PointCorrect: 1, PointWrong: 0}
	_, err := svc.CreateBankQuestion(ctx, q, nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrValidation)

	items, _, err := svc.ListBankQuestions(ctx, repository.QuestionFilter{Search: "<br>", Limit: 10})
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestListBankQuestions_populates_nested_question_and_options(t *testing.T) {
	svc, _ := newRealDBService(t)
	ctx := context.Background()

	// multi_answer (not mcq) so this doesn't collide with the unscoped
	// Format:"mcq" assertion in TestListBankQuestions_filters_and_counts_used_in.
	body := "bank list shape " + uniqueSuffix()
	q := model.Question{Format: "multi_answer", Body: body, PointCorrect: 1, PointWrong: 0}
	opts := []model.QuestionOption{
		{Key: "a", Text: "yes", IsCorrect: true, SortOrder: 1},
		{Key: "b", Text: "no", IsCorrect: false, SortOrder: 2},
	}
	created, err := svc.CreateBankQuestion(ctx, q, opts)
	require.NoError(t, err)

	items, _, err := svc.ListBankQuestions(ctx, repository.QuestionFilter{Search: body, Limit: 10})
	require.NoError(t, err)
	require.Len(t, items, 1)

	// Nested {question, options, attached_count} shape (not flattened/embedded) —
	// the admin bank page destructures item.question and reads item.options.
	assert.Equal(t, created.Question.ID, items[0].Question.ID)
	assert.Equal(t, body, items[0].Question.Body)
	require.Len(t, items[0].Options, 2)
	assert.Equal(t, "a", items[0].Options[0].Key)
	assert.Equal(t, "b", items[0].Options[1].Key)
}

// A fill_blank / short / essay question has no options. Its Options must
// serialize as [] not null — a nil slice becomes JSON null and crashes the
// admin question editor, which reads q.options.length when opening an edit.
func TestListBankQuestions_optionlessFormat_returnsNonNilOptions(t *testing.T) {
	svc, repo := newRealDBService(t)
	ctx := context.Background()

	body := "fill blank no options " + uniqueSuffix()
	seedBankQuestionDirect(t, ctx, repo, "fill_blank", body)

	items, _, err := svc.ListBankQuestions(ctx, repository.QuestionFilter{Search: body, Limit: 10})
	require.NoError(t, err)
	require.Len(t, items, 1)

	assert.NotNil(t, items[0].Options, "options must be a non-nil empty slice, not nil (serializes to null)")
	assert.Len(t, items[0].Options, 0)
}

func TestDeleteQuestion_rejects_when_attached(t *testing.T) {
	svc, repo := newRealDBService(t)
	ctx := context.Background()

	testID := seedTestDirect(t, ctx, repo, "Math "+uniqueSuffix(), "math", "algebra")
	qID := seedBankQuestionDirect(t, ctx, repo, "essay", "explain")
	attachQuestionDirect(t, ctx, repo, testID, qID, 1)

	err := svc.DeleteQuestion(ctx, qID)
	assert.ErrorIs(t, err, ErrValidation)
	assert.Contains(t, err.Error(), "attached")

	// Guard must be a no-op: the question and attachment survive.
	assert.Equal(t, 1, countQuestionAttachments(t, ctx, repo, qID))
}

func TestDeleteQuestion_rejects_when_answered(t *testing.T) {
	svc, repo := newRealDBService(t)
	ctx := context.Background()

	qID := seedBankQuestionDirect(t, ctx, repo, "short", "capital of France")
	answerQuestionDirect(t, ctx, repo, qID)

	err := svc.DeleteQuestion(ctx, qID)
	assert.ErrorIs(t, err, ErrValidation)
	assert.Contains(t, err.Error(), "answered")

	// Guard must be a no-op: the question survives.
	var exists bool
	require.NoError(t, repo.Pool().QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM question WHERE id = $1)`, qID).Scan(&exists))
	assert.True(t, exists)
}

func TestDeleteQuestion_succeeds_when_unattached_and_unanswered(t *testing.T) {
	svc, repo := newRealDBService(t)
	ctx := context.Background()

	qID := seedBankQuestionDirect(t, ctx, repo, "essay", "explain relativity")

	err := svc.DeleteQuestion(ctx, qID)
	require.NoError(t, err)

	var exists bool
	require.NoError(t, repo.Pool().QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM question WHERE id = $1)`, qID).Scan(&exists))
	assert.False(t, exists)
}

func TestListBankQuestions_filters_and_counts_used_in(t *testing.T) {
	svc, repo := newRealDBService(t)
	ctx := context.Background()

	topicA := seedTopicDirect(t, ctx, repo, "Algebra "+uniqueSuffix(), "math")
	topicB := seedTopicDirect(t, ctx, repo, "Geometry "+uniqueSuffix(), "math")

	// Three questions: mcq in topicA (attached to 2 tests), essay in topicB (unattached), short no topic.
	test1 := seedTestDirect(t, ctx, repo, "T1 "+uniqueSuffix(), "math", "algebra")
	test2 := seedTestDirect(t, ctx, repo, "T2 "+uniqueSuffix(), "math", "algebra")

	uniqueToken := "cursorbatch " + uniqueSuffix()

	mcqID := seedBankQuestionDirect(t, ctx, repo, "mcq", uniqueToken+" 2+2")
	_, err := repo.Pool().Exec(ctx, `UPDATE question SET topic_id = $1 WHERE id = $2`, topicA, mcqID)
	require.NoError(t, err)
	attachQuestionDirect(t, ctx, repo, test1, mcqID, 1)
	attachQuestionDirect(t, ctx, repo, test2, mcqID, 2)

	essayBody := uniqueToken + " explain photosynthesis " + uniqueSuffix()
	essayID := seedBankQuestionDirect(t, ctx, repo, "essay", essayBody)
	_, err = repo.Pool().Exec(ctx, `UPDATE question SET topic_id = $1 WHERE id = $2`, topicB, essayID)
	require.NoError(t, err)

	shortID := seedBankQuestionDirect(t, ctx, repo, "short", uniqueToken+" short")

	// Full list (filtered by unique token) returns exactly the three bank questions.
	all, _, err := svc.ListBankQuestions(ctx, repository.QuestionFilter{Search: uniqueToken, Limit: 50})
	require.NoError(t, err)
	ids := map[uuid.UUID]bool{}
	for _, it := range all {
		ids[it.Question.ID] = true
	}
	assert.True(t, ids[mcqID] && ids[essayID] && ids[shortID], "expected all three bank questions")

	// Filter by format.
	items, nextCursor, err := svc.ListBankQuestions(ctx, repository.QuestionFilter{Format: "mcq", Limit: 10})
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, mcqID, items[0].Question.ID)
	assert.Equal(t, 2, items[0].AttachedCount)
	assert.Empty(t, nextCursor)

	// Filter by topic_id.
	items, nextCursor, err = svc.ListBankQuestions(ctx, repository.QuestionFilter{TopicID: topicB.String(), Limit: 10})
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, essayID, items[0].Question.ID)
	assert.Equal(t, 0, items[0].AttachedCount)
	assert.Empty(t, nextCursor)

	// Search by body substring (unique term so leftover rows don't match).
	items, nextCursor, err = svc.ListBankQuestions(ctx, repository.QuestionFilter{Search: "photosynthesis", Limit: 10})
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, essayID, items[0].Question.ID)
	assert.Empty(t, nextCursor)

	// Cursor pagination: limit 2 on the unique-token batch should give first two rows and a cursor.
	items, nextCursor, err = svc.ListBankQuestions(ctx, repository.QuestionFilter{Search: uniqueToken, Limit: 2})
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.NotEmpty(t, nextCursor)
	page1IDs := map[uuid.UUID]bool{items[0].Question.ID: true, items[1].Question.ID: true}

	// Follow cursor should return the remaining row.
	items, nextCursor, err = svc.ListBankQuestions(ctx, repository.QuestionFilter{Search: uniqueToken, Limit: 2, Cursor: nextCursor})
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Empty(t, nextCursor)
	assert.False(t, page1IDs[items[0].Question.ID], "cursor should advance to a new row")
}

// --- FR-21..FR-25: test ↔ question attach / detach / reorder ---

func TestAttachQuestions_appends_after_max_order_and_is_idempotent(t *testing.T) {
	svc, repo := newRealDBService(t)
	ctx := context.Background()

	testID := seedTestDirect(t, ctx, repo, "Attach "+uniqueSuffix(), "math", "algebra")
	q1 := seedBankQuestionDirect(t, ctx, repo, "short", "q1 "+uniqueSuffix())
	q2 := seedBankQuestionDirect(t, ctx, repo, "short", "q2 "+uniqueSuffix())
	q3 := seedBankQuestionDirect(t, ctx, repo, "short", "q3 "+uniqueSuffix())

	// First attach: q1 and q2 get orders 1 and 2.
	require.NoError(t, svc.AttachQuestions(ctx, testID, []uuid.UUID{q1, q2}))
	questions := listTestQuestions(t, ctx, svc, testID)
	require.Len(t, questions, 2)
	assert.Equal(t, q1, questions[0].Question.ID)
	assert.Equal(t, 1, questions[0].SortOrder)
	assert.Equal(t, q2, questions[1].Question.ID)
	assert.Equal(t, 2, questions[1].SortOrder)

	// Second attach includes an already-attached q2 plus a new q3: q3 appends as order 3.
	require.NoError(t, svc.AttachQuestions(ctx, testID, []uuid.UUID{q2, q3}))
	questions = listTestQuestions(t, ctx, svc, testID)
	require.Len(t, questions, 3)
	assert.Equal(t, q3, questions[2].Question.ID)
	assert.Equal(t, 3, questions[2].SortOrder)
}

func TestAttachQuestions_rejects_missing_test(t *testing.T) {
	svc, _ := newRealDBService(t)
	ctx := context.Background()

	missingTest := uuid.New()
	q := uuid.New()
	err := svc.AttachQuestions(ctx, missingTest, []uuid.UUID{q})
	assert.ErrorIs(t, err, ErrTestNotFound)
}

func TestAttachQuestions_rejects_missing_question(t *testing.T) {
	svc, repo := newRealDBService(t)
	ctx := context.Background()

	testID := seedTestDirect(t, ctx, repo, "Attach "+uniqueSuffix(), "math", "algebra")
	realQ := seedBankQuestionDirect(t, ctx, repo, "short", "real "+uniqueSuffix())
	missingQ := uuid.New()

	err := svc.AttachQuestions(ctx, testID, []uuid.UUID{realQ, missingQ})
	assert.ErrorIs(t, err, ErrQuestionNotFound)

	// No partial attachment must occur.
	assert.Equal(t, 0, countQuestionAttachments(t, ctx, repo, realQ))
}

func TestDetachQuestion_removes_only_join(t *testing.T) {
	svc, repo := newRealDBService(t)
	ctx := context.Background()

	testA := seedTestDirect(t, ctx, repo, "A "+uniqueSuffix(), "math", "algebra")
	testB := seedTestDirect(t, ctx, repo, "B "+uniqueSuffix(), "math", "algebra")
	q := seedBankQuestionDirect(t, ctx, repo, "short", "shared "+uniqueSuffix())
	attachQuestionDirect(t, ctx, repo, testA, q, 0)
	attachQuestionDirect(t, ctx, repo, testB, q, 0)

	require.NoError(t, svc.DetachQuestion(ctx, testA, q))

	assert.Equal(t, 1, countQuestionAttachments(t, ctx, repo, q))
	questionsA := listTestQuestions(t, ctx, svc, testA)
	assert.Len(t, questionsA, 0)
	questionsB := listTestQuestions(t, ctx, svc, testB)
	assert.Len(t, questionsB, 1)

	// Bank question survives.
	var exists bool
	require.NoError(t, repo.Pool().QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM question WHERE id = $1)`, q).Scan(&exists))
	assert.True(t, exists)
}

func TestDetachQuestion_rejects_missing_test(t *testing.T) {
	svc, _ := newRealDBService(t)
	ctx := context.Background()

	err := svc.DetachQuestion(ctx, uuid.New(), uuid.New())
	assert.ErrorIs(t, err, ErrTestNotFound)
}

func TestReorderTestQuestions_rewrites_order_without_conflict(t *testing.T) {
	svc, repo := newRealDBService(t)
	ctx := context.Background()

	testID := seedTestDirect(t, ctx, repo, "Reorder "+uniqueSuffix(), "math", "algebra")
	q1 := seedBankQuestionDirect(t, ctx, repo, "short", "r1 "+uniqueSuffix())
	q2 := seedBankQuestionDirect(t, ctx, repo, "short", "r2 "+uniqueSuffix())
	q3 := seedBankQuestionDirect(t, ctx, repo, "short", "r3 "+uniqueSuffix())
	attachQuestionDirect(t, ctx, repo, testID, q1, 0)
	attachQuestionDirect(t, ctx, repo, testID, q2, 1)
	attachQuestionDirect(t, ctx, repo, testID, q3, 2)

	// Reverse the order.
	require.NoError(t, svc.ReorderTestQuestions(ctx, testID, []uuid.UUID{q3, q2, q1}))

	questions := listTestQuestions(t, ctx, svc, testID)
	require.Len(t, questions, 3)
	assert.Equal(t, q3, questions[0].Question.ID)
	assert.Equal(t, 0, questions[0].SortOrder)
	assert.Equal(t, q2, questions[1].Question.ID)
	assert.Equal(t, 1, questions[1].SortOrder)
	assert.Equal(t, q1, questions[2].Question.ID)
	assert.Equal(t, 2, questions[2].SortOrder)
}

func TestReorderTestQuestions_rejects_mismatched_set(t *testing.T) {
	svc, repo := newRealDBService(t)
	ctx := context.Background()

	testID := seedTestDirect(t, ctx, repo, "Reorder "+uniqueSuffix(), "math", "algebra")
	q1 := seedBankQuestionDirect(t, ctx, repo, "short", "m1 "+uniqueSuffix())
	q2 := seedBankQuestionDirect(t, ctx, repo, "short", "m2 "+uniqueSuffix())
	attachQuestionDirect(t, ctx, repo, testID, q1, 0)
	attachQuestionDirect(t, ctx, repo, testID, q2, 1)

	// Missing q2, extra q3 (not attached).
	q3 := seedBankQuestionDirect(t, ctx, repo, "short", "m3 "+uniqueSuffix())
	err := svc.ReorderTestQuestions(ctx, testID, []uuid.UUID{q1, q3})
	assert.ErrorIs(t, err, ErrValidation)
	assert.Contains(t, err.Error(), "must match the current attached set")
}

func TestReorderTestQuestions_rejects_duplicate_id_masquerading_as_full_set(t *testing.T) {
	svc, repo := newRealDBService(t)
	ctx := context.Background()

	testID := seedTestDirect(t, ctx, repo, "Reorder "+uniqueSuffix(), "math", "algebra")
	q1 := seedBankQuestionDirect(t, ctx, repo, "short", "m1 "+uniqueSuffix())
	q2 := seedBankQuestionDirect(t, ctx, repo, "short", "m2 "+uniqueSuffix())
	attachQuestionDirect(t, ctx, repo, testID, q1, 0)
	attachQuestionDirect(t, ctx, repo, testID, q2, 1)

	// Same length as the attached set, but q1 repeated and q2 missing entirely.
	err := svc.ReorderTestQuestions(ctx, testID, []uuid.UUID{q1, q1})
	assert.ErrorIs(t, err, ErrValidation)
	assert.Contains(t, err.Error(), "must match the current attached set")
}

func TestAttachQuestions_rejects_question_already_on_sibling_test_in_same_exam(t *testing.T) {
	svc, repo := newRealDBService(t)
	ctx := context.Background()

	test1 := seedTestDirect(t, ctx, repo, "T1 "+uniqueSuffix(), "math", "algebra")
	test2 := seedTestDirect(t, ctx, repo, "T2 "+uniqueSuffix(), "math", "algebra")
	seedExamWithTestsDirect(t, ctx, repo, test1, test2)

	qID := seedBankQuestionDirect(t, ctx, repo, "short", "shared "+uniqueSuffix())
	attachQuestionDirect(t, ctx, repo, test1, qID, 1)

	err := svc.AttachQuestions(ctx, test2, []uuid.UUID{qID})
	assert.ErrorIs(t, err, ErrValidation)
	assert.Contains(t, err.Error(), "already attached to another test in the same exam")

	// Guard is a no-op: question remains attached only to test1.
	assert.Equal(t, 1, countQuestionAttachments(t, ctx, repo, qID))
}

func TestAttachQuestions_allows_reattaching_to_its_own_test(t *testing.T) {
	svc, repo := newRealDBService(t)
	ctx := context.Background()

	testID := seedTestDirect(t, ctx, repo, "T1 "+uniqueSuffix(), "math", "algebra")
	seedExamWithTestsDirect(t, ctx, repo, testID)
	qID := seedBankQuestionDirect(t, ctx, repo, "short", "self "+uniqueSuffix())
	attachQuestionDirect(t, ctx, repo, testID, qID, 1)

	// Idempotent re-attach to the SAME test must not be blocked by the sibling guard.
	err := svc.AttachQuestions(ctx, testID, []uuid.UUID{qID})
	require.NoError(t, err)
}

func TestAttachQuestions_allows_question_shared_across_different_exams(t *testing.T) {
	svc, repo := newRealDBService(t)
	ctx := context.Background()

	test1 := seedTestDirect(t, ctx, repo, "T1 "+uniqueSuffix(), "math", "algebra")
	test2 := seedTestDirect(t, ctx, repo, "T2 "+uniqueSuffix(), "math", "algebra")
	seedExamWithTestsDirect(t, ctx, repo, test1)
	seedExamWithTestsDirect(t, ctx, repo, test2)

	qID := seedBankQuestionDirect(t, ctx, repo, "short", "crossexam "+uniqueSuffix())
	attachQuestionDirect(t, ctx, repo, test1, qID, 1)

	// Same question reused across tests in DIFFERENT exams is fine — only
	// sibling tests inside the SAME exam collide.
	err := svc.AttachQuestions(ctx, test2, []uuid.UUID{qID})
	require.NoError(t, err)
}

func TestCreateQuestionForTest_creates_bank_question_and_join(t *testing.T) {
	svc, repo := newRealDBService(t)
	ctx := context.Background()

	testID := seedTestDirect(t, ctx, repo, "CreateInTest "+uniqueSuffix(), "math", "algebra")
	// Pre-attach one question so the new one appends after it.
	existingQ := seedBankQuestionDirect(t, ctx, repo, "short", "existing "+uniqueSuffix())
	attachQuestionDirect(t, ctx, repo, testID, existingQ, 0)

	q := model.Question{Format: "essay", Body: "explain relativity", PointCorrect: 1, PointWrong: 0}
	out, err := svc.CreateQuestionForTest(ctx, testID, q, nil)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, out.Question.ID)

	// It lives in the bank.
	var exists bool
	require.NoError(t, repo.Pool().QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM question WHERE id = $1)`, out.Question.ID).Scan(&exists))
	assert.True(t, exists)

	// It is attached to the test as the last item.
	questions := listTestQuestions(t, ctx, svc, testID)
	require.Len(t, questions, 2)
	assert.Equal(t, out.Question.ID, questions[1].Question.ID)
	assert.Equal(t, 1, questions[1].SortOrder)
}

// suppress unused: uuid is imported to avoid unused-import lint if tests get trimmed later
var _ = uuid.Nil

// --- FR-18/19: CreateExam default mode + PublishProduct's sectioned gate (integration) ---
// These exercise the service against the real Postgres fixture (testcontainers),
// matching the existing school_test.go pattern. They verify the CreateExam
// defaulting and that PublishProduct, for an exam-type product, loads every
// attached exam's Tests and delegates to validatePublishSections.

func TestCreateExam_Integration_DefaultsModeToStandard(t *testing.T) {
	svc, _ := newRealDBService(t)
	ctx := context.Background()

	title := "Default Mode Exam " + uniqueSuffix()
	exam, err := svc.CreateExam(ctx, model.Exam{Title: title, Mode: ""})
	if err != nil {
		t.Fatalf("CreateExam: %v", err)
	}
	if exam.Mode != "standard" {
		t.Errorf("CreateExam with empty Mode should default to standard, got %q", exam.Mode)
	}

	// explicit mode must round-trip unchanged.
	exam2, err := svc.CreateExam(ctx, model.Exam{Title: "UTBK Exam " + uniqueSuffix(), Mode: "utbk"})
	if err != nil {
		t.Fatalf("CreateExam utbk: %v", err)
	}
	if exam2.Mode != "utbk" {
		t.Errorf("CreateExam with mode=utbk should persist utbk, got %q", exam2.Mode)
	}
}

func TestPublishProduct_Integration_RejectsSectionedExamWithNoTests(t *testing.T) {
	svc, _ := newRealDBService(t)
	ctx := context.Background()

	exam, err := svc.CreateExam(ctx, model.Exam{Title: "UTBK No-Tests " + uniqueSuffix(), Mode: "utbk"})
	if err != nil {
		t.Fatalf("CreateExam: %v", err)
	}
	product, err := svc.CreateProductWithExams(ctx, model.Product{Type: "exam", Name: exam.Title, Price: 0, Status: "draft"}, []string{exam.ID.String()}, RoleAdminStore)
	if err != nil {
		t.Fatalf("CreateProductWithExams: %v", err)
	}
	err = svc.PublishProduct(ctx, product.ID, RoleAdminStore)
	if !errors.Is(err, ErrValidation) {
		t.Errorf("PublishProduct on a product attaching a utbk exam with 0 tests should return ErrValidation, got %v", err)
	}
}

func TestPublishProduct_Integration_StandardExamSkipsSectionGate(t *testing.T) {
	svc, _ := newRealDBService(t)
	ctx := context.Background()

	// Standard exam with no tests must NOT be rejected by the section gate — it
	// proceeds to the underlying product publish (which may then fail for other
	// product reasons, but not with the sectioned-mode ErrValidation). We assert
	// only that the error is not the sectioned-zero-tests validation.
	exam, err := svc.CreateExam(ctx, model.Exam{Title: "Standard No-Tests " + uniqueSuffix(), Mode: "standard"})
	if err != nil {
		t.Fatalf("CreateExam: %v", err)
	}
	product, err := svc.CreateProductWithExams(ctx, model.Product{Type: "exam", Name: exam.Title, Price: 0, Status: "draft"}, []string{exam.ID.String()}, RoleAdminStore)
	if err != nil {
		t.Fatalf("CreateProductWithExams: %v", err)
	}
	err = svc.PublishProduct(ctx, product.ID, RoleAdminStore)
	if err != nil && strings.Contains(err.Error(), "sectioned exam") {
		t.Errorf("standard exam must not hit the sectioned gate, got %v", err)
	}
}

// --- Slice 3: registration reads + exam card ---

// fakeRegRepo is a minimal stub for the repository methods needed by
// GetExamRegistration / GetExamCard. storeRepo is a concrete *repository.Repository
// in the production Service, so we replicate the relevant logic via a shim that
// matches the existing student_test.go / store_test.go patterns.
type fakeRegRepo struct {
	regsByIDStudent map[[2]uuid.UUID]*model.RegistrationDetail
}

func newFakeRegRepo() *fakeRegRepo {
	return &fakeRegRepo{
		regsByIDStudent: map[[2]uuid.UUID]*model.RegistrationDetail{},
	}
}

func (f *fakeRegRepo) seed(reg model.RegistrationDetail) {
	f.regsByIDStudent[[2]uuid.UUID{reg.ExamRegistration.ID, reg.ExamRegistration.StudentID}] = &reg
}

func (f *fakeRegRepo) GetExamRegistrationByID(_ context.Context, regID, studentID uuid.UUID) (*model.RegistrationDetail, error) {
	key := [2]uuid.UUID{regID, studentID}
	if d, ok := f.regsByIDStudent[key]; ok {
		cp := *d
		return &cp, nil
	}
	return nil, repository.ErrNotFound
}

// shimRegistrationService mirrors Service.GetExamRegistration against a fakeRegRepo.
type shimRegistrationService struct {
	fake *fakeRegRepo
}

func (s *shimRegistrationService) GetExamRegistration(ctx context.Context, regID, studentID string) (*model.RegistrationDetail, error) {
	rid, err := uuid.Parse(regID)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid registration id", ErrValidation)
	}
	sid, err := uuid.Parse(studentID)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid student id", ErrValidation)
	}
	detail, err := s.fake.GetExamRegistrationByID(ctx, rid, sid)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrRegistrationNotFound
	}
	return detail, err
}

func TestGetExamRegistration_NotOwned_ReturnsErrRegistrationNotFound(t *testing.T) {
	ctx := context.Background()
	fake := newFakeRegRepo()

	owner := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	otherStudent := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	regID := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	detail := model.RegistrationDetail{}
	detail.ExamRegistration = model.ExamRegistration{
		ID:        regID,
		StudentID: owner,
		Token:     "ABCD1234",
		Status:    "registered",
	}
	fake.seed(detail)

	svc := &shimRegistrationService{fake: fake}

	_, err := svc.GetExamRegistration(ctx, regID.String(), otherStudent.String())
	if !errors.Is(err, ErrRegistrationNotFound) {
		t.Errorf("non-owner should get ErrRegistrationNotFound, got %v", err)
	}

	_, err = svc.GetExamRegistration(ctx, regID.String(), owner.String())
	if err != nil {
		t.Errorf("owner lookup failed, got %v", err)
	}

	absent := uuid.MustParse("99999999-9999-9999-9999-999999999999")
	_, err = svc.GetExamRegistration(ctx, absent.String(), owner.String())
	if !errors.Is(err, ErrRegistrationNotFound) {
		t.Errorf("absent id should return ErrRegistrationNotFound, got %v", err)
	}
}

// shimExamCardService mirrors Service.GetExamCard against a fakeRegRepo; injected
// studentName and tenantName stand in for the system_config + Me lookups.
type shimExamCardService struct {
	fake        *fakeRegRepo
	studentName string
	tenantName  string
}

func (s *shimExamCardService) GetExamCard(ctx context.Context, regID, studentID string) ([]byte, string, error) {
	detail, err := s.fake.GetExamRegistrationByID(ctx, mustParse(regID), mustParse(studentID))
	if errors.Is(err, repository.ErrNotFound) {
		return nil, "", ErrRegistrationNotFound
	}
	if err != nil {
		return nil, "", err
	}
	pdf, err := generateExamCardPDF(detail, s.studentName, s.tenantName)
	if err != nil {
		return nil, "", err
	}
	return pdf, "kartu-peserta-" + detail.Token + ".pdf", nil
}

func mustParse(s string) uuid.UUID {
	v, err := uuid.Parse(s)
	if err != nil {
		panic(err)
	}
	return v
}

func TestGetExamCard_ReturnsPdfBytes(t *testing.T) {
	ctx := context.Background()
	fake := newFakeRegRepo()

	studentID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	examID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	regID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")

	detail := model.RegistrationDetail{}
	detail.ExamRegistration = model.ExamRegistration{
		ID:        regID,
		StudentID: studentID,
		ExamID:    examID,
		Token:     "AB12CD34",
		Status:    "registered",
	}
	detail.Exam.ID = examID
	detail.Exam.Title = "Finals"
	detail.Exam.RequiresCheckin = false
	fake.seed(detail)

	svc := &shimExamCardService{
		fake:        fake,
		studentName: "Saifullah",
		tenantName:  "Akademi Bimbel",
	}

	pdf, filename, err := svc.GetExamCard(ctx, regID.String(), studentID.String())
	if err != nil {
		t.Fatalf("GetExamCard: %v", err)
	}

	if len(pdf) < 5 {
		t.Fatalf("PDF bytes too short: %d", len(pdf))
	}
	if !bytes.HasPrefix(pdf, []byte("%PDF-")) {
		t.Errorf("PDF bytes should start with %q, got %q", "%PDF-", string(pdf[:5]))
	}

	wantPattern := regexp.MustCompile(`^kartu-peserta-[A-Z0-9]{8}\.pdf$`)
	if !wantPattern.MatchString(filename) {
		t.Errorf("filename %q does not match kartu-peserta-<8-char-token>.pdf", filename)
	}
}

// --- Rich-text question body sanitization (FR-1..FR-7) ---

func TestSanitizeQuestionBody_stripsScriptTag(t *testing.T) {
	got := sanitizeQuestionBody(`<script>alert(1)</script>Hello`)
	if strings.Contains(got, "<script>") {
		t.Errorf("sanitized body must not contain <script>, got %q", got)
	}
	if !strings.Contains(got, "Hello") {
		t.Errorf("sanitized body should preserve plain text, got %q", got)
	}
}

func TestSanitizeQuestionBody_stripsOnErrorAttr(t *testing.T) {
	got := sanitizeQuestionBody(`<img src=x onerror="alert(1)">`)
	if strings.Contains(strings.ToLower(got), "onerror") {
		t.Errorf("sanitized body must not contain onerror attribute, got %q", got)
	}
	if !strings.Contains(got, "<img") {
		t.Errorf("sanitized body should keep a safe <img> tag, got %q", got)
	}
	if !strings.Contains(got, "src=\"x\"") {
		t.Errorf("sanitized body should keep src=\"x\", got %q", got)
	}
}

func TestSanitizeQuestionBody_stripsPositionFromStyle(t *testing.T) {
	got := sanitizeQuestionBody(`<img src="a" style="position:fixed;top:0">`)
	lower := strings.ToLower(got)
	if strings.Contains(lower, "position") {
		t.Errorf("sanitized style must not contain 'position', got %q", got)
	}
}

func TestSanitizeQuestionBody_preservesAllowlistedTags(t *testing.T) {
	in := `<b>bold</b> <i>italic</i> <u>under</u> <sup>2</sup> <sub>i</sub>`
	got := sanitizeQuestionBody(in)
	if got != in {
		t.Errorf("allowlisted tags must round-trip unchanged\n in: %q\nout: %q", in, got)
	}
}

func TestSanitizeQuestionBody_plainTextRoundTrip(t *testing.T) {
	in := "what is 2 + 2?"
	got := sanitizeQuestionBody(in)
	if got != in {
		t.Errorf("plain text body must round-trip byte-for-byte\n in: %q\nout: %q", in, got)
	}
}

func TestSanitizeQuestionBody_preservesListTags(t *testing.T) {
	in := `<ul><li>one</li><li>two</li></ul>`
	got := sanitizeQuestionBody(in)
	if got != in {
		t.Errorf("list tags must round-trip unchanged\n in: %q\nout: %q", in, got)
	}
}

// --- Question option text sanitization (FR-14) ---

func TestCreateBankQuestion_sanitizes_option_text(t *testing.T) {
	svc, _ := newRealDBService(t)
	ctx := context.Background()

	// Test that malicious option text is sanitized at persist time
	body := "option sanitize mcq " + uniqueSuffix()
	q := model.Question{
		Format:       "mcq",
		Body:         body,
		PointCorrect: 1,
		PointWrong:   0,
	}
	opts := []model.QuestionOption{
		{Key: "a", Text: "<script>alert(1)</script>ok", IsCorrect: true, SortOrder: 1},
		{Key: "b", Text: "<b>bold</b> text", IsCorrect: false, SortOrder: 2},
	}

	_, err := svc.CreateBankQuestion(ctx, q, opts)
	require.NoError(t, err)

	// Fetch back the created question with options via ListBankQuestions
	items, _, err := svc.ListBankQuestions(ctx, repository.QuestionFilter{Search: body, Limit: 10})
	require.NoError(t, err)
	require.Len(t, items, 1)
	fetched := items[0]

	// Verify first option: malicious script removed, text preserved
	require.Len(t, fetched.Options, 2)
	if strings.Contains(fetched.Options[0].Text, "<script>") {
		t.Errorf("option text must not contain <script>, got %q", fetched.Options[0].Text)
	}
	if !strings.Contains(fetched.Options[0].Text, "ok") {
		t.Errorf("option text must preserve plain text, got %q", fetched.Options[0].Text)
	}

	// Verify second option: rich text preserved
	if fetched.Options[1].Text != "<b>bold</b> text" {
		t.Errorf("option text must preserve allowed tags\n in: %q\nout: %q", "<b>bold</b> text", fetched.Options[1].Text)
	}
}

func TestSaveQuestion_sanitizes_option_text(t *testing.T) {
	svc, _ := newRealDBService(t)
	ctx := context.Background()

	// Create a question first
	body := "save question sanitize " + uniqueSuffix()
	q := model.Question{
		Format:       "mcq",
		Body:         body,
		PointCorrect: 1,
		PointWrong:   0,
	}
	opts := []model.QuestionOption{
		{Key: "a", Text: "<script>alert(1)</script>safe", IsCorrect: true, SortOrder: 1},
		{Key: "b", Text: "no", IsCorrect: false, SortOrder: 2},
	}

	out, err := svc.CreateBankQuestion(ctx, q, opts)
	require.NoError(t, err)
	qid := out.Question.ID

	// Update the question with new malicious option text
	updatedBody := "save question updated " + uniqueSuffix()
	updatedOpts := []model.QuestionOption{
		{Key: "a", Text: "<img src=x onerror=\"alert(1)\">updated", IsCorrect: true, SortOrder: 1},
		{Key: "b", Text: "no update", IsCorrect: false, SortOrder: 2},
	}
	q.ID = qid
	q.Body = updatedBody

	_, err = svc.SaveQuestion(ctx, q, updatedOpts)
	require.NoError(t, err)

	// Verify sanitization happened at persist time via ListBankQuestions
	items, _, err := svc.ListBankQuestions(ctx, repository.QuestionFilter{Search: updatedBody, Limit: 10})
	require.NoError(t, err)
	require.Len(t, items, 1)
	fetched := items[0]

	if strings.Contains(strings.ToLower(fetched.Options[0].Text), "onerror") {
		t.Errorf("option text must not contain onerror attribute, got %q", fetched.Options[0].Text)
	}
	if !strings.Contains(fetched.Options[0].Text, "updated") {
		t.Errorf("option text must preserve plain text, got %q", fetched.Options[0].Text)
	}
}

func TestCreateQuestionForTest_sanitizes_option_text(t *testing.T) {
	svc, repo := newRealDBService(t)
	ctx := context.Background()

	// Create a test first
	testID := seedTestDirect(t, ctx, repo, "test for question "+uniqueSuffix(), "math", "algebra")

	body := "question for test sanitize " + uniqueSuffix()
	q := model.Question{
		Format:       "mcq",
		Body:         body,
		PointCorrect: 1,
		PointWrong:   0,
	}
	opts := []model.QuestionOption{
		{Key: "a", Text: "<script>alert(1)</script>answer", IsCorrect: true, SortOrder: 1},
		{Key: "b", Text: "plain answer", IsCorrect: false, SortOrder: 2},
	}

	_, err := svc.CreateQuestionForTest(ctx, testID, q, opts)
	require.NoError(t, err)

	// Verify option text was sanitized via ListBankQuestions
	items, _, err := svc.ListBankQuestions(ctx, repository.QuestionFilter{Search: body, Limit: 10})
	require.NoError(t, err)
	require.Len(t, items, 1)
	fetched := items[0]

	if strings.Contains(fetched.Options[0].Text, "<script>") {
		t.Errorf("option text must not contain <script>, got %q", fetched.Options[0].Text)
	}
	if !strings.Contains(fetched.Options[0].Text, "answer") {
		t.Errorf("option text must preserve plain text, got %q", fetched.Options[0].Text)
	}
}

func TestProcessQuestionImportRows_sanitizes_option_text(t *testing.T) {
	svc, repo := newRealDBService(t)
	ctx := context.Background()

	// Create a topic first for import
	subject := "Math"
	topicName := "Algebra " + uniqueSuffix()
	seedTopicDirect(t, ctx, repo, topicName, subject)

	// Create an import row with malicious option text
	rows := []QuestionImportRow{
		{
			Subject:      subject,
			Topic:        topicName,
			Format:       "mcq",
			Body:         "What is 2+2? " + uniqueSuffix(),
			PointCorrect: 1,
			PointWrong:   0,
			Options: []model.QuestionOption{
				{Key: "a", Text: "<script>alert(1)</script>4", IsCorrect: true, SortOrder: 1},
				{Key: "b", Text: "5", IsCorrect: false, SortOrder: 2},
			},
		},
	}

	result, err := svc.ProcessQuestionImportRows(ctx, rows)
	require.NoError(t, err)
	require.Len(t, result.Rows, 1)
	require.Equal(t, "inserted", result.Rows[0].Status)
	require.NotNil(t, result.Rows[0].QuestionID)

	// Verify option text was sanitized via ListBankQuestions
	body := rows[0].Body
	items, _, err := svc.ListBankQuestions(ctx, repository.QuestionFilter{Search: body, Limit: 10})
	require.NoError(t, err)
	require.Len(t, items, 1)
	fetched := items[0]

	if strings.Contains(fetched.Options[0].Text, "<script>") {
		t.Errorf("option text must not contain <script>, got %q", fetched.Options[0].Text)
	}
	if !strings.Contains(fetched.Options[0].Text, "4") {
		t.Errorf("option text must preserve plain text, got %q", fetched.Options[0].Text)
	}
}