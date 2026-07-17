package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/microcosm-cc/bluemonday"

	"akademi-bimbel/internal/model"
	"akademi-bimbel/internal/repository"
)

var (
	ErrTestNotFound          = errors.New("test not found")
	ErrQuestionNotFound      = errors.New("question not found")
	ErrExamNotFound          = errors.New("exam not found")
	ErrRegistrationNotFound  = errors.New("registration not found")
	ErrValidation            = errors.New("validation failed")
)

var validQuestionFormats = map[string]bool{
	"mcq":          true,
	"multi_answer": true,
	"short":        true,
	"fill_blank":   true,
	"multi_blank":  true,
	"essay":        true,
}

// validModes enumerates exam.mode values (FR-1). Empty is allowed here so that
// PATCH-overlay (absent preserves) and CreateExam defaulting both work — the
// service layer sets Mode='standard' on create before validateExam runs, and
// the PATCH handler only overlays Mode when the request supplies a non-empty one.
var validModes = map[string]bool{
	"standard": true,
	"utbk":     true,
	"ielts":    true,
}

// validSectionTypes enumerates test.section_type values (FR-2). NULL is allowed
// (standard/utbk tests may be untyped); only a supplied non-null value is checked.
var validSectionTypes = map[string]bool{
	"listening": true,
	"reading":   true,
	"writing":   true,
}

// sectionedModes are the exam modes that run attached Tests as sequential sections
// (FR-19 publish-gate trigger). 'standard' is excluded.
func isSectionedMode(mode string) bool { return mode == "utbk" || mode == "ielts" }

// questionBodyPolicy is the single allowlist used by sanitizeQuestionBody. It is
// built once at package init and used on every write path. See sanitizeQuestionBody
// for the rationale behind each element/attribute.
var questionBodyPolicy = func() *bluemonday.Policy {
	p := bluemonday.NewPolicy()
	p.AllowElements("b", "i", "u", "ul", "ol", "li", "sup", "sub")
	p.AllowAttrs("src", "alt").OnElements("img")
	p.AllowElements("img")
	// Restricted style: only a safe subset is allowed, and "position" is
	// explicitly rejected by handler (covers "position:fixed" and any
	// other value). url() in style is rejected wholesale — images carry
	// their URL via the src attribute, not in style.
	p.AllowStyles("color", "background-color", "text-align", "font-weight", "font-style", "text-decoration").
		MatchingHandler(func(s string) bool { return !strings.Contains(s, "url(") }).
		OnElements("img")
	return p
}()

// sanitizeQuestionBody strips disallowed HTML from a question body at the
// service trust boundary. The same call is invoked at every write path
// (CreateBankQuestion, SaveQuestion, CreateQuestionForTest,
// ProcessQuestionImportRows) so the persisted value is the sanitized one.
//
// Allowlist: b, i, u, ul, ol, li, sup, sub, img (with src/alt and a restricted
// style attribute). On* handlers, <script>, <iframe>, javascript: URLs, and
// any non-allowlisted tag are stripped. "position" in style is rejected via
// the regex below; url() in style is rejected via the policy's value handler.
func sanitizeQuestionBody(body string) string {
	if body == "" {
		return body
	}
	// Reject "position" declarations regardless of value (fixed/absolute/...).
	// bluemonday processes the value handler first; this regexp is a
	// belt-and-suspenders guard for any value that doesn't trigger a
	// direct string check (e.g. "POSITION:fixed" via casing).
	positionDecl := regexp.MustCompile(`(?i)position\s*:`)
	if positionDecl.MatchString(body) {
		body = positionDecl.ReplaceAllString(body, "")
	}
	cleaned := questionBodyPolicy.Sanitize(body)
	// Collapse runs of ";" left behind by stripping a declaration.
	cleaned = regexp.MustCompile(`;\s*;`).ReplaceAllString(cleaned, ";")
	cleaned = strings.TrimSpace(cleaned)
	return cleaned
}

// sanitizeQuestionOptions sanitizes the Text field of each option using the same
// policy as sanitizeQuestionBody. Called at every write path (CreateBankQuestion,
// SaveQuestion, CreateQuestionForTest, ProcessQuestionImportRows) so the persisted
// value is the sanitized one.
func sanitizeQuestionOptions(options []model.QuestionOption) []model.QuestionOption {
	sanitized := make([]model.QuestionOption, len(options))
	for i, opt := range options {
		sanitized[i] = opt
		sanitized[i].Text = sanitizeQuestionBody(opt.Text)
	}
	return sanitized
}

// extractTokensFromStem finds all {{N}} tokens in the stem and returns their
// numeric indices in order. Returns the slice and error if any token is malformed.
func extractTokensFromStem(body string) ([]int, error) {
	tokenPattern := regexp.MustCompile(`\{\{(\d+)\}\}`)
	matches := tokenPattern.FindAllStringSubmatch(body, -1)
	if len(matches) == 0 {
		return []int{}, nil
	}

	tokens := make([]int, len(matches))
	for i, match := range matches {
		// match[1] is the captured group (the number inside {{N}})
		num := 0
		_, err := fmt.Sscanf(match[1], "%d", &num)
		if err != nil {
			return nil, fmt.Errorf("%w: malformed token number", ErrValidation)
		}
		tokens[i] = num
	}
	return tokens, nil
}

// validateTokenSequence checks that the given token indices form an exact 1..N
// contiguous set with no duplicates or gaps.
func validateTokenSequence(tokens []int) error {
	if len(tokens) == 0 {
		return fmt.Errorf("%w: multi_blank requires at least one {{N}} token", ErrValidation)
	}

	seen := make(map[int]bool)
	for _, t := range tokens {
		if t < 1 {
			return fmt.Errorf("%w: token index must be >= 1", ErrValidation)
		}
		if seen[t] {
			return fmt.Errorf("%w: duplicate token index: %d", ErrValidation, t)
		}
		seen[t] = true
	}

	for i := 1; i <= len(tokens); i++ {
		if !seen[i] {
			return fmt.Errorf("%w: token sequence has gap; expected {{%d}}", ErrValidation, i)
		}
	}

	return nil
}

// validateQuestion enforces the format-validation matrix from spec.md §4.
// All error returns wrap ErrValidation with a sub-message so callers can
// use errors.Is(err, ErrValidation) AND err.Error() carries the WHY.
func validateQuestion(q model.Question, options []model.QuestionOption, blanks []model.QuestionBlank) error {
	if !validQuestionFormats[q.Format] {
		return fmt.Errorf("%w: unknown question format: %s", ErrValidation, q.Format)
	}

	// Callers sanitize q.Body via sanitizeQuestionBody before calling here, which
	// strips non-allowlisted tags (e.g. <br>) with no text content down to "".
	// Reject that post-sanitize emptiness so a blank question can't be saved.
	if strings.TrimSpace(q.Body) == "" {
		return fmt.Errorf("%w: body cannot be empty", ErrValidation)
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
	case "multi_blank":
		// multi_blank: stem contains {{N}} tokens, no options, no scalar correct_answer,
		// blanks array has one entry per token with non-empty correct_answer
		if hasOptions {
			return fmt.Errorf("%w: multi_blank cannot have options", ErrValidation)
		}
		if hasCorrectAnswer {
			return fmt.Errorf("%w: multi_blank cannot have correct_answer", ErrValidation)
		}

		tokens, err := extractTokensFromStem(q.Body)
		if err != nil {
			return err
		}
		if err := validateTokenSequence(tokens); err != nil {
			return err
		}

		if len(blanks) != len(tokens) {
			return fmt.Errorf("%w: blanks count (%d) must match token count (%d)", ErrValidation, len(blanks), len(tokens))
		}

		for i, blank := range blanks {
			if strings.TrimSpace(blank.CorrectAnswer) == "" {
				return fmt.Errorf("%w: blank at index %d has empty correct_answer", ErrValidation, i)
			}
		}
	case "essay":
		if hasOptions {
			return fmt.Errorf("%w: essay cannot have options", ErrValidation)
		}
		if hasCorrectAnswer {
			return fmt.Errorf("%w: essay cannot have correct_answer", ErrValidation)
		}
	}

	// FR-S5-02: point_correct/point_wrong are unsigned magnitudes; the scoring
	// engine (not the author) applies the sign for wrong answers.
	if q.PointCorrect < 1 {
		return fmt.Errorf("%w: point_correct must be >= 1", ErrValidation)
	}
	if q.PointWrong < 0 {
		return fmt.Errorf("%w: point_wrong must be >= 0", ErrValidation)
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
	// FR-18: section_type ∈ {listening,reading,writing} when supplied; a listening
	// section requires audio_url (FR-25 audio player). NULL section_type is allowed
	// (standard/utbk tests may be untyped).
	if t.SectionType != nil {
		if !validSectionTypes[*t.SectionType] {
			return fmt.Errorf("%w: section_type must be listening, reading, or writing", ErrValidation)
		}
		if *t.SectionType == "listening" {
			if t.AudioURL == nil || strings.TrimSpace(*t.AudioURL) == "" {
				return fmt.Errorf("%w: audio_url required when section_type=listening", ErrValidation)
			}
		}
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
func (s *Service) SaveQuestion(ctx context.Context, q model.Question, options []model.QuestionOption, blanks []model.QuestionBlank) (model.QuestionWithOptions, error) {
	q.Body = sanitizeQuestionBody(q.Body)
	options = sanitizeQuestionOptions(options)
	if err := validateQuestion(q, options, blanks); err != nil {
		return model.QuestionWithOptions{}, err
	}

	tx, err := s.storeRepo.BeginTx(ctx)
	if err != nil {
		return model.QuestionWithOptions{}, err
	}
	defer tx.Rollback(ctx)

	if q.ID == uuid.Nil {
		if err := s.storeRepo.CreateQuestionTx(ctx, tx, &q, options, blanks); err != nil {
			return model.QuestionWithOptions{}, err
		}
	} else {
		if err := s.storeRepo.UpdateQuestionTx(ctx, tx, &q, options, blanks); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return model.QuestionWithOptions{}, ErrQuestionNotFound
			}
			return model.QuestionWithOptions{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return model.QuestionWithOptions{}, err
	}

	return model.QuestionWithOptions{Question: q, Options: options, Blanks: blanks}, nil
}

// CreateQuestionForTest creates a bank question and atomically attaches it to the
// given test (FR-25). This preserves the existing POST /admin/tests/:id/questions
// behavior after migration 0025 moved attachment to test_question.
func (s *Service) CreateQuestionForTest(ctx context.Context, testID uuid.UUID, q model.Question, options []model.QuestionOption, blanks []model.QuestionBlank) (model.QuestionWithOptions, error) {
	q.Body = sanitizeQuestionBody(q.Body)
	options = sanitizeQuestionOptions(options)
	if err := validateQuestion(q, options, blanks); err != nil {
		return model.QuestionWithOptions{}, err
	}

	tx, err := s.storeRepo.BeginTx(ctx)
	if err != nil {
		return model.QuestionWithOptions{}, err
	}
	defer tx.Rollback(ctx)

	if err := s.storeRepo.CreateQuestionTx(ctx, tx, &q, options, blanks); err != nil {
		return model.QuestionWithOptions{}, err
	}
	if err := s.storeRepo.AttachQuestionToTestTx(ctx, tx, testID, q.ID); err != nil {
		return model.QuestionWithOptions{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return model.QuestionWithOptions{}, err
	}

	return model.QuestionWithOptions{Question: q, Options: options, Blanks: blanks}, nil
}

// AttachQuestions appends bank questions to a test. Already-attached questions are
// idempotent (no duplicate, no error). Every supplied question must exist; the
// test must exist (FR-21).
func (s *Service) AttachQuestions(ctx context.Context, testID uuid.UUID, questionIDs []uuid.UUID) error {
	if len(questionIDs) == 0 {
		return fmt.Errorf("%w: question_ids cannot be empty", ErrValidation)
	}
	if _, err := s.storeRepo.GetTestByID(ctx, testID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrTestNotFound
		}
		return err
	}

	exists, err := s.storeRepo.CountQuestionsByIDs(ctx, questionIDs)
	if err != nil {
		return err
	}
	if exists != len(questionIDs) {
		return ErrQuestionNotFound
	}

	colliding, err := s.storeRepo.FindQuestionsAttachedToSiblingTests(ctx, testID, questionIDs)
	if err != nil {
		return err
	}
	if len(colliding) > 0 {
		return fmt.Errorf("%w: question(s) already attached to another test in the same exam: %v", ErrValidation, colliding)
	}

	tx, err := s.storeRepo.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := s.storeRepo.AttachQuestionsToTestTx(ctx, tx, testID, questionIDs); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// DetachQuestion removes only the test_question join row; the bank question and any
// other attachments survive (FR-22).
func (s *Service) DetachQuestion(ctx context.Context, testID, questionID uuid.UUID) error {
	if _, err := s.storeRepo.GetTestByID(ctx, testID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrTestNotFound
		}
		return err
	}
	return s.storeRepo.DetachQuestionFromTest(ctx, testID, questionID)
}

// ReorderTestQuestions validates that the supplied id set equals the currently
// attached set, then atomically rewrites sort_order to match the list position
// (FR-23).
func (s *Service) ReorderTestQuestions(ctx context.Context, testID uuid.UUID, orderedQuestionIDs []uuid.UUID) error {
	if len(orderedQuestionIDs) == 0 {
		return fmt.Errorf("%w: question_ids cannot be empty", ErrValidation)
	}
	if _, err := s.storeRepo.GetTestByID(ctx, testID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrTestNotFound
		}
		return err
	}

	attached, err := s.storeRepo.ListAttachedQuestionIDs(ctx, testID)
	if err != nil {
		return err
	}
	if !sameUUIDSet(attached, orderedQuestionIDs) {
		return fmt.Errorf("%w: question_ids must match the current attached set", ErrValidation)
	}

	tx, err := s.storeRepo.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := s.storeRepo.ReorderTestQuestionsTx(ctx, tx, testID, orderedQuestionIDs); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func sameUUIDSet(a, b []uuid.UUID) bool {
	if len(a) != len(b) {
		return false
	}
	m := make(map[uuid.UUID]bool, len(a))
	for _, v := range a {
		m[v] = true
	}
	seen := make(map[uuid.UUID]bool, len(b))
	for _, v := range b {
		if !m[v] || seen[v] {
			return false
		}
		seen[v] = true
	}
	return true
}

// CreateBankQuestion creates a question in the bank with no test attachment (FR-9).
func (s *Service) CreateBankQuestion(ctx context.Context, q model.Question, options []model.QuestionOption, blanks []model.QuestionBlank) (model.QuestionWithOptions, error) {
	q.Body = sanitizeQuestionBody(q.Body)
	options = sanitizeQuestionOptions(options)
	if err := validateQuestion(q, options, blanks); err != nil {
		return model.QuestionWithOptions{}, err
	}

	tx, err := s.storeRepo.BeginTx(ctx)
	if err != nil {
		return model.QuestionWithOptions{}, err
	}
	defer tx.Rollback(ctx)

	if err := s.storeRepo.CreateQuestionTx(ctx, tx, &q, options, blanks); err != nil {
		return model.QuestionWithOptions{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return model.QuestionWithOptions{}, err
	}

	return model.QuestionWithOptions{Question: q, Options: options, Blanks: blanks}, nil
}

func (s *Service) DeleteQuestion(ctx context.Context, id uuid.UUID) error {
	attached, err := s.storeRepo.CountQuestionAttachments(ctx, id)
	if err != nil {
		return err
	}
	if attached > 0 {
		return fmt.Errorf("%w: question is attached to %d test(s); detach before deleting", ErrValidation, attached)
	}

	answered, err := s.storeRepo.CountAnswerReferences(ctx, id)
	if err != nil {
		return err
	}
	if answered > 0 {
		return fmt.Errorf("%w: question has been answered in %d session(s) and cannot be deleted", ErrValidation, answered)
	}

	if err := s.storeRepo.DeleteQuestion(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrQuestionNotFound
		}
		return err
	}
	return nil
}

// ListBankQuestions returns cursor-paginated bank questions with topic name and
// attached-test count (FR-14).
func (s *Service) ListBankQuestions(ctx context.Context, filter repository.QuestionFilter) ([]model.BankQuestionListItem, string, error) {
	return s.storeRepo.ListBankQuestions(ctx, filter)
}

var validTimerModes = map[string]bool{
	"overall":      true,
	"per_test": true,
}

var validResultConfigs = map[string]bool{
	"hidden":           true,
	"score_only":       true,
	"score_pembahasan": true,
}

// validateExam enforces exam-level invariants: title required, timer_mode
// ∈ {overall, per_test} (empty allowed for legacy rows), duration
// required when timer_mode=overall, and result_config ∈ {hidden, score_only,
// score_pembahasan} when set (empty allowed here — CreateExam defaults it
// before this runs; the DB CHECK constraint added by migration 0015 rejects
// empty string, so this must never reach validateExam still empty on create).
func validateExam(e model.Exam) error {
	if strings.TrimSpace(e.Title) == "" {
		return fmt.Errorf("%w: exam title required", ErrValidation)
	}
	// FR-18: mode ∈ {standard,utbk,ielts} when supplied. Empty is accepted here
	// because CreateExam defaults it to 'standard' before validateExam runs and
	// the PATCH handler only overlays Mode when the request supplies a non-empty
	// value (absent preserves the stored value).
	if e.Mode != "" && !validModes[e.Mode] {
		return fmt.Errorf("%w: mode must be standard, utbk, or ielts", ErrValidation)
	}
	if e.TimerMode != "" && !validTimerModes[e.TimerMode] {
		return fmt.Errorf("%w: timer_mode must be overall or per_test", ErrValidation)
	}
	if e.TimerMode == "overall" {
		if e.DurationMinutes == nil || *e.DurationMinutes <= 0 {
			return fmt.Errorf("%w: duration_minutes required and positive when timer_mode=overall", ErrValidation)
		}
	}
	if e.ResultConfig != "" && !validResultConfigs[e.ResultConfig] {
		return fmt.Errorf("%w: result_config must be hidden, score_only, or score_pembahasan", ErrValidation)
	}
	if e.CertificateTemplate != "" {
		if err := validateCertificateTemplate(e.CertificateTemplate); err != nil {
			return err
		}
	}
	if e.CheckInWindowMinutes != nil && *e.CheckInWindowMinutes < 0 {
		return fmt.Errorf("%w: check_in_window_minutes cannot be negative", ErrValidation)
	}
	if e.GraceWindowMinutes != nil && *e.GraceWindowMinutes < 0 {
		return fmt.Errorf("%w: grace_window_minutes cannot be negative", ErrValidation)
	}
	if e.MaxAttempts != nil && *e.MaxAttempts < 0 {
		return fmt.Errorf("%w: max_attempts cannot be negative", ErrValidation)
	}
	return nil
}

// CreateExam creates a standalone exam — no product is created here. Selling the
// exam is a separate step: attach it to a Product via the generic Product flow
// (POST/PATCH /admin/products with exam_ids), mirroring how course-type products
// attach existing Courses.
func (s *Service) CreateExam(ctx context.Context, m model.Exam) (model.Exam, error) {
	if m.ResultConfig == "" {
		m.ResultConfig = "hidden"
	}
	if m.CertificateTemplate == "" {
		m.CertificateTemplate = "classic"
	}
	if m.Status == "" {
		m.Status = "draft"
	}
	if m.Mode == "" {
		m.Mode = "standard"
	}
	if err := validateExam(m); err != nil {
		return model.Exam{}, err
	}
	if err := s.storeRepo.CreateExam(ctx, &m); err != nil {
		return model.Exam{}, err
	}
	return m, nil
}

func (s *Service) UpdateExam(ctx context.Context, id uuid.UUID, m model.Exam) (model.Exam, error) {
	if err := validateExam(m); err != nil {
		return model.Exam{}, err
	}
	existing, err := s.storeRepo.GetExamByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.Exam{}, ErrExamNotFound
		}
		return model.Exam{}, err
	}
	m.ID = id
	m.CreatedAt = existing.CreatedAt
	if err := s.storeRepo.UpdateExam(ctx, id, &m); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.Exam{}, ErrExamNotFound
		}
		return model.Exam{}, err
	}
	return m, nil
}

func (s *Service) ReplaceExamTests(ctx context.Context, examID uuid.UUID, testIDs []uuid.UUID) error {
	if _, err := s.storeRepo.GetExamByID(ctx, examID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrExamNotFound
		}
		return err
	}

	for _, testID := range testIDs {
		if _, err := s.storeRepo.GetTestByID(ctx, testID); err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return ErrTestNotFound
			}
			return err
		}
	}

	tests := make([]model.ExamTest, len(testIDs))
	for i, testID := range testIDs {
		tests[i] = model.ExamTest{ExamID: examID, TestID: testID, SortOrder: i}
	}

	tx, err := s.storeRepo.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := s.storeRepo.ReplaceExamTestsTx(ctx, tx, examID, tests); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// validatePublishSections enforces FR-19: a sectioned exam (mode ∈ {utbk,ielts})
// must have at least one attached Test, and an ielts exam must have a non-null
// section_type on every attached Test (the offending section is named in the
// error message). Standard/empty modes skip the gate entirely. The function is
// pure over (exam, tests) so it can be unit-tested without a DB; PublishProduct
// runs this for every exam attached to an exam-type product before publishing.
func validatePublishSections(exam model.Exam, tests []model.ExamTestEntry) error {
	if !isSectionedMode(exam.Mode) {
		return nil
	}
	if len(tests) == 0 {
		return fmt.Errorf("%w: sectioned exam (mode=%s) requires at least one test", ErrValidation, exam.Mode)
	}
	if exam.Mode == "ielts" {
		for _, te := range tests {
			if te.Test.SectionType == nil {
				title := te.Test.Title
				if title == "" {
					title = te.TestID.String()
				}
				return fmt.Errorf("%w: ielts exam has an untyped section (%q); section_type required", ErrValidation, title)
			}
		}
	}
	return nil
}

func (s *Service) GetExam(ctx context.Context, id uuid.UUID) (model.ExamDetail, error) {
	d, err := s.storeRepo.GetExamDetail(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.ExamDetail{}, ErrExamNotFound
		}
		return model.ExamDetail{}, err
	}
	return *d, nil
}

func (s *Service) ListExams(ctx context.Context, filter repository.ExamFilter) ([]model.ExamListItem, string, error) {
	return s.storeRepo.ListExams(ctx, filter)
}

func (s *Service) GetExamRegistrations(ctx context.Context, studentID string) ([]model.RegistrationListItem, error) {
	sid, err := uuid.Parse(studentID)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid student id", ErrValidation)
	}
	return s.storeRepo.GetExamRegistrationsByStudent(ctx, sid)
}

func (s *Service) GetExamRegistration(ctx context.Context, regID, studentID string) (*model.RegistrationDetail, error) {
	rid, err := uuid.Parse(regID)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid registration id", ErrValidation)
	}
	sid, err := uuid.Parse(studentID)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid student id", ErrValidation)
	}
	detail, err := s.storeRepo.GetExamRegistrationByID(ctx, rid, sid)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrRegistrationNotFound
	}
	return detail, err
}

func (s *Service) GetExamCard(ctx context.Context, regID, studentID string) ([]byte, string, error) {
	detail, err := s.GetExamRegistration(ctx, regID, studentID)
	if err != nil {
		return nil, "", err
	}
	studentName := ""
	user, err := s.Me(ctx, studentID)
	if err == nil && user != nil {
		studentName = user.Name
	}
	tenantName := ""
	cfg, err := s.GetSystemConfig(ctx)
	if err == nil && cfg != nil {
		if v, ok := cfg["app_name"]; ok && v != "" {
			tenantName = v
		}
	}
	if tenantName == "" {
		tenantName = "Akademi Bimbel"
	}
	pdf, err := generateExamCardPDF(detail, studentName, tenantName)
	if err != nil {
		return nil, "", err
	}
	return pdf, "kartu-peserta-" + detail.Token + ".pdf", nil
}