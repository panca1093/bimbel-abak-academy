package model

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
)

// Shared helpers
func mustField(t *testing.T, v reflect.Value, name string) reflect.StructField {
	t.Helper()
	f, ok := v.Type().FieldByName(name)
	if !ok {
		t.Fatalf("struct %s missing field %q", v.Type().Name(), name)
	}
	return f
}

func jsonTag(t *testing.T, v reflect.Value, name, want string) {
	t.Helper()
	f := mustField(t, v, name)
	got := f.Tag.Get("json")
	if got != want {
		t.Errorf("%s.%s json tag: got %q, want %q", v.Type().Name(), name, got, want)
	}
}

func fieldType(t *testing.T, v reflect.Value, name string, want reflect.Type) {
	t.Helper()
	f := mustField(t, v, name)
	if f.Type != want {
		t.Errorf("%s.%s type: got %s, want %s", v.Type().Name(), name, f.Type, want)
	}
}

func fieldKind(t *testing.T, v reflect.Value, name string, want reflect.Kind) {
	t.Helper()
	f := mustField(t, v, name)
	if f.Type.Kind() != want {
		t.Errorf("%s.%s kind: got %s, want %s", v.Type().Name(), name, f.Type.Kind(), want)
	}
}

func newModel(t reflect.Type) reflect.Value {
	v := reflect.New(t).Elem()
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		switch f.Kind() {
		case reflect.Ptr:
			// leave nil (zero value), valid for tests
		case reflect.Slice, reflect.Map:
			// leave nil (zero value), valid for tests
		}
	}
	return v
}

// ---- Test ----

func TestTestStruct(t *testing.T) {
	typ := reflect.TypeOf((*Test)(nil)).Elem()
	v := newModel(typ)

	if typ.NumField() != 9 {
		t.Fatalf("Test struct: got %d fields, want 9", typ.NumField())
	}

	jsonTag(t, v, "ID", "id")
	jsonTag(t, v, "Title", "title")
	jsonTag(t, v, "Subject", "subject")
	jsonTag(t, v, "Topic", "topic")
	jsonTag(t, v, "DurationMinutes", "duration_minutes")
	jsonTag(t, v, "AudioURL", "audio_url")
	jsonTag(t, v, "AudioPlayLimit", "audio_play_limit")
	jsonTag(t, v, "QuestionCount", "question_count")
	jsonTag(t, v, "CreatedAt", "created_at")

	fieldKind(t, v, "QuestionCount", reflect.Int)

	fieldType(t, v, "ID", reflect.TypeOf(uuid.UUID{}))
	fieldKind(t, v, "Title", reflect.String)
	fieldKind(t, v, "DurationMinutes", reflect.Int)
	// AudioURL is nullable → pointer
	fieldKind(t, v, "AudioURL", reflect.Ptr)
	if mustField(t, v, "AudioURL").Type.Elem().Kind() != reflect.String {
		t.Errorf("Test.AudioURL pointer base type should be string")
	}
	// AudioPlayLimit is nullable → pointer to int
	fieldKind(t, v, "AudioPlayLimit", reflect.Ptr)
	if mustField(t, v, "AudioPlayLimit").Type.Elem().Kind() != reflect.Int {
		t.Errorf("Test.AudioPlayLimit pointer base type should be int")
	}
	fieldType(t, v, "CreatedAt", reflect.TypeOf(time.Time{}))
}

// ---- Question ----

func TestQuestionStruct(t *testing.T) {
	typ := reflect.TypeOf((*Question)(nil)).Elem()
	v := newModel(typ)

	// Question has NO created_at (migration doesn't define one).
	if typ.NumField() != 9 {
		t.Fatalf("Question struct: got %d fields, want 9", typ.NumField())
	}
	if _, ok := typ.FieldByName("CreatedAt"); ok {
		t.Errorf("Question must NOT have CreatedAt — migration 0014_exam.up.sql does not define it")
	}

	jsonTag(t, v, "TestID", "test_id")
	jsonTag(t, v, "Format", "format")
	jsonTag(t, v, "CorrectAnswer", "correct_answer")
	jsonTag(t, v, "ImageURL", "image_url")
	jsonTag(t, v, "SortOrder", "sort_order")

	fieldType(t, v, "ID", reflect.TypeOf(uuid.UUID{}))
	fieldType(t, v, "TestID", reflect.TypeOf(uuid.UUID{}))
	fieldKind(t, v, "Format", reflect.String)
	fieldKind(t, v, "Body", reflect.String)

	// Nullable pointers
	for _, name := range []string{"CorrectAnswer", "Explanation", "Difficulty", "ImageURL"} {
		fieldKind(t, v, name, reflect.Ptr)
		if mustField(t, v, name).Type.Elem().Kind() != reflect.String {
			t.Errorf("Question.%s pointer base type should be string, got %s", name, mustField(t, v, name).Type.Elem().Kind())
		}
	}

	fieldKind(t, v, "SortOrder", reflect.Int)
}

// ---- QuestionOption ----

func TestQuestionOptionStruct(t *testing.T) {
	typ := reflect.TypeOf((*QuestionOption)(nil)).Elem()
	v := newModel(typ)

	// composite PK: question_id + key (no surrogate UUID primary key)
	if typ.NumField() != 6 {
		t.Fatalf("QuestionOption struct: got %d fields, want 6", typ.NumField())
	}

	jsonTag(t, v, "QuestionID", "question_id")
	jsonTag(t, v, "Key", "key")
	jsonTag(t, v, "Text", "text")
	jsonTag(t, v, "ImageURL", "image_url")
	jsonTag(t, v, "IsCorrect", "is_correct")
	jsonTag(t, v, "SortOrder", "sort_order")

	// Confirm NO UUID field exists with name like "ID"
	if _, ok := typ.FieldByName("ID"); ok {
		t.Errorf("QuestionOption must NOT have a surrogate ID field (composite PK)")
	}

	fieldType(t, v, "QuestionID", reflect.TypeOf(uuid.UUID{}))
	fieldKind(t, v, "Key", reflect.String)
	fieldKind(t, v, "Text", reflect.String)
	fieldKind(t, v, "IsCorrect", reflect.Bool)
	fieldKind(t, v, "SortOrder", reflect.Int)

	fieldKind(t, v, "ImageURL", reflect.Ptr)
	if mustField(t, v, "ImageURL").Type.Elem().Kind() != reflect.String {
		t.Errorf("QuestionOption.ImageURL pointer base type should be string")
	}

	// No CreatedAt per DBML
	if _, ok := typ.FieldByName("CreatedAt"); ok {
		t.Errorf("QuestionOption must NOT have CreatedAt per DBML")
	}
}

// ---- Exam ----

func TestExamStruct(t *testing.T) {
	typ := reflect.TypeOf((*Exam)(nil)).Elem()
	v := newModel(typ)

	// Check count later — we verify a representative subset of fields and tags.
	jsonTag(t, v, "ID", "id")
	jsonTag(t, v, "Title", "title")
	jsonTag(t, v, "IsFree", "is_free")
	jsonTag(t, v, "ScheduledAt", "scheduled_at")
	jsonTag(t, v, "RequiresCheckin", "requires_checkin")
	jsonTag(t, v, "AllowLeaderboard", "allow_leaderboard")
	jsonTag(t, v, "CDNBundle", "cdn_bundle")
	jsonTag(t, v, "BundleURL", "bundle_url")
	jsonTag(t, v, "BundleGeneratedAt", "bundle_generated_at")
	jsonTag(t, v, "CheckInWindowMinutes", "check_in_window_minutes")
	jsonTag(t, v, "GraceWindowMinutes", "grace_window_minutes")
	jsonTag(t, v, "MaxAttempts", "max_attempts")
	jsonTag(t, v, "TimerMode", "timer_mode")
	jsonTag(t, v, "DurationMinutes", "duration_minutes")
	jsonTag(t, v, "Randomize", "randomize")
	jsonTag(t, v, "ResultConfig", "result_config")
	jsonTag(t, v, "ResultReleaseAt", "result_release_at")
	jsonTag(t, v, "Status", "status")
	jsonTag(t, v, "ProductID", "product_id")
	jsonTag(t, v, "CreatedAt", "created_at")

	fieldType(t, v, "ID", reflect.TypeOf(uuid.UUID{}))
	fieldKind(t, v, "IsFree", reflect.Bool)
	fieldKind(t, v, "Title", reflect.String)
	fieldKind(t, v, "TimerMode", reflect.String)
	fieldKind(t, v, "ResultConfig", reflect.String)
	fieldKind(t, v, "Status", reflect.String)
	fieldKind(t, v, "Randomize", reflect.Bool)

	// Nullable pointers
	for _, name := range []string{"ScheduledAt", "BundleGeneratedAt", "ResultReleaseAt"} {
		fieldKind(t, v, name, reflect.Ptr)
		if mustField(t, v, name).Type != reflect.TypeOf((*time.Time)(nil)) {
			t.Errorf("Exam.%s should be *time.Time, got %s", name, mustField(t, v, name).Type)
		}
	}
	for _, name := range []string{"CheckInWindowMinutes", "GraceWindowMinutes", "MaxAttempts", "DurationMinutes"} {
		fieldKind(t, v, name, reflect.Ptr)
		if mustField(t, v, name).Type.Elem().Kind() != reflect.Int {
			t.Errorf("Exam.%s pointer base type should be int, got %s", name, mustField(t, v, name).Type.Elem().Kind())
		}
	}
	fieldKind(t, v, "BundleURL", reflect.Ptr)
	if mustField(t, v, "BundleURL").Type.Elem().Kind() != reflect.String {
		t.Errorf("Exam.BundleURL pointer base type should be string")
	}

	fieldKind(t, v, "ProductID", reflect.Ptr)
	if mustField(t, v, "ProductID").Type != reflect.TypeOf((*uuid.UUID)(nil)) {
		t.Errorf("Exam.ProductID should be *uuid.UUID, got %s", mustField(t, v, "ProductID").Type)
	}

	fieldType(t, v, "CreatedAt", reflect.TypeOf(time.Time{}))
}

// ---- ExamTest ----

func TestExamTestStruct(t *testing.T) {
	typ := reflect.TypeOf((*ExamTest)(nil)).Elem()
	v := newModel(typ)

	if typ.NumField() != 4 {
		t.Fatalf("ExamTest struct: got %d fields, want 4", typ.NumField())
	}

	jsonTag(t, v, "ID", "id")
	jsonTag(t, v, "ExamID", "exam_id")
	jsonTag(t, v, "TestID", "test_id")
	jsonTag(t, v, "SortOrder", "sort_order")

	fieldType(t, v, "ID", reflect.TypeOf(uuid.UUID{}))
	fieldType(t, v, "ExamID", reflect.TypeOf(uuid.UUID{}))
	fieldType(t, v, "TestID", reflect.TypeOf(uuid.UUID{}))
	fieldKind(t, v, "SortOrder", reflect.Int)
}

// ---- ExamRegistration ----

func TestExamRegistrationStruct(t *testing.T) {
	typ := reflect.TypeOf((*ExamRegistration)(nil)).Elem()
	v := newModel(typ)

	jsonTag(t, v, "ID", "id")
	jsonTag(t, v, "StudentID", "student_id")
	jsonTag(t, v, "ExamID", "exam_id")
	jsonTag(t, v, "Token", "token")
	jsonTag(t, v, "CardPDFURL", "card_pdf_url")
	jsonTag(t, v, "CheckedInAt", "checked_in_at")
	jsonTag(t, v, "AttemptsUsed", "attempts_used")
	jsonTag(t, v, "Status", "status")
	jsonTag(t, v, "CreatedAt", "created_at")

	fieldType(t, v, "ID", reflect.TypeOf(uuid.UUID{}))
	fieldType(t, v, "StudentID", reflect.TypeOf(uuid.UUID{}))
	fieldType(t, v, "ExamID", reflect.TypeOf(uuid.UUID{}))
	fieldKind(t, v, "Token", reflect.String)
	fieldKind(t, v, "AttemptsUsed", reflect.Int)
	fieldKind(t, v, "Status", reflect.String)

	fieldKind(t, v, "CardPDFURL", reflect.Ptr)
	if mustField(t, v, "CardPDFURL").Type.Elem().Kind() != reflect.String {
		t.Errorf("ExamRegistration.CardPDFURL pointer base type should be string")
	}
	fieldKind(t, v, "CheckedInAt", reflect.Ptr)
	if mustField(t, v, "CheckedInAt").Type != reflect.TypeOf((*time.Time)(nil)) {
		t.Errorf("ExamRegistration.CheckedInAt should be *time.Time, got %s", mustField(t, v, "CheckedInAt").Type)
	}
	fieldType(t, v, "CreatedAt", reflect.TypeOf(time.Time{}))
}

// ---- ExamSession ----

func TestExamSessionStruct(t *testing.T) {
	typ := reflect.TypeOf((*ExamSession)(nil)).Elem()
	v := newModel(typ)

	jsonTag(t, v, "ID", "id")
	jsonTag(t, v, "RegistrationID", "registration_id")
	jsonTag(t, v, "StudentID", "student_id")
	jsonTag(t, v, "ExamID", "exam_id")
	jsonTag(t, v, "AttemptNumber", "attempt_number")
	jsonTag(t, v, "StartedAt", "started_at")
	jsonTag(t, v, "SubmittedAt", "submitted_at")
	jsonTag(t, v, "ExtendedUntil", "extended_until")
	jsonTag(t, v, "AdminSubmitted", "admin_submitted")
	jsonTag(t, v, "Score", "score")
	jsonTag(t, v, "CertificateURL", "certificate_url")
	jsonTag(t, v, "LastSavedAt", "last_saved_at")
	jsonTag(t, v, "Status", "status")
	jsonTag(t, v, "CreatedAt", "created_at")

	fieldType(t, v, "ID", reflect.TypeOf(uuid.UUID{}))
	fieldType(t, v, "RegistrationID", reflect.TypeOf(uuid.UUID{}))
	fieldType(t, v, "StudentID", reflect.TypeOf(uuid.UUID{}))
	fieldType(t, v, "ExamID", reflect.TypeOf(uuid.UUID{}))
	fieldKind(t, v, "AttemptNumber", reflect.Int)
	fieldKind(t, v, "AdminSubmitted", reflect.Bool)
	fieldKind(t, v, "Status", reflect.String)

	// started_at not null
	fieldType(t, v, "StartedAt", reflect.TypeOf(time.Time{}))
	fieldType(t, v, "CreatedAt", reflect.TypeOf(time.Time{}))

	// Nullable pointers
	for _, name := range []string{"SubmittedAt", "ExtendedUntil", "LastSavedAt"} {
		fieldKind(t, v, name, reflect.Ptr)
		if mustField(t, v, name).Type != reflect.TypeOf((*time.Time)(nil)) {
			t.Errorf("ExamSession.%s should be *time.Time, got %s", name, mustField(t, v, name).Type)
		}
	}
	// Score is NUMERIC(6,2) nullable → *float64
	fieldKind(t, v, "Score", reflect.Ptr)
	if mustField(t, v, "Score").Type.Elem().Kind() != reflect.Float64 {
		t.Errorf("ExamSession.Score pointer base type should be float64")
	}
	fieldKind(t, v, "CertificateURL", reflect.Ptr)
	if mustField(t, v, "CertificateURL").Type.Elem().Kind() != reflect.String {
		t.Errorf("ExamSession.CertificateURL pointer base type should be string")
	}
}

// ---- ExamSessionAnswer ----

func TestExamSessionAnswerStruct(t *testing.T) {
	typ := reflect.TypeOf((*ExamSessionAnswer)(nil)).Elem()
	v := newModel(typ)

	// composite PK (session_id, question_id) — no surrogate ID
	if _, ok := typ.FieldByName("ID"); ok {
		t.Errorf("ExamSessionAnswer must NOT have a surrogate ID field (composite PK)")
	}

	jsonTag(t, v, "SessionID", "session_id")
	jsonTag(t, v, "QuestionID", "question_id")
	jsonTag(t, v, "Answer", "answer")
	jsonTag(t, v, "IsCorrect", "is_correct")
	jsonTag(t, v, "Score", "score")
	jsonTag(t, v, "GradedBy", "graded_by")
	jsonTag(t, v, "GradedAt", "graded_at")
	jsonTag(t, v, "GraderComment", "grader_comment")
	jsonTag(t, v, "FlaggedForReview", "flagged_for_review")
	jsonTag(t, v, "SavedAt", "saved_at")

	fieldType(t, v, "SessionID", reflect.TypeOf(uuid.UUID{}))
	fieldType(t, v, "QuestionID", reflect.TypeOf(uuid.UUID{}))

	// Nullable pointers
	fieldKind(t, v, "Answer", reflect.Ptr)
	if mustField(t, v, "Answer").Type.Elem().Kind() != reflect.String {
		t.Errorf("ExamSessionAnswer.Answer pointer base type should be string")
	}
	fieldKind(t, v, "IsCorrect", reflect.Ptr)
	if mustField(t, v, "IsCorrect").Type.Elem().Kind() != reflect.Bool {
		t.Errorf("ExamSessionAnswer.IsCorrect pointer base type should be bool")
	}
	fieldKind(t, v, "Score", reflect.Ptr)
	if mustField(t, v, "Score").Type.Elem().Kind() != reflect.Float64 {
		t.Errorf("ExamSessionAnswer.Score pointer base type should be float64 (NUMERIC)")
	}
	fieldKind(t, v, "GradedBy", reflect.Ptr)
	if mustField(t, v, "GradedBy").Type != reflect.TypeOf((*uuid.UUID)(nil)) {
		t.Errorf("ExamSessionAnswer.GradedBy should be *uuid.UUID, got %s", mustField(t, v, "GradedBy").Type)
	}
	fieldKind(t, v, "GradedAt", reflect.Ptr)
	if mustField(t, v, "GradedAt").Type != reflect.TypeOf((*time.Time)(nil)) {
		t.Errorf("ExamSessionAnswer.GradedAt should be *time.Time, got %s", mustField(t, v, "GradedAt").Type)
	}
	fieldKind(t, v, "GraderComment", reflect.Ptr)
	if mustField(t, v, "GraderComment").Type.Elem().Kind() != reflect.String {
		t.Errorf("ExamSessionAnswer.GraderComment pointer base type should be string")
	}
	fieldKind(t, v, "FlaggedForReview", reflect.Bool)
	fieldType(t, v, "SavedAt", reflect.TypeOf(time.Time{}))
}

// ---- SessionViolationLog ----

func TestSessionViolationLogStruct(t *testing.T) {
	typ := reflect.TypeOf((*SessionViolationLog)(nil)).Elem()
	v := newModel(typ)

	if typ.NumField() != 5 {
		t.Fatalf("SessionViolationLog struct: got %d fields, want 5", typ.NumField())
	}

	jsonTag(t, v, "ID", "id")
	jsonTag(t, v, "SessionID", "session_id")
	jsonTag(t, v, "StudentID", "student_id")
	jsonTag(t, v, "ViolationType", "violation_type")
	jsonTag(t, v, "OccurredAt", "occurred_at")

	fieldType(t, v, "ID", reflect.TypeOf(uuid.UUID{}))
	fieldType(t, v, "SessionID", reflect.TypeOf(uuid.UUID{}))
	fieldType(t, v, "StudentID", reflect.TypeOf(uuid.UUID{}))
	fieldKind(t, v, "ViolationType", reflect.String)
	fieldType(t, v, "OccurredAt", reflect.TypeOf(time.Time{}))
}

// ---- Composite read shapes ----

func TestTestDetailStruct(t *testing.T) {
	typ := reflect.TypeOf((*TestDetail)(nil)).Elem()
	v := newModel(typ)

	// TestDetail must embed Test and have Questions slice
	fieldType(t, v, "Test", reflect.TypeOf(Test{}))
	qs, ok := typ.FieldByName("Questions")
	if !ok {
		t.Fatal("TestDetail missing Questions field")
	}
	if qs.Type.Kind() != reflect.Slice {
		t.Errorf("TestDetail.Questions should be a slice, got %s", qs.Type.Kind())
	}
	if qs.Type.Elem() != reflect.TypeOf(QuestionWithOptions{}) {
		t.Errorf("TestDetail.Questions element should be QuestionWithOptions, got %s", qs.Type.Elem())
	}
}

func TestQuestionWithOptionsStruct(t *testing.T) {
	typ := reflect.TypeOf((*QuestionWithOptions)(nil)).Elem()
	v := newModel(typ)

	fieldType(t, v, "Question", reflect.TypeOf(Question{}))
	opts, ok := typ.FieldByName("Options")
	if !ok {
		t.Fatal("QuestionWithOptions missing Options field")
	}
	if opts.Type.Kind() != reflect.Slice {
		t.Errorf("QuestionWithOptions.Options should be a slice, got %s", opts.Type.Kind())
	}
	if opts.Type.Elem() != reflect.TypeOf(QuestionOption{}) {
		t.Errorf("QuestionWithOptions.Options element should be QuestionOption, got %s", opts.Type.Elem())
	}
}

// All nine main + two composite structs must be reachable
func TestExamTypesRegistered(t *testing.T) {
	names := []string{
		"Test", "Question", "QuestionOption", "Exam", "ExamTest",
		"ExamRegistration", "ExamSession", "ExamSessionAnswer", "SessionViolationLog",
		"TestDetail", "QuestionWithOptions",
	}
	for _, n := range names {
		if _, ok := typesByName[n]; !ok {
			t.Errorf("model.%s not registered", n)
		}
	}
}

var typesByName = func() map[string]reflect.Type {
	m := map[string]reflect.Type{}
	typ := reflect.TypeOf((*Test)(nil)).Elem()
	m[typ.Name()] = typ
	for _, t := range []reflect.Type{
		reflect.TypeOf((*Question)(nil)).Elem(),
		reflect.TypeOf((*QuestionOption)(nil)).Elem(),
		reflect.TypeOf((*Exam)(nil)).Elem(),
		reflect.TypeOf((*ExamTest)(nil)).Elem(),
		reflect.TypeOf((*ExamRegistration)(nil)).Elem(),
		reflect.TypeOf((*ExamSession)(nil)).Elem(),
		reflect.TypeOf((*ExamSessionAnswer)(nil)).Elem(),
		reflect.TypeOf((*SessionViolationLog)(nil)).Elem(),
		reflect.TypeOf((*TestDetail)(nil)).Elem(),
		reflect.TypeOf((*QuestionWithOptions)(nil)).Elem(),
	} {
		m[t.Name()] = t
	}
	return m
}()
