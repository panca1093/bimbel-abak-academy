package model

import (
	"time"

	"github.com/google/uuid"
)

// Test is the top-level authoring unit (a set of questions). Nullable audio fields
// are pointer types so we can persist / return "not set" distinctly from empty strings.
type Test struct {
	ID              uuid.UUID `json:"id"`
	Title           string    `json:"title"`
	Subject         string    `json:"subject"`
	Topic           string    `json:"topic"`
	DurationMinutes int       `json:"duration_minutes"`
	AudioURL        *string   `json:"audio_url"`
	AudioPlayLimit  *int      `json:"audio_play_limit"`
	// QuestionCount is only populated by list-style reads (e.g. ListTests LEFT JOIN).
	// It is zero on freshly created tests and on direct GetByID reads.
	QuestionCount int       `json:"question_count"`
	CreatedAt     time.Time `json:"created_at"`
}

// Question belongs to a Test. `Format` is one of: mcq, multi_answer, short, fill_blank, essay.
// Options are stored separately on QuestionOption (composite PK) and surfaced via
// QuestionWithOptions for read paths.
type Question struct {
	ID            uuid.UUID `json:"id"`
	TestID        uuid.UUID `json:"test_id"`
	Format        string    `json:"format"`
	Body          string    `json:"body"`
	CorrectAnswer *string   `json:"correct_answer"`
	Explanation   *string   `json:"explanation"`
	Difficulty    *string   `json:"difficulty"`
	ImageURL      *string   `json:"image_url"`
	SortOrder     int       `json:"sort_order"`
}

// QuestionOption has a composite PK (QuestionID, Key); no surrogate ID. `Key` is the
// letter shown to candidates (a, b, c, d…). `IsCorrect` is server-validated per format.
type QuestionOption struct {
	QuestionID uuid.UUID `json:"question_id"`
	Key        string    `json:"key"`
	Text       string    `json:"text"`
	ImageURL   *string   `json:"image_url"`
	IsCorrect  bool      `json:"is_correct"`
	SortOrder  int       `json:"sort_order"`
}

// Exam is a scheduled test offering. It bundles one or more Tests via ExamTest and may
// be sold via product (Exam.ProductID, unique when set).
type Exam struct {
	ID                   uuid.UUID  `json:"id"`
	Title                string     `json:"title"`
	IsFree               bool       `json:"is_free"`
	ScheduledAt          *time.Time `json:"scheduled_at"`
	RequiresCheckin      bool       `json:"requires_checkin"`
	AllowLeaderboard     bool       `json:"allow_leaderboard"`
	CDNBundle            bool       `json:"cdn_bundle"`
	BundleURL            *string    `json:"bundle_url"`
	BundleGeneratedAt    *time.Time `json:"bundle_generated_at"`
	CheckInWindowMinutes *int       `json:"check_in_window_minutes"`
	GraceWindowMinutes   *int       `json:"grace_window_minutes"`
	MaxAttempts          *int       `json:"max_attempts"`
	TimerMode            string     `json:"timer_mode"`
	DurationMinutes      *int       `json:"duration_minutes"`
	Randomize            bool       `json:"randomize"`
	ResultConfig         string     `json:"result_config"`
	ResultReleaseAt      *time.Time `json:"result_release_at"`
	Status               string     `json:"status"`
	ProductID            *uuid.UUID `json:"product_id"`
	CreatedAt            time.Time  `json:"created_at"`
}

// ExamTest is the M:N join between Exam and Test with sort order.
type ExamTest struct {
	ID        uuid.UUID `json:"id"`
	ExamID    uuid.UUID `json:"exam_id"`
	TestID    uuid.UUID `json:"test_id"`
	SortOrder int       `json:"sort_order"`
}

// ExamRegistration enrolls a student in an exam; (student_id, exam_id) is unique.
type ExamRegistration struct {
	ID           uuid.UUID  `json:"id"`
	StudentID    uuid.UUID  `json:"student_id"`
	ExamID       uuid.UUID  `json:"exam_id"`
	Token        string     `json:"token"`
	CardPDFURL   *string    `json:"card_pdf_url"`
	CheckedInAt  *time.Time `json:"checked_in_at"`
	AttemptsUsed int        `json:"attempts_used"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
}

// ExamSession is one in-flight attempt by a student; multiple sessions per registration
// are numbered by AttemptNumber.
type ExamSession struct {
	ID             uuid.UUID  `json:"id"`
	RegistrationID uuid.UUID  `json:"registration_id"`
	StudentID      uuid.UUID  `json:"student_id"`
	ExamID         uuid.UUID  `json:"exam_id"`
	AttemptNumber  int        `json:"attempt_number"`
	StartedAt      time.Time  `json:"started_at"`
	SubmittedAt    *time.Time `json:"submitted_at"`
	ExtendedUntil  *time.Time `json:"extended_until"`
	AdminSubmitted bool       `json:"admin_submitted"`
	Score          *float64   `json:"score"`
	CertificateURL *string    `json:"certificate_url"`
	LastSavedAt    *time.Time `json:"last_saved_at"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
}

// ExamSessionAnswer is one answer record per (session, question) — composite PK.
// Score is NUMERIC(6,2) → float64; nullable because essay answers are graded later.
type ExamSessionAnswer struct {
	SessionID        uuid.UUID  `json:"session_id"`
	QuestionID       uuid.UUID  `json:"question_id"`
	Answer           *string    `json:"answer"`
	IsCorrect        *bool      `json:"is_correct"`
	Score            *float64   `json:"score"`
	GradedBy         *uuid.UUID `json:"graded_by"`
	GradedAt         *time.Time `json:"graded_at"`
	GraderComment    *string    `json:"grader_comment"`
	FlaggedForReview bool       `json:"flagged_for_review"`
	SavedAt          time.Time  `json:"saved_at"`
}

// SessionViolationLog records integrity events (tab-switch, copy-paste, etc.) for a session.
type SessionViolationLog struct {
	ID            uuid.UUID `json:"id"`
	SessionID     uuid.UUID `json:"session_id"`
	StudentID     uuid.UUID `json:"student_id"`
	ViolationType string    `json:"violation_type"`
	OccurredAt    time.Time `json:"occurred_at"`
}

// TestDetail is a composite read shape used by the authoring API for a single test page:
// the parent Test and its full ordered question list with inline options.
type TestDetail struct {
	Test      Test                  `json:"test"`
	Questions []QuestionWithOptions `json:"questions"`
}

// QuestionWithOptions is a composite read shape: a Question plus its inline option list.
// Options are empty for non-option formats (short / fill_blank / essay).
type QuestionWithOptions struct {
	Question Question         `json:"question"`
	Options  []QuestionOption `json:"options"`
}

// ExamListItem is the read shape returned by GET /admin/exams — an Exam row joined
// with product.price and product.status. Cursor pagination assembles a slice of these.
type ExamListItem struct {
	Exam          `json:",inline"`
	ProductPrice  int64  `json:"product_price"`
	ProductStatus string `json:"product_status"`
}

// ExamTestEntry is the read shape for an exam_test row plus the inline Test metadata
// (title, subject, topic, duration_minutes, question_count) needed by the admin detail
// page without a second round-trip.
type ExamTestEntry struct {
	ExamTest `json:",inline"`
	Test     struct {
		ID              uuid.UUID `json:"id"`
		Title           string    `json:"title"`
		Subject         string    `json:"subject"`
		Topic           *string   `json:"topic"`
		DurationMinutes *int      `json:"duration_minutes"`
		QuestionCount   int       `json:"question_count"`
	} `json:"test"`
}

// ExamDetail is the read shape returned by GET /admin/exams/:id — full Exam config
// joined with product price/status and an ordered list of attached tests.
type ExamDetail struct {
	Exam          `json:",inline"`
	ProductPrice  int64            `json:"product_price"`
	ProductStatus string           `json:"product_status"`
	Tests         []ExamTestEntry  `json:"tests"`
}

// RegistrationListItem is the read shape returned by GET /api/v1/exam/registrations:
// an ExamRegistration joined with exam.title and exam.scheduled_at.
type RegistrationListItem struct {
	ExamRegistration `json:",inline"`
	ExamTitle        string     `json:"exam_title"`
	ScheduledAt      *time.Time `json:"scheduled_at"`
}

// RegistrationDetail is the read shape returned by GET /api/v1/exam/registrations/:id:
// an ExamRegistration joined with the nested exam config needed by the student detail page.
type RegistrationDetail struct {
	ExamRegistration `json:",inline"`
	Exam             struct {
		ID                   uuid.UUID  `json:"id"`
		Title                string     `json:"title"`
		ScheduledAt          *time.Time `json:"scheduled_at"`
		RequiresCheckin      bool       `json:"requires_checkin"`
		CheckInWindowMinutes *int       `json:"check_in_window_minutes"`
		TimerMode            string     `json:"timer_mode"`
		DurationMinutes      *int       `json:"duration_minutes"`
		ResultConfig         string     `json:"result_config"`
	} `json:"exam"`
}
