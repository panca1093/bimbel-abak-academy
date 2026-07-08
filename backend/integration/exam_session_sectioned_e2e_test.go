package integration_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seedSectionedE2EExam seeds a student, a sectioned utbk exam with N tests (each
// with one mcq question having two options, a=correct), exam_test links ordered
// by sort_order, and a registration with attempts_used=0 ready for HTTP start.
// result_config is 'score_pembahasan' so the result endpoint returns breakdown
// immediately after submit.
func seedSectionedE2EExam(t *testing.T, env *testEnv, sectionDurations []int) (studentID, examID, registrationID string, testIDs, questionIDs []string) {
	t.Helper()
	ctx := context.Background()

	studentID = seedUser(t, env, "student", "active", false)

	var eID string
	err := env.pool.QueryRow(ctx,
		`INSERT INTO exam (title, is_free, requires_checkin, timer_mode, result_config, mode)
		 VALUES ($1, true, false, 'per_test', 'score_pembahasan', 'utbk') RETURNING id`,
		"E2E Timer Test "+uuid.NewString()[:8],
	).Scan(&eID)
	require.NoError(t, err)

	for i, dur := range sectionDurations {
		var testID string
		err = env.pool.QueryRow(ctx,
			`INSERT INTO test (title, subject, topic, duration_minutes)
			 VALUES ($1, 'umum', 'penalaran', $2) RETURNING id`,
			fmt.Sprintf("Section %d E2E", i+1), dur,
		).Scan(&testID)
		require.NoError(t, err)
		testIDs = append(testIDs, testID)

		var qID string
		err = env.pool.QueryRow(ctx,
			`INSERT INTO question (test_id, format, body, sort_order, point_correct, point_wrong)
			 VALUES ($1, 'mcq', $2, 0, 1, 0) RETURNING id`,
			testID, fmt.Sprintf("E2E Q%d", i+1),
		).Scan(&qID)
		require.NoError(t, err)
		questionIDs = append(questionIDs, qID)

		_, err = env.pool.Exec(ctx,
			`INSERT INTO question_option (question_id, key, text, is_correct, sort_order)
			 VALUES ($1, 'a', 'correct', true, 0), ($1, 'b', 'wrong', false, 1)`,
			qID,
		)
		require.NoError(t, err)

		_, err = env.pool.Exec(ctx,
			`INSERT INTO exam_test (exam_id, test_id, sort_order) VALUES ($1, $2, $3)`,
			eID, testID, i,
		)
		require.NoError(t, err)
	}

	token := "tok-" + uuid.NewString()[:8]
	var regID string
	err = env.pool.QueryRow(ctx,
		`INSERT INTO exam_registration (student_id, exam_id, token, status)
		 VALUES ($1, $2, $3, 'registered') RETURNING id`,
		studentID, eID, token,
	).Scan(&regID)
	require.NoError(t, err)

	return studentID, eID, regID, testIDs, questionIDs
}

// TestExam_SectionedTimerCorrectness_E2E covers NFR-1: the per-section timer must
// be correct end-to-end over an HTTP start → save → locked-section reject → advance
// → monitor-overdue → submit → grade flow. It directly guards the PR#25 "per-test
// timer broken E2E → instant 0-score auto-submit" regression class.
func TestExam_SectionedTimerCorrectness_E2E(t *testing.T) {
	env := newTestEnv(t)
	ctx := context.Background()

	// ---------- Seed ----------
	// Two sections: section 1 = 30 min, section 2 = 45 min (distinct durations).
	studentID, examID, regID, testIDs, questionIDs := seedSectionedE2EExam(t, env, []int{30, 45})
	require.Len(t, testIDs, 2)
	require.Len(t, questionIDs, 2)

	studentToken := authToken(t, env, studentID, "student")

	// ---------- Start session ----------
	resp, startOut := doJSONBody(t, env, http.MethodPost, "/api/v1/exam/sessions",
		map[string]any{"registration_id": regID}, studentToken)
	require.Equal(t, http.StatusOK, resp.StatusCode, "start session body=%v", startOut)

	sessionID := startOut["session_id"].(string)

	// FR-7/FR-9: verify timer on fresh start.
	assert.Equal(t, "utbk", startOut["mode"])
	require.NotNil(t, startOut["active_test_id"], "active_test_id must be present")
	assert.Equal(t, testIDs[0], startOut["active_test_id"])

	tests, ok := startOut["tests"].([]any)
	require.True(t, ok, "tests must be array")
	require.Len(t, tests, 2)

	sec1 := tests[0].(map[string]any)
	sec2 := tests[1].(map[string]any)

	// FR-9: first section active with remaining > 0 (guards PR#25 regression).
	// jsonNum handles omitempty (zero int64→JSON absent→map nil→0).
	assert.Equal(t, "active", sec1["status"])
	rem1 := jsonNum(sec1["remaining_seconds"])
	assert.Greater(t, rem1, 0.0, "FR-9: section 1 remaining must be > 0 (no broken timer)")
	assert.LessOrEqual(t, rem1, float64(30*60), "section 1 remaining must not exceed 30 min")

	dur1, ok := sec1["duration_minutes"].(float64)
	require.True(t, ok, "duration_minutes must be present on active section")
	assert.Equal(t, float64(30), dur1)

	// Second section is pending with 0 remaining (omitempty → absent from JSON).
	assert.Equal(t, "pending", sec2["status"])
	assert.Equal(t, float64(0), jsonNum(sec2["remaining_seconds"]),
		"section 2 pending remaining must be 0")
	dur2, ok := sec2["duration_minutes"].(float64)
	require.True(t, ok, "duration_minutes must be present on pending section")
	assert.Equal(t, float64(45), dur2)

	// Extract question IDs from the start payload.
	q1ID := tests[0].(map[string]any)["questions"].([]any)[0].(map[string]any)["id"].(string)
	q2ID := tests[1].(map[string]any)["questions"].([]any)[0].(map[string]any)["id"].(string)

	// ---------- Save answer in section 1 (active) → success ----------
	t.Run("save answer in active section succeeds", func(t *testing.T) {
		resp, out := doJSONBody(t, env, http.MethodPatch, "/api/v1/exam/sessions/"+sessionID+"/answers",
			map[string]any{"answers": []map[string]any{
				{"question_id": q1ID, "answer": "a"},
			}}, studentToken)
		require.Equal(t, http.StatusOK, resp.StatusCode, "body=%v", out)
	})

	// ---------- Save answer targeting section 2 (pending) → rejected ----------
	t.Run("save answer in locked section rejects with ErrSectionLocked", func(t *testing.T) {
		resp, out := doJSONBody(t, env, http.MethodPatch, "/api/v1/exam/sessions/"+sessionID+"/answers",
			map[string]any{"answers": []map[string]any{
				{"question_id": q2ID, "answer": "a"},
			}}, studentToken)
		require.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode, "body=%v", out)
		assert.Equal(t, "section_locked", out["code"], "must reject with section_locked error code")
	})

	// ---------- Advance section 1 → promote section 2 ----------
	t.Run("advance section 1 and verify FR-9/FR-10/FR-11", func(t *testing.T) {
		resp, out := doJSONBody(t, env, http.MethodPost,
			"/api/v1/exam/sessions/"+sessionID+"/sections/"+testIDs[0]+"/advance",
			nil, studentToken)
		require.Equal(t, http.StatusOK, resp.StatusCode, "advance body=%v", out)

		// Advance closes section 1 (submitted) and promotes section 2 (active).
		require.NotNil(t, out["active_test_id"], "active_test_id must be set after advance")
		assert.Equal(t, testIDs[1], out["active_test_id"])
		assert.Equal(t, false, out["completed"], "not completed when next section exists")

		// FR-10: mode preserved in response.
		assert.Equal(t, "utbk", out["mode"])

		advTests, ok := out["tests"].([]any)
		require.True(t, ok)
		require.Len(t, advTests, 2)

		// Section 1 is now submitted with 0 remaining (omitempty → absent).
		advSec1 := advTests[0].(map[string]any)
		assert.Equal(t, "submitted", advSec1["status"])
		assert.Equal(t, float64(0), jsonNum(advSec1["remaining_seconds"]),
			"submitted section remaining must be 0")

		// FR-9 after advance: section 2 active with remaining > 0 (guards PR#25
		// regression where the next section would get 0 remaining).
		advSec2 := advTests[1].(map[string]any)
		assert.Equal(t, "active", advSec2["status"])
		rem2 := jsonNum(advSec2["remaining_seconds"])
		assert.Greater(t, rem2, 0.0, "FR-9: promoted section remaining must be > 0 (no broken timer)")
		assert.LessOrEqual(t, rem2, float64(45*60), "promoted section remaining must not exceed 45 min")
	})

	// ---------- Force the active section deadline past → monitor overdue ----------
	t.Run("forced deadline past appears as overdue in monitor", func(t *testing.T) {
		// Push the active section's started_at 60 minutes into the past so its deadline
		// (45 min duration) is well past.
		_, err := env.pool.Exec(ctx,
			`UPDATE exam_session_section SET started_at = now() - interval '60 minutes' WHERE session_id = $1 AND test_id = $2`,
			sessionID, testIDs[1],
		)
		require.NoError(t, err)

		adminID := seedUser(t, env, "admin_exam", "active", false)
		adminToken := authToken(t, env, adminID, "admin_exam")

		monResp, monOut := doJSONBody(t, env, http.MethodGet,
			"/api/v1/admin/sessions/monitor?exam_id="+examID,
			nil, adminToken)
		require.Equal(t, http.StatusOK, monResp.StatusCode, "monitor body=%v", monOut)

		rows, ok := monOut["rows"].([]any)
		require.True(t, ok, "rows must be array")
		require.Len(t, rows, 1, "should have exactly 1 registrant row")

		row := rows[0].(map[string]any)
		assert.Equal(t, "overdue", row["status"],
			"section with past deadline must be overdue")
	})

	// ---------- Save section 2 answer, then final-submit ----------
	t.Run("submit grades both sections", func(t *testing.T) {
		// First save an answer in section 2 (now active).
		saveResp, saveOut := doJSONBody(t, env, http.MethodPatch, "/api/v1/exam/sessions/"+sessionID+"/answers",
			map[string]any{"answers": []map[string]any{
				{"question_id": q2ID, "answer": "a"},
			}}, studentToken)
		require.Equal(t, http.StatusOK, saveResp.StatusCode, "save section 2 body=%v", saveOut)

		// Submit session.
		submitResp, submitOut := doJSONBody(t, env, http.MethodPost,
			"/api/v1/exam/sessions/"+sessionID+"/submit",
			nil, studentToken)
		require.Equal(t, http.StatusOK, submitResp.StatusCode, "submit body=%v", submitOut)

		assert.Equal(t, "submitted", submitOut["status"])
		require.NotNil(t, submitOut["score"])
		score := submitOut["score"].(float64)
		assert.Equal(t, float64(2), score, "both answers correct -> score = 1+1 = 2")
	})

	// ---------- Verify per-section breakdown in result ----------
	t.Run("result breakdown includes both sections with earned/max", func(t *testing.T) {
		resultResp, resultOut := doJSONBody(t, env, http.MethodGet,
			"/api/v1/exam/sessions/"+sessionID+"/result",
			nil, studentToken)
		require.Equal(t, http.StatusOK, resultResp.StatusCode, "result body=%v", resultOut)

		assert.Equal(t, "result", resultOut["state"])
		assert.Equal(t, "score_pembahasan", resultOut["result_config"])

		breakdown, ok := resultOut["breakdown"].([]any)
		require.True(t, ok, "breakdown must be present with score_pembahasan config")
		require.Len(t, breakdown, 2, "breakdown should have 2 rows (one per section)")

		// Both sections have correct answers (earned=1, max=1 each).
		b1 := breakdown[0].(map[string]any)
		assert.Equal(t, testIDs[0], b1["test_id"])
		assert.Equal(t, float64(1), b1["earned"], "section 1 correct")
		assert.Equal(t, float64(1), b1["max"], "section 1 max points")

		b2 := breakdown[1].(map[string]any)
		assert.Equal(t, testIDs[1], b2["test_id"])
		assert.Equal(t, float64(1), b2["earned"], "section 2 correct")
		assert.Equal(t, float64(1), b2["max"], "section 2 max points")
	})

}
