package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"akademi-bimbel/internal/service"
)

// ---------------------------------------------------------------------------
// Local seed helpers for monitor integration tests
// ---------------------------------------------------------------------------

func seedMonitorExam(t *testing.T, pool *pgxpool.Pool, title string, durationMinutes int, graceMinutes int) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	err := pool.QueryRow(context.Background(),
		`INSERT INTO exam (title, timer_mode, duration_minutes, grace_window_minutes, status)
		 VALUES ($1, 'overall', $2, $3, 'published') RETURNING id`,
		title, durationMinutes, graceMinutes,
	).Scan(&id)
	if err != nil {
		t.Fatalf("insert exam: %v", err)
	}
	return id
}

func seedMonitorPerTestExam(t *testing.T, pool *pgxpool.Pool, title string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	err := pool.QueryRow(context.Background(),
		`INSERT INTO exam (title, timer_mode, duration_minutes, status)
		 VALUES ($1, 'per_test', NULL, 'published') RETURNING id`,
		title,
	).Scan(&id)
	if err != nil {
		t.Fatalf("insert per_test exam: %v", err)
	}
	return id
}

func seedMonitorCheckIn(t *testing.T, pool *pgxpool.Pool, regID uuid.UUID) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		`UPDATE exam_registration SET checked_in_at = now() WHERE id = $1`, regID)
	if err != nil {
		t.Fatalf("update checked_in_at: %v", err)
	}
}

func seedMonitorSession(t *testing.T, pool *pgxpool.Pool, regID, studentID, examID uuid.UUID, startedAt time.Time, status string, extendedUntil *time.Time, submittedAt *time.Time) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	err := pool.QueryRow(context.Background(),
		`INSERT INTO exam_session (registration_id, student_id, exam_id, started_at, status, extended_until, submitted_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		regID, studentID, examID, startedAt, status, extendedUntil, submittedAt,
	).Scan(&id)
	if err != nil {
		t.Fatalf("insert exam_session: %v", err)
	}
	return id
}

func seedMonitorViolation(t *testing.T, pool *pgxpool.Pool, sessionID, studentID uuid.UUID, vtype string, occurredAt time.Time) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO session_violation_log (session_id, student_id, violation_type, occurred_at)
		 VALUES ($1, $2, $3, $4)`,
		sessionID, studentID, vtype, occurredAt,
	)
	if err != nil {
		t.Fatalf("insert violation: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Session Monitor — full-flow integration test
// ---------------------------------------------------------------------------

// TestAdminSessionMonitor_FullFlow_Returns200 seeds one exam with 5 registrants
// in each derived status (registered, checked_in, in_progress, overdue,
// submitted), adds answers and violations, and asserts the full monitor
// payload: exam summary, row count = 5, each derived status correct,
// answers_saved / total_questions, violation_count, and violations_recent
// (newest-first, correct aggregation).
func TestAdminSessionMonitor_FullFlow_Returns200(t *testing.T) {
	env := newTestEnvWithStore(t)

	// School + users
	schoolID := seedSchool(t, env.pool)

	adminID := seedUser(t, env.pool, "admin_exam", "Monitor Admin")

	studReg := seedUserWithSchool(t, env.pool, "student", "Reg Student", schoolID)
	studCI := seedUserWithSchool(t, env.pool, "student", "CheckedIn Student", schoolID)
	studIP := seedUserWithSchool(t, env.pool, "student", "InProgress Student", schoolID)
	studOD := seedUserWithSchool(t, env.pool, "student", "Overdue Student", schoolID)
	studSub := seedUserWithSchool(t, env.pool, "student", "Submitted Student", schoolID)

	// Exam with duration=60, grace=5
	examID := seedMonitorExam(t, env.pool, "Monitor Exam", 60, 5)

	// Test with 3 MC questions
	testID := seedTest(t, env.pool)
	seedMCQuestion(t, env.pool, testID, "Q1", 1, 1)
	seedMCQuestion(t, env.pool, testID, "Q2", 1, 2)
	seedMCQuestion(t, env.pool, testID, "Q3", 1, 3)
	seedExamTest(t, env.pool, examID, testID, 1)

	// Registrations (5)
	seedRegistration(t, env.pool, studReg, examID)
	regCI := seedRegistration(t, env.pool, studCI, examID)
	regIP := seedRegistration(t, env.pool, studIP, examID)
	regOD := seedRegistration(t, env.pool, studOD, examID)
	regSub := seedRegistration(t, env.pool, studSub, examID)

	// Student 2: checked_in (no session)
	seedMonitorCheckIn(t, env.pool, regCI)

	now := time.Now()

	// Student 3: in_progress (started 10 min ago → deadline in future)
	seedMonitorCheckIn(t, env.pool, regIP)
	sessIP := seedMonitorSession(t, env.pool, regIP, studIP, examID, now.Add(-10*time.Minute), "in_progress", nil, nil)

	// Student 4: overdue (started 75 min ago → deadline 65 min: 75+60+5 = 10 min ago)
	seedMonitorCheckIn(t, env.pool, regOD)
	seedMonitorSession(t, env.pool, regOD, studOD, examID, now.Add(-75*time.Minute), "in_progress", nil, nil)

	// Student 5: submitted (started 30 min ago)
	seedMonitorCheckIn(t, env.pool, regSub)
	submittedAt := now.Add(-1 * time.Minute)
	sessSub := seedMonitorSession(t, env.pool, regSub, studSub, examID, now.Add(-30*time.Minute), "submitted", nil, &submittedAt)

	// Answers: Student 5 answered 2 of 3 questions
	// Re-fetch question IDs for the answer inserts
	rows, err := env.pool.Query(context.Background(),
		`SELECT q.id FROM question q
		 JOIN exam_test et ON et.test_id = q.test_id
		 WHERE et.exam_id = $1 ORDER BY q.sort_order`, examID)
	if err != nil {
		t.Fatalf("query questions: %v", err)
	}
	var qIDs []uuid.UUID
	for rows.Next() {
		var qID uuid.UUID
		if err := rows.Scan(&qID); err != nil {
			t.Fatalf("scan question: %v", err)
		}
		qIDs = append(qIDs, qID)
	}
	rows.Close()
	if len(qIDs) < 3 {
		t.Fatalf("want 3 questions, got %d", len(qIDs))
	}
	seedAnswer(t, env.pool, sessSub, qIDs[0], "a", 1)
	seedAnswer(t, env.pool, sessSub, qIDs[1], "b", 0)

	// Violations: 2 on sessSub (tab_switch older, copy_attempt newer),
	// 1 on sessIP (tab_switch)
	seedMonitorViolation(t, env.pool, sessSub, studSub, "tab_switch", now.Add(-5*time.Minute))
	seedMonitorViolation(t, env.pool, sessSub, studSub, "copy_attempt", now.Add(-2*time.Minute))
	seedMonitorViolation(t, env.pool, sessIP, studIP, "tab_switch", now.Add(-3*time.Minute))

	// Mint admin_exam token and call monitor endpoint
	token := mintTokenForEnv(t, env, adminID.String(), service.RoleAdminExam)
	rec := getRequest(t, env.e, "/api/v1/admin/sessions/monitor?exam_id="+examID.String(), token)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	// --- Assert exam summary block ---
	examField, ok := resp["exam"].(map[string]any)
	if !ok {
		t.Fatal("response missing exam field")
	}
	if idStr := examField["id"].(string); idStr != examID.String() {
		t.Errorf("exam.id: want %s, got %s", examID.String(), idStr)
	}
	if title := examField["title"].(string); title != "Monitor Exam" {
		t.Errorf("exam.title: want Monitor Exam, got %s", title)
	}
	if d := examField["duration_minutes"].(float64); d != 60 {
		t.Errorf("exam.duration_minutes: want 60, got %v", d)
	}
	if g := examField["grace_window_minutes"].(float64); g != 5 {
		t.Errorf("exam.grace_window_minutes: want 5, got %v", g)
	}
	if s := examField["status"].(string); s != "published" {
		t.Errorf("exam.status: want published, got %s", s)
	}

	// --- Assert rows ---
	rawRows, ok := resp["rows"].([]any)
	if !ok {
		t.Fatal("rows is not an array")
	}
	if len(rawRows) != 5 {
		t.Fatalf("want 5 rows, got %d", len(rawRows))
	}

	// Index by student_name
	byName := make(map[string]map[string]any, 5)
	for _, r := range rawRows {
		row := r.(map[string]any)
		name, _ := row["student_name"].(string)
		byName[name] = row
	}

	// registered
	regRow, ok := byName["Reg Student"]
	if !ok {
		t.Fatal("missing row for Reg Student")
	}
	if s := regRow["status"].(string); s != "registered" {
		t.Errorf("Reg Student status: want registered, got %s", s)
	}
	if v := regRow["answers_saved"].(float64); v != 0 {
		t.Errorf("Reg Student answers_saved: want 0, got %v", v)
	}
	if v := regRow["total_questions"].(float64); v != 3 {
		t.Errorf("Reg Student total_questions: want 3, got %v", v)
	}
	if v := regRow["violation_count"].(float64); v != 0 {
		t.Errorf("Reg Student violation_count: want 0, got %v", v)
	}
	if regRow["session_id"] != nil {
		t.Error("Reg Student session_id should be nil")
	}

	// checked_in
	ciRow, ok := byName["CheckedIn Student"]
	if !ok {
		t.Fatal("missing row for CheckedIn Student")
	}
	if s := ciRow["status"].(string); s != "checked_in" {
		t.Errorf("CheckedIn Student status: want checked_in, got %s", s)
	}
	if v := ciRow["answers_saved"].(float64); v != 0 {
		t.Errorf("CheckedIn Student answers_saved: want 0, got %v", v)
	}
	if v := ciRow["total_questions"].(float64); v != 3 {
		t.Errorf("CheckedIn Student total_questions: want 3, got %v", v)
	}
	if v := ciRow["violation_count"].(float64); v != 0 {
		t.Errorf("CheckedIn Student violation_count: want 0, got %v", v)
	}
	if ciRow["session_id"] != nil {
		t.Error("CheckedIn Student session_id should be nil")
	}
	if ciRow["checked_in_at"] == nil {
		t.Error("CheckedIn Student checked_in_at should not be nil")
	}

	// in_progress
	ipRow, ok := byName["InProgress Student"]
	if !ok {
		t.Fatal("missing row for InProgress Student")
	}
	if s := ipRow["status"].(string); s != "in_progress" {
		t.Errorf("InProgress Student status: want in_progress, got %s", s)
	}
	if v := ipRow["total_questions"].(float64); v != 3 {
		t.Errorf("InProgress Student total_questions: want 3, got %v", v)
	}
	if v := ipRow["violation_count"].(float64); v != 1 {
		t.Errorf("InProgress Student violation_count: want 1, got %v", v)
	}
	if ipRow["session_id"] == nil {
		t.Error("InProgress Student session_id should not be nil")
	}

	// overdue
	odRow, ok := byName["Overdue Student"]
	if !ok {
		t.Fatal("missing row for Overdue Student")
	}
	if s := odRow["status"].(string); s != "overdue" {
		t.Errorf("Overdue Student status: want overdue, got %s", s)
	}
	if v := odRow["total_questions"].(float64); v != 3 {
		t.Errorf("Overdue Student total_questions: want 3, got %v", v)
	}
	if v := odRow["violation_count"].(float64); v != 0 {
		t.Errorf("Overdue Student violation_count: want 0, got %v", v)
	}
	if odRow["session_id"] == nil {
		t.Error("Overdue Student session_id should not be nil")
	}

	// submitted
	subRow, ok := byName["Submitted Student"]
	if !ok {
		t.Fatal("missing row for Submitted Student")
	}
	if s := subRow["status"].(string); s != "submitted" {
		t.Errorf("Submitted Student status: want submitted, got %s", s)
	}
	if v := subRow["answers_saved"].(float64); v != 2 {
		t.Errorf("Submitted Student answers_saved: want 2, got %v", v)
	}
	if v := subRow["total_questions"].(float64); v != 3 {
		t.Errorf("Submitted Student total_questions: want 3, got %v", v)
	}
	if v := subRow["violation_count"].(float64); v != 2 {
		t.Errorf("Submitted Student violation_count: want 2, got %v", v)
	}
	if subRow["session_id"] == nil {
		t.Error("Submitted Student session_id should not be nil")
	}
	if adminSub, ok := subRow["admin_submitted"]; ok && adminSub == true {
		t.Error("Submitted Student admin_submitted should be false")
	}

	// --- Assert violations_recent (newest-first, <= 20) ---
	rawViolations, ok := resp["violations_recent"].([]any)
	if !ok {
		t.Fatal("violations_recent is not an array")
	}
	if len(rawViolations) < 2 {
		t.Fatalf("want >= 2 violations_recent, got %d", len(rawViolations))
	}

	v1 := rawViolations[0].(map[string]any)
	if cnt := v1["count"].(float64); cnt != 2 {
		t.Errorf("first violation_recent count: want 2, got %v", cnt)
	}
	if typ := v1["latest_type"].(string); typ != "copy_attempt" {
		t.Errorf("first violation_recent latest_type: want copy_attempt, got %s", typ)
	}

	v2 := rawViolations[1].(map[string]any)
	if cnt := v2["count"].(float64); cnt != 1 {
		t.Errorf("second violation_recent count: want 1, got %v", cnt)
	}
	if typ := v2["latest_type"].(string); typ != "tab_switch" {
		t.Errorf("second violation_recent latest_type: want tab_switch, got %s", typ)
	}
}

// TestAdminSessionMonitor_OverdueViaExtendedUntil_Returns200 verifies that a
// session with duration=NULL (per_test) and extended_until in the past
// resolves to status=overdue via the deadline helper.
func TestAdminSessionMonitor_OverdueViaExtendedUntil_Returns200(t *testing.T) {
	env := newTestEnvWithStore(t)

	adminID := seedUser(t, env.pool, "admin_exam", "ExtAdmin")
	studentID := seedUserWithSchool(t, env.pool, "student", "ExtendOverdue Student", seedSchool(t, env.pool))

	// per_test exam: duration=NULL, grace=NULL
	examID := seedMonitorPerTestExam(t, env.pool, "PerTest Exam")

	testID := seedTest(t, env.pool)
	seedMCQuestion(t, env.pool, testID, "Q1", 1, 1)
	seedExamTest(t, env.pool, examID, testID, 1)

	regID := seedRegistration(t, env.pool, studentID, examID)
	seedMonitorCheckIn(t, env.pool, regID)

	now := time.Now()
	extendedUntil := now.Add(-1 * time.Minute) // passed
	seedMonitorSession(t, env.pool, regID, studentID, examID, now.Add(-30*time.Minute), "in_progress", &extendedUntil, nil)

	token := mintTokenForEnv(t, env, adminID.String(), service.RoleAdminExam)
	rec := getRequest(t, env.e, "/api/v1/admin/sessions/monitor?exam_id="+examID.String(), token)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)

	rows, ok := resp["rows"].([]any)
	if !ok || len(rows) != 1 {
		t.Fatalf("want 1 row, got %d", len(rows))
	}
	row := rows[0].(map[string]any)
	if s := row["status"].(string); s != "overdue" {
		t.Errorf("status: want overdue, got %s", s)
	}
}

// TestAdminSessionMonitor_UnknownExam_Returns404 verifies that a monitor
// request for a non-existent exam returns 404.
func TestAdminSessionMonitor_UnknownExam_Returns404(t *testing.T) {
	env := newTestEnvWithStore(t)
	adminID := seedUser(t, env.pool, "admin_exam", "Admin404")
	token := mintTokenForEnv(t, env, adminID.String(), service.RoleAdminExam)

	rec := getRequest(t, env.e, "/api/v1/admin/sessions/monitor?exam_id=00000000-0000-0000-0000-0000000000aa", token)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d body=%s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	json.NewDecoder(rec.Body).Decode(&body)
	if code, _ := body["code"].(string); code != "exam_not_found" {
		t.Errorf("code: want exam_not_found, got %v", body["code"])
	}
}

// ---------------------------------------------------------------------------
// Session Violations Detail — full-flow integration test
// ---------------------------------------------------------------------------

// TestAdminSessionViolations_FullFlow_Returns200 seeds a session with two
// violations (tab_switch older, copy_attempt newer) and asserts the detail
// endpoint returns them newest-first.
func TestAdminSessionViolations_FullFlow_Returns200(t *testing.T) {
	env := newTestEnvWithStore(t)

	adminID := seedUser(t, env.pool, "admin_exam", "ViolAdmin")
	studentID := seedUser(t, env.pool, "student", "Viol Student")

	examID := seedMonitorExam(t, env.pool, "Viol Exam", 60, 5)
	testID := seedTest(t, env.pool)
	seedMCQuestion(t, env.pool, testID, "Q1", 1, 1)
	seedExamTest(t, env.pool, examID, testID, 1)

	regID := seedRegistration(t, env.pool, studentID, examID)
	now := time.Now()
	submittedAt := now.Add(-1 * time.Minute)
	sessID := seedMonitorSession(t, env.pool, regID, studentID, examID, now.Add(-30*time.Minute), "submitted", nil, &submittedAt)

	seedMonitorViolation(t, env.pool, sessID, studentID, "tab_switch", now.Add(-5*time.Minute))
	seedMonitorViolation(t, env.pool, sessID, studentID, "copy_attempt", now.Add(-2*time.Minute))

	token := mintTokenForEnv(t, env, adminID.String(), service.RoleAdminExam)
	rec := getRequest(t, env.e, "/api/v1/admin/sessions/"+sessID.String()+"/violations", token)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	data, ok := resp["data"].([]any)
	if !ok {
		t.Fatal("data is not an array")
	}
	if len(data) != 2 {
		t.Fatalf("want 2 violations, got %d", len(data))
	}

	// Newest first: copy_attempt then tab_switch
	v1 := data[0].(map[string]any)
	if typ := v1["violation_type"].(string); typ != "copy_attempt" {
		t.Errorf("first violation type: want copy_attempt, got %s", typ)
	}
	if idStr := v1["session_id"].(string); idStr != sessID.String() {
		t.Errorf("first violation session_id mismatch: want %s, got %s", sessID.String(), idStr)
	}
	if idStr := v1["student_id"].(string); idStr != studentID.String() {
		t.Errorf("first violation student_id mismatch: want %s, got %s", studentID.String(), idStr)
	}

	v2 := data[1].(map[string]any)
	if typ := v2["violation_type"].(string); typ != "tab_switch" {
		t.Errorf("second violation type: want tab_switch, got %s", typ)
	}
}
