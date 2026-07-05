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
	if !strings.Contains(err.Error(), "timer_mode must be overall or per_question") {
		t.Errorf("invalid timer_mode msg should mention 'timer_mode must be overall or per_question', got %q", err.Error())
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

func TestValidateExam_accepts_valid_per_question(t *testing.T) {
	e := model.Exam{Title: "Finals", TimerMode: "per_question"}
	if err := validateExam(e); err != nil {
		t.Errorf("valid per_question should pass, got %v", err)
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

// suppress unused: uuid is imported to avoid unused-import lint if tests get trimmed later
var _ = uuid.Nil

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