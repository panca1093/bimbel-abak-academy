package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func doJSONBody(t *testing.T, env *testEnv, method, path string, body any, token string) (*http.Response, map[string]any) {
	t.Helper()
	resp := env.doJSON(t, method, path, body, token)
	defer resp.Body.Close()
	out := map[string]any{}
	_ = json.NewDecoder(resp.Body).Decode(&out)
	return resp, out
}

// seedTest directly inserts a test row, returns the testID.
func seedTest(t *testing.T, env *testEnv, title, subject, topic string, duration int) string {
	t.Helper()
	ctx := context.Background()
	var id string
	err := env.pool.QueryRow(ctx,
		`INSERT INTO test (title, subject, topic, duration_minutes)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		title, subject, topic, duration,
	).Scan(&id)
	require.NoError(t, err)
	return id
}

// seedQuestion inserts a question row and returns its ID.
func seedQuestion(t *testing.T, env *testEnv, testID, format, body string, sortOrder int) string {
	t.Helper()
	ctx := context.Background()
	var id string
	err := env.pool.QueryRow(ctx,
		`INSERT INTO question (test_id, format, body, sort_order)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		testID, format, body, sortOrder,
	).Scan(&id)
	require.NoError(t, err)
	return id
}

// ----- Test CRUD -----

func TestExam_AdminCreateTest_returns_201_and_test(t *testing.T) {
	env := newTestEnv(t)
	adminID := seedUser(t, env, "admin_exam", "active", false)
	token := authToken(t, env, adminID, "admin_exam")

	body := map[string]any{
		"title":            "Algebra Mid-Term",
		"subject":          "math",
		"topic":            "algebra",
		"duration_minutes": 60,
	}
	resp, out := doJSONBody(t, env, http.MethodPost, "/api/v1/admin/tests", body, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "body=%v", out)
	assert.NotEmpty(t, out["id"])
	assert.Equal(t, "Algebra Mid-Term", out["title"])
	assert.Equal(t, "math", out["subject"])
	assert.Equal(t, float64(60), out["duration_minutes"])
}

func TestExam_AdminCreateTest_returns_400_on_missing_title(t *testing.T) {
	env := newTestEnv(t)
	adminID := seedUser(t, env, "admin_exam", "active", false)
	token := authToken(t, env, adminID, "admin_exam")

	// service-layer validateTest rejects blank title via ErrValidation → 422
	body := map[string]any{
		"title":            "   ",
		"subject":          "math",
		"topic":            "algebra",
		"duration_minutes": 60,
	}
	resp, out := doJSONBody(t, env, http.MethodPost, "/api/v1/admin/tests", body, token)
	assert.True(t, resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnprocessableEntity,
		"want 400 or 422, got %d body=%v", resp.StatusCode, out)
}

func TestExam_AdminCreateTest_returns_400_on_zero_duration(t *testing.T) {
	env := newTestEnv(t)
	adminID := seedUser(t, env, "admin_exam", "active", false)
	token := authToken(t, env, adminID, "admin_exam")

	body := map[string]any{
		"title":            "X",
		"subject":          "math",
		"topic":            "algebra",
		"duration_minutes": 0,
	}
	resp, out := doJSONBody(t, env, http.MethodPost, "/api/v1/admin/tests", body, token)
	assert.True(t, resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnprocessableEntity,
		"want 400 or 422, got %d body=%v", resp.StatusCode, out)
}

func TestExam_AdminListTests_returns_data_and_cursor(t *testing.T) {
	env := newTestEnv(t)
	adminID := seedUser(t, env, "admin_exam", "active", false)
	token := authToken(t, env, adminID, "admin_exam")

	for i := 0; i < 3; i++ {
		body := map[string]any{
			"title":            fmt.Sprintf("Test %d", i),
			"subject":          "math",
			"topic":            "algebra",
			"duration_minutes": 60,
		}
		resp, _ := doJSONBody(t, env, http.MethodPost, "/api/v1/admin/tests", body, token)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	resp, out := doJSONBody(t, env, http.MethodGet, "/api/v1/admin/tests?subject=math", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	data, ok := out["data"].([]any)
	require.True(t, ok, "data should be array, got %T", out["data"])
	assert.Len(t, data, 3)
	assert.NotNil(t, out["next_cursor"])
}

func TestExam_AdminListTests_filters_by_subject(t *testing.T) {
	env := newTestEnv(t)
	adminID := seedUser(t, env, "admin_exam", "active", false)
	token := authToken(t, env, adminID, "admin_exam")

	body1 := map[string]any{"title": "A", "subject": "math", "topic": "algebra", "duration_minutes": 60}
	resp, _ := doJSONBody(t, env, http.MethodPost, "/api/v1/admin/tests", body1, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	body2 := map[string]any{"title": "B", "subject": "biology", "topic": "cells", "duration_minutes": 30}
	resp, _ = doJSONBody(t, env, http.MethodPost, "/api/v1/admin/tests", body2, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	resp, out := doJSONBody(t, env, http.MethodGet, "/api/v1/admin/tests?subject=biology", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	data := out["data"].([]any)
	require.Len(t, data, 1)
	first := data[0].(map[string]any)
	assert.Equal(t, "biology", first["subject"])
}

func TestExam_AdminGetTest_returns_TestDetail(t *testing.T) {
	env := newTestEnv(t)
	adminID := seedUser(t, env, "admin_exam", "active", false)
	token := authToken(t, env, adminID, "admin_exam")

	testID := seedTest(t, env, "X", "math", "algebra", 60)
	_ = seedQuestion(t, env, testID, "essay", "explain", 1)

	resp, out := doJSONBody(t, env, http.MethodGet, "/api/v1/admin/tests/"+testID, nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "body=%v", out)
	assert.NotNil(t, out["test"])
	assert.NotNil(t, out["questions"])
}

func TestExam_AdminUpdateTest_updates_fields(t *testing.T) {
	env := newTestEnv(t)
	adminID := seedUser(t, env, "admin_exam", "active", false)
	token := authToken(t, env, adminID, "admin_exam")

	testID := seedTest(t, env, "Old Title", "math", "algebra", 60)
	body := map[string]any{"title": "New Title", "duration_minutes": 90}
	resp, out := doJSONBody(t, env, http.MethodPatch, "/api/v1/admin/tests/"+testID, body, token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "body=%v", out)
	assert.Equal(t, "New Title", out["title"])
	assert.Equal(t, float64(90), out["duration_minutes"])
}

func TestExam_AdminDeleteTest_returns_204_and_cascades(t *testing.T) {
	env := newTestEnv(t)
	adminID := seedUser(t, env, "admin_exam", "active", false)
	token := authToken(t, env, adminID, "admin_exam")

	testID := seedTest(t, env, "X", "math", "algebra", 60)
	qID := seedQuestion(t, env, testID, "essay", "explain", 1)

	resp := env.doJSON(t, http.MethodDelete, "/api/v1/admin/tests/"+testID, nil, token)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify cascade: question should be gone
	ctx := context.Background()
	var count int
	err := env.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM question WHERE id = $1`, qID,
	).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "question should be cascade-deleted")
}

// ----- Question CRUD -----

func TestExam_AdminCreateQuestion_mcq_with_one_correct_returns_201(t *testing.T) {
	env := newTestEnv(t)
	adminID := seedUser(t, env, "admin_exam", "active", false)
	token := authToken(t, env, adminID, "admin_exam")

	testID := seedTest(t, env, "X", "math", "algebra", 60)
	body := map[string]any{
		"format":     "mcq",
		"body":       "2+2",
		"sort_order": 1,
		"options": []map[string]any{
			{"key": "a", "text": "4", "is_correct": true, "sort_order": 1},
			{"key": "b", "text": "5", "is_correct": false, "sort_order": 2},
		},
	}
	resp, out := doJSONBody(t, env, http.MethodPost, "/api/v1/admin/tests/"+testID+"/questions", body, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "body=%v", out)
	q := out["question"].(map[string]any)
	assert.Equal(t, "mcq", q["format"])
	options := out["options"].([]any)
	assert.Len(t, options, 2)
}

func TestExam_AdminCreateQuestion_mcq_with_two_correct_returns_422(t *testing.T) {
	env := newTestEnv(t)
	adminID := seedUser(t, env, "admin_exam", "active", false)
	token := authToken(t, env, adminID, "admin_exam")

	testID := seedTest(t, env, "X", "math", "algebra", 60)
	body := map[string]any{
		"format":     "mcq",
		"body":       "2+2",
		"sort_order": 1,
		"options": []map[string]any{
			{"key": "a", "text": "4", "is_correct": true, "sort_order": 1},
			{"key": "b", "text": "5", "is_correct": true, "sort_order": 2},
		},
	}
	resp, out := doJSONBody(t, env, http.MethodPost, "/api/v1/admin/tests/"+testID+"/questions", body, token)
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode, "body=%v", out)
	assert.Equal(t, "validation_failed", out["code"])
}

func TestExam_AdminCreateQuestion_short_with_no_correct_answer_returns_422(t *testing.T) {
	env := newTestEnv(t)
	adminID := seedUser(t, env, "admin_exam", "active", false)
	token := authToken(t, env, adminID, "admin_exam")

	testID := seedTest(t, env, "X", "math", "algebra", 60)
	body := map[string]any{
		"format":     "short",
		"body":       "capital of France",
		"sort_order": 1,
	}
	resp, out := doJSONBody(t, env, http.MethodPost, "/api/v1/admin/tests/"+testID+"/questions", body, token)
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode, "body=%v", out)
	assert.Equal(t, "validation_failed", out["code"])
}

func TestExam_AdminCreateQuestion_essay_accepts_no_options_no_correct_answer_returns_201(t *testing.T) {
	env := newTestEnv(t)
	adminID := seedUser(t, env, "admin_exam", "active", false)
	token := authToken(t, env, adminID, "admin_exam")

	testID := seedTest(t, env, "X", "math", "algebra", 60)
	body := map[string]any{
		"format":     "essay",
		"body":       "explain gravity",
		"sort_order": 1,
	}
	resp, out := doJSONBody(t, env, http.MethodPost, "/api/v1/admin/tests/"+testID+"/questions", body, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "body=%v", out)
	q := out["question"].(map[string]any)
	assert.Equal(t, "essay", q["format"])
	options := out["options"].([]any)
	assert.Len(t, options, 0)
}

func TestExam_AdminUpdateQuestion_replaces_options_atomically(t *testing.T) {
	env := newTestEnv(t)
	adminID := seedUser(t, env, "admin_exam", "active", false)
	token := authToken(t, env, adminID, "admin_exam")

	testID := seedTest(t, env, "X", "math", "algebra", 60)
	ctx := context.Background()
	qID := seedQuestion(t, env, testID, "mcq", "2+2", 1)
	_, err := env.pool.Exec(ctx,
		`INSERT INTO question_option (question_id, key, text, is_correct, sort_order)
		 VALUES ($1, 'a', '4', true, 1), ($1, 'b', '5', false, 2)`,
		qID,
	)
	require.NoError(t, err)

	// PATCH with completely different option set — old options gone, new options present
	body := map[string]any{
		"format":     "mcq",
		"body":       "3+3",
		"sort_order": 1,
		"options": []map[string]any{
			{"key": "x", "text": "6", "is_correct": true, "sort_order": 1},
			{"key": "y", "text": "7", "is_correct": false, "sort_order": 2},
		},
	}
	resp, out := doJSONBody(t, env, http.MethodPatch, "/api/v1/admin/questions/"+qID, body, token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "body=%v", out)
	options := out["options"].([]any)
	require.Len(t, options, 2)
	keys := []string{}
	for _, o := range options {
		keys = append(keys, o.(map[string]any)["key"].(string))
	}
	assert.Contains(t, keys, "x")
	assert.Contains(t, keys, "y")
	assert.NotContains(t, keys, "a")
}

func TestExam_AdminDeleteQuestion_returns_204(t *testing.T) {
	env := newTestEnv(t)
	adminID := seedUser(t, env, "admin_exam", "active", false)
	token := authToken(t, env, adminID, "admin_exam")

	testID := seedTest(t, env, "X", "math", "algebra", 60)
	qID := seedQuestion(t, env, testID, "essay", "explain", 1)

	resp := env.doJSON(t, http.MethodDelete, "/api/v1/admin/questions/"+qID, nil, token)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestExam_AdminCreateQuestion_sort_order_conflict_returns_422(t *testing.T) {
	env := newTestEnv(t)
	adminID := seedUser(t, env, "admin_exam", "active", false)
	token := authToken(t, env, adminID, "admin_exam")

	testID := seedTest(t, env, "X", "math", "algebra", 60)
	_ = seedQuestion(t, env, testID, "essay", "first", 1)

	body := map[string]any{
		"format":     "essay",
		"body":       "duplicate sort order",
		"sort_order": 1,
	}
	resp, out := doJSONBody(t, env, http.MethodPost, "/api/v1/admin/tests/"+testID+"/questions", body, token)
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode, "body=%v", out)
	assert.Equal(t, "validation_failed", out["code"])
}

// ----- RBAC enforcement -----

func TestExam_NonAdminRole_gets_403_on_tests_endpoints(t *testing.T) {
	env := newTestEnv(t)
	studentID := seedUser(t, env, "student", "active", false)
	token := authToken(t, env, studentID, "student")

	resp, out := doJSONBody(t, env, http.MethodPost, "/api/v1/admin/tests",
		map[string]any{"title": "X", "subject": "math", "topic": "algebra", "duration_minutes": 60}, token)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "body=%v", out)
	assert.Equal(t, "forbidden", out["code"])
}

func TestExam_NonAdminRole_gets_403_on_questions_endpoints(t *testing.T) {
	env := newTestEnv(t)
	testID := seedTest(t, env, "X", "math", "algebra", 60)
	qID := seedQuestion(t, env, testID, "essay", "explain", 1)

	studentID := seedUser(t, env, "student", "active", false)
	studentToken := authToken(t, env, studentID, "student")

	body := map[string]any{"format": "essay", "body": "x", "sort_order": 1}
	resp, out := doJSONBody(t, env, http.MethodPatch, "/api/v1/admin/questions/"+qID, body, studentToken)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "body=%v", out)
	assert.Equal(t, "forbidden", out["code"])
}