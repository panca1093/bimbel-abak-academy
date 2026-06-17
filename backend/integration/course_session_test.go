package integration_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seedCourseSession inserts a course_session row directly via SQL (bypasses fan-out).
// orderID may be "" to leave it NULL.
func seedCourseSession(t *testing.T, env *testEnv, studentID, courseID, orderID string) string {
	t.Helper()
	ctx := context.Background()
	var sessionID string
	if orderID == "" {
		err := env.pool.QueryRow(ctx,
			`INSERT INTO course_session (student_id, course_id, status, source)
			 VALUES ($1, $2, 'active', 'order') RETURNING id`,
			studentID, courseID,
		).Scan(&sessionID)
		require.NoError(t, err)
	} else {
		err := env.pool.QueryRow(ctx,
			`INSERT INTO course_session (student_id, course_id, order_id, status, source)
			 VALUES ($1, $2, $3, 'active', 'order') RETURNING id`,
			studentID, courseID, orderID,
		).Scan(&sessionID)
		require.NoError(t, err)
	}
	return sessionID
}

// seedSection inserts a section under a course and returns its ID.
func seedSection(t *testing.T, env *testEnv, courseID string) string {
	t.Helper()
	ctx := context.Background()
	var id string
	err := env.pool.QueryRow(ctx,
		`INSERT INTO section (course_id, title, position) VALUES ($1, 'Section 1', 0) RETURNING id`,
		courseID,
	).Scan(&id)
	require.NoError(t, err)
	return id
}

// seedLesson inserts a lesson under a section and returns its ID.
func seedLesson(t *testing.T, env *testEnv, sectionID string) string {
	t.Helper()
	ctx := context.Background()
	var id string
	err := env.pool.QueryRow(ctx,
		`INSERT INTO lesson (section_id, title, position) VALUES ($1, 'Lesson 1', 0) RETURNING id`,
		sectionID,
	).Scan(&id)
	require.NoError(t, err)
	return id
}

// scanCompletedLessons reads completed_lessons JSONB as raw bytes from the DB.
func scanCompletedLessons(t *testing.T, env *testEnv, studentID, courseID string) map[string]string {
	t.Helper()
	ctx := context.Background()
	var raw []byte
	err := env.pool.QueryRow(ctx,
		`SELECT completed_lessons FROM course_session
		 WHERE student_id = $1 AND course_id = $2 AND status = 'active'`,
		studentID, courseID,
	).Scan(&raw)
	require.NoError(t, err)
	var m map[string]string
	require.NoError(t, json.Unmarshal(raw, &m))
	return m
}

// TestCourseSession covers FR-INT-26 and FR-INT-27.
func TestCourseSession(t *testing.T) {
	env := newTestEnv(t)

	t.Run("FR-INT-26 list courses returns only the calling student's active sessions", func(t *testing.T) {
		studentA := seedUser(t, env, "student", "active", false)
		studentB := seedUser(t, env, "student", "active", false)
		tokenA := authToken(t, env, studentA, "student")
		tokenB := authToken(t, env, studentB, "student")

		courseA := seedCourse(t, env, "Course for A")
		courseB := seedCourse(t, env, "Course for B")

		seedCourseSession(t, env, studentA, courseA, "")
		seedCourseSession(t, env, studentB, courseB, "")

		// Student A should see only courseA's session.
		respA := env.doJSON(t, http.MethodGet, "/api/v1/courses", nil, tokenA)
		bodyA := decodeBody(t, respA)
		require.Equal(t, http.StatusOK, respA.StatusCode, "body: %v", bodyA)

		dataA, ok := bodyA["data"].([]any)
		require.True(t, ok, "expected 'data' array in response, got: %v", bodyA)
		require.Len(t, dataA, 1, "student A should see exactly 1 session")

		// model.CourseSession has no json tags — serialized with PascalCase keys.
		sessionA := dataA[0].(map[string]any)
		assert.Equal(t, courseA, sessionA["CourseID"], "student A's session must reference courseA")

		// Student B should see only courseB's session.
		respB := env.doJSON(t, http.MethodGet, "/api/v1/courses", nil, tokenB)
		bodyB := decodeBody(t, respB)
		require.Equal(t, http.StatusOK, respB.StatusCode, "body: %v", bodyB)

		dataB, ok := bodyB["data"].([]any)
		require.True(t, ok, "expected 'data' array in response, got: %v", bodyB)
		require.Len(t, dataB, 1, "student B should see exactly 1 session")

		sessionB := dataB[0].(map[string]any)
		assert.Equal(t, courseB, sessionB["CourseID"], "student B's session must reference courseB")
	})

	t.Run("FR-INT-27 mark lesson complete updates JSONB; re-call is idempotent", func(t *testing.T) {
		student := seedUser(t, env, "student", "active", false)
		token := authToken(t, env, student, "student")

		course := seedCourse(t, env, "Course with Lessons")
		section := seedSection(t, env, course)
		lesson := seedLesson(t, env, section)

		seedCourseSession(t, env, student, course, "")

		// First call: mark lesson complete.
		resp1 := env.doJSON(t, http.MethodPost, "/api/v1/courses/"+course+"/lessons/"+lesson+"/complete", nil, token)
		body1 := decodeBody(t, resp1)
		require.Equal(t, http.StatusOK, resp1.StatusCode, "first mark-complete body: %v", body1)

		// Verify completed_lessons JSONB has the lesson key with a timestamp.
		lessons1 := scanCompletedLessons(t, env, student, course)
		ts1, ok := lessons1[lesson]
		require.True(t, ok, "lesson key should be present in completed_lessons after first call")
		assert.NotEmpty(t, ts1, "timestamp should be non-empty")

		// Parse the timestamp to ensure it's valid RFC3339 format.
		_, err := time.Parse(time.RFC3339Nano, ts1)
		require.NoError(t, err, "stored timestamp must be RFC3339Nano")

		// Brief pause so a clock-based second write would produce a different timestamp.
		time.Sleep(10 * time.Millisecond)

		// Second call: same request — must be idempotent.
		resp2 := env.doJSON(t, http.MethodPost, "/api/v1/courses/"+course+"/lessons/"+lesson+"/complete", nil, token)
		body2 := decodeBody(t, resp2)
		require.Equal(t, http.StatusOK, resp2.StatusCode, "second mark-complete body: %v", body2)

		// Timestamp must be unchanged (WHERE NOT (completed_lessons ? $2) guard).
		lessons2 := scanCompletedLessons(t, env, student, course)
		ts2, ok := lessons2[lesson]
		require.True(t, ok, "lesson key should still be present after second call")
		assert.Equal(t, ts1, ts2, "timestamp must not change on re-call (idempotency guard)")

		// Completed count must not double-count.
		respProg := env.doJSON(t, http.MethodGet, "/api/v1/courses/"+course+"/progress", nil, token)
		progBody := decodeBody(t, respProg)
		require.Equal(t, http.StatusOK, respProg.StatusCode, "progress body: %v", progBody)

		completedCount, ok := progBody["completed"].(float64)
		require.True(t, ok, "expected 'completed' float in progress body, got: %v", progBody)
		assert.Equal(t, float64(1), completedCount, "completed count must not double-count after re-call")
	})
}
