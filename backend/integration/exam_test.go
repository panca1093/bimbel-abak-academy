package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

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

func TestExam_AdminListTests_includes_question_count_per_row(t *testing.T) {
	env := newTestEnv(t)
	adminID := seedUser(t, env, "admin_exam", "active", false)
	token := authToken(t, env, adminID, "admin_exam")

	testID := seedTest(t, env, "Counted", "math", "algebra", 60)
	_ = seedQuestion(t, env, testID, "mcq", "q1", 1)
	_ = seedQuestion(t, env, testID, "mcq", "q2", 2)
	_ = seedQuestion(t, env, testID, "essay", "q3", 3)
	seedTest(t, env, "Empty", "math", "algebra", 60)

	resp, out := doJSONBody(t, env, http.MethodGet, "/api/v1/admin/tests?subject=math", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "body=%v", out)
	data := out["data"].([]any)
	require.Len(t, data, 2)

	countedCount := float64(-1)
	emptyCount := float64(-1)
	for _, raw := range data {
		row := raw.(map[string]any)
		qc, ok := row["question_count"].(float64)
		require.True(t, ok, "row must include question_count as number; row=%v", row)
		switch row["title"].(string) {
		case "Counted":
			countedCount = qc
		case "Empty":
			emptyCount = qc
		}
	}
	assert.Equal(t, float64(3), countedCount, "test with 3 questions should report question_count=3")
	assert.Equal(t, float64(0), emptyCount, "test with 0 questions should report question_count=0")
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

func TestExam_AdminCreateQuestion_mcq_with_zero_correct_returns_422(t *testing.T) {
	env := newTestEnv(t)
	adminID := seedUser(t, env, "admin_exam", "active", false)
	token := authToken(t, env, adminID, "admin_exam")

	testID := seedTest(t, env, "X", "math", "algebra", 60)
	body := map[string]any{
		"format":     "mcq",
		"body":       "pick one",
		"sort_order": 1,
		"options": []map[string]any{
			{"key": "a", "text": "1", "is_correct": false, "sort_order": 1},
			{"key": "b", "text": "2", "is_correct": false, "sort_order": 2},
		},
	}
	resp, out := doJSONBody(t, env, http.MethodPost, "/api/v1/admin/tests/"+testID+"/questions", body, token)
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode, "body=%v", out)
	assert.Equal(t, "validation_failed", out["code"])
}

func TestExam_AdminCreateQuestion_mcq_with_only_one_option_returns_422(t *testing.T) {
	env := newTestEnv(t)
	adminID := seedUser(t, env, "admin_exam", "active", false)
	token := authToken(t, env, adminID, "admin_exam")

	testID := seedTest(t, env, "X", "math", "algebra", 60)
	body := map[string]any{
		"format":     "mcq",
		"body":       "pick one",
		"sort_order": 1,
		"options": []map[string]any{
			{"key": "a", "text": "1", "is_correct": true, "sort_order": 1},
		},
	}
	resp, out := doJSONBody(t, env, http.MethodPost, "/api/v1/admin/tests/"+testID+"/questions", body, token)
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode, "body=%v", out)
	assert.Equal(t, "validation_failed", out["code"])
}

func TestExam_AdminCreateQuestion_multi_answer_with_two_correct_returns_201(t *testing.T) {
	env := newTestEnv(t)
	adminID := seedUser(t, env, "admin_exam", "active", false)
	token := authToken(t, env, adminID, "admin_exam")

	testID := seedTest(t, env, "X", "math", "algebra", 60)
	body := map[string]any{
		"format":     "multi_answer",
		"body":       "pick any",
		"sort_order": 1,
		"options": []map[string]any{
			{"key": "a", "text": "x", "is_correct": true, "sort_order": 1},
			{"key": "b", "text": "y", "is_correct": true, "sort_order": 2},
		},
	}
	resp, out := doJSONBody(t, env, http.MethodPost, "/api/v1/admin/tests/"+testID+"/questions", body, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "body=%v", out)
	q := out["question"].(map[string]any)
	assert.Equal(t, "multi_answer", q["format"])
	options := out["options"].([]any)
	assert.Len(t, options, 2)
}

func TestExam_AdminCreateQuestion_multi_answer_with_zero_correct_returns_422(t *testing.T) {
	env := newTestEnv(t)
	adminID := seedUser(t, env, "admin_exam", "active", false)
	token := authToken(t, env, adminID, "admin_exam")

	testID := seedTest(t, env, "X", "math", "algebra", 60)
	body := map[string]any{
		"format":     "multi_answer",
		"body":       "pick any",
		"sort_order": 1,
		"options": []map[string]any{
			{"key": "a", "text": "x", "is_correct": false, "sort_order": 1},
			{"key": "b", "text": "y", "is_correct": false, "sort_order": 2},
		},
	}
	resp, out := doJSONBody(t, env, http.MethodPost, "/api/v1/admin/tests/"+testID+"/questions", body, token)
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode, "body=%v", out)
	assert.Equal(t, "validation_failed", out["code"])
}

func TestExam_AdminCreateQuestion_short_with_correct_answer_returns_201(t *testing.T) {
	env := newTestEnv(t)
	adminID := seedUser(t, env, "admin_exam", "active", false)
	token := authToken(t, env, adminID, "admin_exam")

	testID := seedTest(t, env, "X", "math", "algebra", 60)
	body := map[string]any{
		"format":         "short",
		"body":           "capital of France",
		"sort_order":     1,
		"correct_answer": "Paris",
	}
	resp, out := doJSONBody(t, env, http.MethodPost, "/api/v1/admin/tests/"+testID+"/questions", body, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "body=%v", out)
	q := out["question"].(map[string]any)
	assert.Equal(t, "short", q["format"])
	assert.Equal(t, "Paris", q["correct_answer"])
}

func TestExam_AdminCreateQuestion_fill_blank_with_correct_answer_returns_201(t *testing.T) {
	env := newTestEnv(t)
	adminID := seedUser(t, env, "admin_exam", "active", false)
	token := authToken(t, env, adminID, "admin_exam")

	testID := seedTest(t, env, "X", "math", "algebra", 60)
	body := map[string]any{
		"format":         "fill_blank",
		"body":           "2 + 2 = ___",
		"sort_order":     1,
		"correct_answer": "4",
	}
	resp, out := doJSONBody(t, env, http.MethodPost, "/api/v1/admin/tests/"+testID+"/questions", body, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "body=%v", out)
	q := out["question"].(map[string]any)
	assert.Equal(t, "fill_blank", q["format"])
	assert.Equal(t, "4", q["correct_answer"])
}

func TestExam_AdminCreateQuestion_fill_blank_without_correct_answer_returns_422(t *testing.T) {
	env := newTestEnv(t)
	adminID := seedUser(t, env, "admin_exam", "active", false)
	token := authToken(t, env, adminID, "admin_exam")

	testID := seedTest(t, env, "X", "math", "algebra", 60)
	body := map[string]any{
		"format":     "fill_blank",
		"body":       "2 + 2 = ___",
		"sort_order": 1,
	}
	resp, out := doJSONBody(t, env, http.MethodPost, "/api/v1/admin/tests/"+testID+"/questions", body, token)
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode, "body=%v", out)
	assert.Equal(t, "validation_failed", out["code"])
}

func TestExam_AdminCreateQuestion_essay_with_options_returns_422(t *testing.T) {
	env := newTestEnv(t)
	adminID := seedUser(t, env, "admin_exam", "active", false)
	token := authToken(t, env, adminID, "admin_exam")

	testID := seedTest(t, env, "X", "math", "algebra", 60)
	body := map[string]any{
		"format":     "essay",
		"body":       "explain",
		"sort_order": 1,
		"options": []map[string]any{
			{"key": "a", "text": "nope", "is_correct": false, "sort_order": 1},
		},
	}
	resp, out := doJSONBody(t, env, http.MethodPost, "/api/v1/admin/tests/"+testID+"/questions", body, token)
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode, "body=%v", out)
	assert.Equal(t, "validation_failed", out["code"])
}

func TestExam_AdminPatchQuestion_nonexistent_id_returns_404(t *testing.T) {
	env := newTestEnv(t)
	adminID := seedUser(t, env, "admin_exam", "active", false)
	token := authToken(t, env, adminID, "admin_exam")

	body := map[string]any{
		"format":     "essay",
		"body":       "doesn't matter",
		"sort_order": 1,
	}
	// Use a non-zero UUID so the service takes the UPDATE branch (not the CREATE-on-uuid.Nil branch
	// which would FK-violate on question.test_id). This exercises the UpdateQuestionTx
	// not-found path → ErrQuestionNotFound → 404.
	fakeQuestionID := "11111111-1111-1111-1111-111111111111"
	resp, out := doJSONBody(t, env, http.MethodPatch, "/api/v1/admin/questions/"+fakeQuestionID, body, token)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode, "body=%v", out)
	assert.Equal(t, "question_not_found", out["code"])
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

// ----- Exam CRUD (Slice 2) -----

func TestExam_AdminCreateExam_creates_linked_product_draft_zero_price(t *testing.T) {
	env := newTestEnv(t)
	adminID := seedUser(t, env, "admin_exam", "active", false)
	token := authToken(t, env, adminID, "admin_exam")

	body := map[string]any{
		"title":             "Tryout Akbar",
		"timer_mode":        "per_question",
		"is_free":           true,
		"requires_checkin":  false,
		"allow_leaderboard": true,
		"randomize":         false,
	}
	resp, out := doJSONBody(t, env, http.MethodPost, "/api/v1/admin/exams", body, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "body=%v", out)

	exam, ok := out["exam"].(map[string]any)
	require.True(t, ok, "expected exam object, got %T", out["exam"])
	assert.NotEmpty(t, exam["id"])

	product, ok := out["product"].(map[string]any)
	require.True(t, ok, "expected product object, got %T", out["product"])
	productID, ok := product["id"].(string)
	require.True(t, ok, "expected product.id string, got %T", product["id"])
	require.NotEmpty(t, productID)
	assert.Equal(t, "exam", product["type"])
	assert.Equal(t, "draft", product["status"])
	assert.Equal(t, float64(0), product["price"])

	ctx := context.Background()
	var dbType, dbStatus string
	var dbPrice int64
	err := env.pool.QueryRow(ctx,
		`SELECT type, status, price FROM product WHERE id = $1`, productID,
	).Scan(&dbType, &dbStatus, &dbPrice)
	require.NoError(t, err)
	assert.Equal(t, "exam", dbType)
	assert.Equal(t, "draft", dbStatus)
	assert.Equal(t, int64(0), dbPrice)

	var dbExamProductID string
	err = env.pool.QueryRow(ctx,
		`SELECT product_id FROM exam WHERE id = $1`, exam["id"].(string),
	).Scan(&dbExamProductID)
	require.NoError(t, err)
	assert.Equal(t, productID, dbExamProductID, "exam.product_id in DB must match product.id from response")
}

func TestExam_AdminListExams_returns_data_and_cursor(t *testing.T) {
	env := newTestEnv(t)
	adminID := seedUser(t, env, "admin_exam", "active", false)
	token := authToken(t, env, adminID, "admin_exam")

	createdIDs := map[string]bool{}
	for i := 0; i < 3; i++ {
		body := map[string]any{
			"title":      fmt.Sprintf("Paket %d", i),
			"timer_mode": "per_question",
			"is_free":    true,
		}
		resp, out := doJSONBody(t, env, http.MethodPost, "/api/v1/admin/exams", body, token)
		require.Equal(t, http.StatusCreated, resp.StatusCode, "body=%v", out)
		exam := out["exam"].(map[string]any)
		createdIDs[exam["id"].(string)] = true
	}

	resp, out := doJSONBody(t, env, http.MethodGet, "/api/v1/admin/exams", nil, token)
	require.Equal(t, http.StatusOK, resp.StatusCode, "body=%v", out)
	data := out["data"].([]any)
	assert.GreaterOrEqual(t, len(data), 3, "list should include all 3 created exams")
	for _, raw := range data {
		row := raw.(map[string]any)
		assert.True(t, createdIDs[row["id"].(string)], "listed exam %v should be one of the 3 we created", row["id"])
	}
}

func TestExam_AdminGetExam_detail_with_and_without_tests(t *testing.T) {
	env := newTestEnv(t)
	adminID := seedUser(t, env, "admin_exam", "active", false)
	token := authToken(t, env, adminID, "admin_exam")

	emptyBody := map[string]any{"title": "Empty Paket", "timer_mode": "per_question", "is_free": true}
	emptyResp, emptyOut := doJSONBody(t, env, http.MethodPost, "/api/v1/admin/exams", emptyBody, token)
	require.Equal(t, http.StatusCreated, emptyResp.StatusCode)
	emptyExamID := emptyOut["exam"].(map[string]any)["id"].(string)

	emptyGet, emptyDetail := doJSONBody(t, env, http.MethodGet, "/api/v1/admin/exams/"+emptyExamID, nil, token)
	require.Equal(t, http.StatusOK, emptyGet.StatusCode, "body=%v", emptyDetail)
	assert.Equal(t, float64(0), emptyDetail["product_price"])
	assert.Equal(t, "draft", emptyDetail["product_status"])
	testsAny, ok := emptyDetail["tests"].([]any)
	require.True(t, ok, "tests should be array, got %T", emptyDetail["tests"])
	assert.Len(t, testsAny, 0)
	assert.Equal(t, "Empty Paket", emptyDetail["title"])

	withTestResp, withTestOut := doJSONBody(t, env, http.MethodPost, "/api/v1/admin/exams",
		map[string]any{"title": "Full Paket", "timer_mode": "per_question", "is_free": true}, token)
	require.Equal(t, http.StatusCreated, withTestResp.StatusCode)
	withExamID := withTestOut["exam"].(map[string]any)["id"].(string)

	testB := seedTest(t, env, "B-Test", "math", "algebra", 30)
	testA := seedTest(t, env, "A-Test", "math", "algebra", 30)
	_ = seedQuestion(t, env, testA, "essay", "a", 1)

	putBody := []string{testB, testA}
	putResp := env.doJSON(t, http.MethodPut, "/api/v1/admin/exams/"+withExamID+"/tests", putBody, token)
	require.Equal(t, http.StatusNoContent, putResp.StatusCode)
	putResp.Body.Close()

	detailResp, detail := doJSONBody(t, env, http.MethodGet, "/api/v1/admin/exams/"+withExamID, nil, token)
	require.Equal(t, http.StatusOK, detailResp.StatusCode, "body=%v", detail)
	detailTests := detail["tests"].([]any)
	require.Len(t, detailTests, 2)
	first := detailTests[0].(map[string]any)["test"].(map[string]any)
	second := detailTests[1].(map[string]any)["test"].(map[string]any)
	assert.Equal(t, testB, first["id"], "sort_order 0 should be testB")
	assert.Equal(t, testA, second["id"], "sort_order 1 should be testA")
	assert.Equal(t, float64(0), first["question_count"], "testB has no seeded questions")
	assert.Equal(t, float64(1), second["question_count"], "testA has 1 seeded question")
}

func TestExam_AdminUpdateExam_overlays_fields(t *testing.T) {
	env := newTestEnv(t)
	adminID := seedUser(t, env, "admin_exam", "active", false)
	token := authToken(t, env, adminID, "admin_exam")

	createResp, createOut := doJSONBody(t, env, http.MethodPost, "/api/v1/admin/exams",
		map[string]any{"title": "Old", "timer_mode": "per_question", "is_free": true}, token)
	require.Equal(t, http.StatusCreated, createResp.StatusCode)
	examID := createOut["exam"].(map[string]any)["id"].(string)

	_, originalDetail := doJSONBody(t, env, http.MethodGet, "/api/v1/admin/exams/"+examID, nil, token)
	originalProductID := originalDetail["product_id"].(string)
	require.NotEmpty(t, originalProductID)

	scheduled := time.Date(2030, 1, 15, 10, 0, 0, 0, time.UTC)
	patchBody := map[string]any{
		"title":            "New Title",
		"scheduled_at":     scheduled.Format(time.RFC3339),
		"timer_mode":       "overall",
		"duration_minutes": 60,
	}
	patchResp, patchOut := doJSONBody(t, env, http.MethodPatch, "/api/v1/admin/exams/"+examID, patchBody, token)
	require.Equal(t, http.StatusOK, patchResp.StatusCode, "body=%v", patchOut)
	assert.Equal(t, "New Title", patchOut["title"])
	assert.Equal(t, "overall", patchOut["timer_mode"])
	assert.Equal(t, float64(60), patchOut["duration_minutes"])

	getResp, getOut := doJSONBody(t, env, http.MethodGet, "/api/v1/admin/exams/"+examID, nil, token)
	require.Equal(t, http.StatusOK, getResp.StatusCode, "body=%v", getOut)
	assert.Equal(t, "New Title", getOut["title"])
	assert.Equal(t, "overall", getOut["timer_mode"])
	assert.Equal(t, float64(60), getOut["duration_minutes"])
	assert.Equal(t, originalProductID, getOut["product_id"], "product_id must not change on PATCH")
}

func TestExam_AdminReplaceExamTests_declarative_and_bad_id_leaves_rows_intact(t *testing.T) {
	env := newTestEnv(t)
	adminID := seedUser(t, env, "admin_exam", "active", false)
	token := authToken(t, env, adminID, "admin_exam")

	createResp, createOut := doJSONBody(t, env, http.MethodPost, "/api/v1/admin/exams",
		map[string]any{"title": "Replaceable", "timer_mode": "per_question", "is_free": true}, token)
	require.Equal(t, http.StatusCreated, createResp.StatusCode)
	examID := createOut["exam"].(map[string]any)["id"].(string)

	t1 := seedTest(t, env, "T1", "math", "algebra", 30)
	t2 := seedTest(t, env, "T2", "math", "algebra", 30)
	t3 := seedTest(t, env, "T3", "math", "algebra", 30)
	fakeID := "11111111-2222-3333-4444-555555555555"

	putResp := env.doJSON(t, http.MethodPut, "/api/v1/admin/exams/"+examID+"/tests",
		[]string{t1, t2}, token)
	require.Equal(t, http.StatusNoContent, putResp.StatusCode)
	putResp.Body.Close()

	detailResp, detail := doJSONBody(t, env, http.MethodGet, "/api/v1/admin/exams/"+examID, nil, token)
	require.Equal(t, http.StatusOK, detailResp.StatusCode, "body=%v", detail)
	cur := detail["tests"].([]any)
	require.Len(t, cur, 2)

	putResp2 := env.doJSON(t, http.MethodPut, "/api/v1/admin/exams/"+examID+"/tests",
		[]string{t3}, token)
	require.Equal(t, http.StatusNoContent, putResp2.StatusCode)
	putResp2.Body.Close()

	_, detail2 := doJSONBody(t, env, http.MethodGet, "/api/v1/admin/exams/"+examID, nil, token)
	cur2 := detail2["tests"].([]any)
	require.Len(t, cur2, 1, "after replace down to 1 test_id, list should be exactly 1")
	assert.Equal(t, t3, cur2[0].(map[string]any)["test"].(map[string]any)["id"])

	badResp, badOut := doJSONBody(t, env, http.MethodPut, "/api/v1/admin/exams/"+examID+"/tests",
		[]string{fakeID}, token)
	assert.True(t, badResp.StatusCode == http.StatusBadRequest || badResp.StatusCode == http.StatusNotFound,
		"want 400 or 404, got %d body=%v", badResp.StatusCode, badOut)

	_, detailAfter := doJSONBody(t, env, http.MethodGet, "/api/v1/admin/exams/"+examID, nil, token)
	curAfter := detailAfter["tests"].([]any)
	require.Len(t, curAfter, 1, "bad test_id must not mutate exam_test rows")
	assert.Equal(t, t3, curAfter[0].(map[string]any)["test"].(map[string]any)["id"])
}

func TestExam_AdminUpdateExamPrice_updates_linked_product(t *testing.T) {
	env := newTestEnv(t)
	adminID := seedUser(t, env, "admin_exam", "active", false)
	token := authToken(t, env, adminID, "admin_exam")

	createResp, createOut := doJSONBody(t, env, http.MethodPost, "/api/v1/admin/exams",
		map[string]any{"title": "Pricing Paket", "timer_mode": "per_question", "is_free": false}, token)
	require.Equal(t, http.StatusCreated, createResp.StatusCode)
	examID := createOut["exam"].(map[string]any)["id"].(string)
	productID := createOut["product"].(map[string]any)["id"].(string)

	priceResp, priceOut := doJSONBody(t, env, http.MethodPatch, "/api/v1/admin/exams/"+examID+"/price",
		map[string]any{"price": 50000}, token)
	require.Equal(t, http.StatusOK, priceResp.StatusCode, "body=%v", priceOut)
	assert.Equal(t, float64(50000), priceOut["price"])

	getResp, detail := doJSONBody(t, env, http.MethodGet, "/api/v1/admin/exams/"+examID, nil, token)
	require.Equal(t, http.StatusOK, getResp.StatusCode, "body=%v", detail)
	assert.Equal(t, float64(50000), detail["product_price"], "ExamDetail must reflect new price")

	ctx := context.Background()
	var dbPrice int64
	err := env.pool.QueryRow(ctx, `SELECT price FROM product WHERE id = $1`, productID).Scan(&dbPrice)
	require.NoError(t, err)
	assert.Equal(t, int64(50000), dbPrice, "product.price in DB must equal 50000")
}

func TestExam_AdminPublishExam_marks_product_published(t *testing.T) {
	env := newTestEnv(t)
	adminID := seedUser(t, env, "admin_exam", "active", false)
	token := authToken(t, env, adminID, "admin_exam")

	createResp, createOut := doJSONBody(t, env, http.MethodPost, "/api/v1/admin/exams",
		map[string]any{"title": "Publishable", "timer_mode": "overall", "duration_minutes": 90, "is_free": false}, token)
	require.Equal(t, http.StatusCreated, createResp.StatusCode)
	examID := createOut["exam"].(map[string]any)["id"].(string)
	productID := createOut["product"].(map[string]any)["id"].(string)

	priceResp, _ := doJSONBody(t, env, http.MethodPatch, "/api/v1/admin/exams/"+examID+"/price",
		map[string]any{"price": 25000}, token)
	require.Equal(t, http.StatusOK, priceResp.StatusCode)

	pubResp, pubOut := doJSONBody(t, env, http.MethodPost, "/api/v1/admin/exams/"+examID+"/publish",
		nil, token)
	require.Equal(t, http.StatusOK, pubResp.StatusCode, "body=%v", pubOut)

	getResp, detail := doJSONBody(t, env, http.MethodGet, "/api/v1/admin/exams/"+examID, nil, token)
	require.Equal(t, http.StatusOK, getResp.StatusCode, "body=%v", detail)
	assert.Equal(t, "published", detail["product_status"], "ExamDetail.product_status should be 'published'")

	ctx := context.Background()
	var dbStatus string
	err := env.pool.QueryRow(ctx, `SELECT status FROM product WHERE id = $1`, productID).Scan(&dbStatus)
	require.NoError(t, err)
	assert.Equal(t, "published", dbStatus, "product.status in DB must equal 'published'")
}