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

// jsonNum reads a numeric field from a decoded JSON response map, treating a missing key
// as zero. Some struct fields (e.g. model.SessionResult's score/correct_count/empty_count)
// carry `omitempty`, so a legitimately-zero value is absent from the wire payload rather
// than serialized as JSON `0`.
func jsonNum(v any) float64 {
	if v == nil {
		return 0
	}
	return v.(float64)
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

func TestExam_AdminCreateQuestion_roundtrips_nondefault_points(t *testing.T) {
	env := newTestEnv(t)
	adminID := seedUser(t, env, "admin_exam", "active", false)
	token := authToken(t, env, adminID, "admin_exam")

	testID := seedTest(t, env, "X", "math", "algebra", 60)
	body := map[string]any{
		"format":        "essay",
		"body":          "explain gravity",
		"sort_order":    1,
		"point_correct": 5,
		"point_wrong":   3,
	}
	resp, out := doJSONBody(t, env, http.MethodPost, "/api/v1/admin/tests/"+testID+"/questions", body, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "body=%v", out)
	q := out["question"].(map[string]any)
	assert.Equal(t, float64(5), q["point_correct"])
	assert.Equal(t, float64(3), q["point_wrong"])

	resp2, out2 := doJSONBody(t, env, http.MethodGet, "/api/v1/admin/tests/"+testID+"/questions", nil, token)
	require.Equal(t, http.StatusOK, resp2.StatusCode, "body=%v", out2)
	data := out2["data"].([]any)
	require.Len(t, data, 1)
	readQ := data[0].(map[string]any)["question"].(map[string]any)
	assert.Equal(t, float64(5), readQ["point_correct"])
	assert.Equal(t, float64(3), readQ["point_wrong"])
}

func TestExam_AdminCreateQuestion_defaults_points_when_omitted(t *testing.T) {
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
	assert.Equal(t, float64(1), q["point_correct"])
	assert.Equal(t, float64(0), q["point_wrong"])
}

func TestExam_AdminCreateQuestion_point_correct_below_1_returns_422(t *testing.T) {
	env := newTestEnv(t)
	adminID := seedUser(t, env, "admin_exam", "active", false)
	token := authToken(t, env, adminID, "admin_exam")

	testID := seedTest(t, env, "X", "math", "algebra", 60)
	body := map[string]any{
		"format":        "essay",
		"body":          "explain gravity",
		"sort_order":    1,
		"point_correct": 0,
	}
	resp, out := doJSONBody(t, env, http.MethodPost, "/api/v1/admin/tests/"+testID+"/questions", body, token)
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode, "body=%v", out)
	assert.Equal(t, "validation_failed", out["code"])
}

func TestExam_AdminCreateQuestion_negative_point_wrong_returns_422(t *testing.T) {
	env := newTestEnv(t)
	adminID := seedUser(t, env, "admin_exam", "active", false)
	token := authToken(t, env, adminID, "admin_exam")

	testID := seedTest(t, env, "X", "math", "algebra", 60)
	body := map[string]any{
		"format":      "essay",
		"body":        "explain gravity",
		"sort_order":  1,
		"point_wrong": -1,
	}
	resp, out := doJSONBody(t, env, http.MethodPost, "/api/v1/admin/tests/"+testID+"/questions", body, token)
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode, "body=%v", out)
	assert.Equal(t, "validation_failed", out["code"])
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

// ----- Slice 5 (Task 8): result endpoint + essay grading, end-to-end -----
//
// AdminCreateExam is a known, pre-existing, out-of-scope defect on this branch (empty
// req.ResultConfig binds to "" and violates the migration-0015 CHECK(result_config IN (...))
// constraint on INSERT, 500ing every one of the 7 TestExam_Admin{CreateExam,ListExams,GetExam,
// UpdateExam,ReplaceExamTests,UpdateExamPrice,PublishExam}_* tests above). The tests below seed
// `exam`/`exam_test`/`exam_registration` directly via SQL (same pattern seedTest/seedQuestion
// already use) to sidestep that defect entirely and exercise only the Slice 5 student session +
// result + admin grading endpoints, which do not go through AdminCreateExam.

// seedQuestionWithPoints inserts a question row with explicit point_correct/point_wrong
// (seedQuestion always takes the column defaults of 1/0) and returns its ID.
func seedQuestionWithPoints(t *testing.T, env *testEnv, testID, format, body string, sortOrder, pointCorrect, pointWrong int) string {
	t.Helper()
	ctx := context.Background()
	var id string
	err := env.pool.QueryRow(ctx,
		`INSERT INTO question (test_id, format, body, sort_order, point_correct, point_wrong)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		testID, format, body, sortOrder, pointCorrect, pointWrong,
	).Scan(&id)
	require.NoError(t, err)
	return id
}

// seedMCQOptions inserts two options ("a", "b") for a mcq question; correctKey marks which
// option key is_correct=true.
func seedMCQOptions(t *testing.T, env *testEnv, questionID, correctKey string) {
	t.Helper()
	ctx := context.Background()
	_, err := env.pool.Exec(ctx,
		`INSERT INTO question_option (question_id, key, text, is_correct, sort_order)
		 VALUES ($1, 'a', 'Option A', $2, 1), ($1, 'b', 'Option B', $3, 2)`,
		questionID, correctKey == "a", correctKey == "b",
	)
	require.NoError(t, err)
}

// seedExam inserts an exam row directly via SQL (bypassing the broken AdminCreateExam —
// see comment above), giving the caller full control over result_config/result_release_at.
func seedExam(t *testing.T, env *testEnv, title, resultConfig string, resultReleaseAt *time.Time) string {
	t.Helper()
	ctx := context.Background()
	var id string
	err := env.pool.QueryRow(ctx,
		`INSERT INTO exam (title, is_free, requires_checkin, timer_mode, result_config, result_release_at)
		 VALUES ($1, true, false, 'per_question', $2, $3) RETURNING id`,
		title, resultConfig, resultReleaseAt,
	).Scan(&id)
	require.NoError(t, err)
	return id
}

// linkExamTest inserts an exam_test join row.
func linkExamTest(t *testing.T, env *testEnv, examID, testID string, sortOrder int) {
	t.Helper()
	ctx := context.Background()
	_, err := env.pool.Exec(ctx,
		`INSERT INTO exam_test (exam_id, test_id, sort_order) VALUES ($1, $2, $3)`,
		examID, testID, sortOrder,
	)
	require.NoError(t, err)
}

// seedExamRegistration inserts an exam_registration row for a student, returns registrationID.
func seedExamRegistration(t *testing.T, env *testEnv, studentID, examID string) string {
	t.Helper()
	ctx := context.Background()
	var id string
	token := fmt.Sprintf("tok-%d", time.Now().UnixNano())
	err := env.pool.QueryRow(ctx,
		`INSERT INTO exam_registration (student_id, exam_id, token) VALUES ($1, $2, $3) RETURNING id`,
		studentID, examID, token,
	).Scan(&id)
	require.NoError(t, err)
	return id
}

// startSubmitSession drives the student session flow over HTTP (start -> save answers ->
// submit) and returns the sessionID.
func startSubmitSession(t *testing.T, env *testEnv, studentToken, registrationID string, answers []map[string]any) string {
	t.Helper()
	resp, out := doJSONBody(t, env, http.MethodPost, "/api/v1/exam/sessions",
		map[string]any{"registration_id": registrationID}, studentToken)
	require.Equal(t, http.StatusOK, resp.StatusCode, "start session body=%v", out)
	sessionID, ok := out["session_id"].(string)
	require.True(t, ok, "expected session_id string, got %T", out["session_id"])

	saveResp, saveOut := doJSONBody(t, env, http.MethodPatch, "/api/v1/exam/sessions/"+sessionID+"/answers",
		map[string]any{"answers": answers}, studentToken)
	require.Equal(t, http.StatusOK, saveResp.StatusCode, "save answers body=%v", saveOut)

	submitResp, submitOut := doJSONBody(t, env, http.MethodPost, "/api/v1/exam/sessions/"+sessionID+"/submit", nil, studentToken)
	require.Equal(t, http.StatusOK, submitResp.StatusCode, "submit body=%v", submitOut)
	require.Equal(t, "submitted", submitOut["status"])

	return sessionID
}

// TestExam_EssayFlow_EndToEnd_GradingAndResultGating drives the full student+admin flow for
// a session with one wrong mcq and one ungraded essay: register/start/submit (asserting the
// essay's graded_at stays NULL post-submit — PG note 1 regression), the admin grading queue +
// essays read, grading the essay (asserting the session total is recomputed by folding raw
// persisted per-answer scores and re-clamped at 0, not just added on top of the already-
// clamped submit-time total), the result endpoint's grading->result gate transition, the
// score_pembahasan breakdown/pembahasan visibility, and the ownership 404.
func TestExam_EssayFlow_EndToEnd_GradingAndResultGating(t *testing.T) {
	env := newTestEnv(t)

	adminID := seedUser(t, env, "admin_exam", "active", false)
	adminToken := authToken(t, env, adminID, "admin_exam")
	studentID := seedUser(t, env, "student", "active", false)
	studentToken := authToken(t, env, studentID, "student")

	testID := seedTest(t, env, "Essay Mix", "bahasa", "menulis", 60)
	mcqQID := seedQuestionWithPoints(t, env, testID, "mcq", "2+2", 1, 1, 8)
	seedMCQOptions(t, env, mcqQID, "a")
	essayQID := seedQuestionWithPoints(t, env, testID, "essay", "Explain gravity", 2, 5, 0)

	examID := seedExam(t, env, "Essay Paket", "score_pembahasan", nil)
	linkExamTest(t, env, examID, testID, 0)
	regID := seedExamRegistration(t, env, studentID, examID)

	sessionID := startSubmitSession(t, env, studentToken, regID, []map[string]any{
		{"question_id": mcqQID, "answer": "b"}, // wrong: correct key is "a"
		{"question_id": essayQID, "answer": "my essay answer text"},
	})

	ctx := context.Background()

	t.Run("submit leaves the essay ungraded (PG note 1 fix) and grades the mcq", func(t *testing.T) {
		var essayAnswer string
		var essayGradedAt *time.Time
		err := env.pool.QueryRow(ctx,
			`SELECT answer, graded_at FROM exam_session_answer WHERE session_id = $1 AND question_id = $2`,
			sessionID, essayQID,
		).Scan(&essayAnswer, &essayGradedAt)
		require.NoError(t, err)
		assert.Equal(t, "my essay answer text", essayAnswer)
		assert.Nil(t, essayGradedAt, "essay graded_at must stay NULL right after submit")

		var mcqScore float64
		var mcqCorrect bool
		var mcqGradedAt *time.Time
		err = env.pool.QueryRow(ctx,
			`SELECT score, is_correct, graded_at FROM exam_session_answer WHERE session_id = $1 AND question_id = $2`,
			sessionID, mcqQID,
		).Scan(&mcqScore, &mcqCorrect, &mcqGradedAt)
		require.NoError(t, err)
		assert.Equal(t, -8.0, mcqScore, "wrong mcq stores the raw negative point_wrong magnitude, unclamped")
		assert.False(t, mcqCorrect)
		assert.NotNil(t, mcqGradedAt, "objective answers are graded at submit time")
	})

	t.Run("result is gated to grading while the essay is ungraded", func(t *testing.T) {
		resp, out := doJSONBody(t, env, http.MethodGet, "/api/v1/exam/sessions/"+sessionID+"/result", nil, studentToken)
		require.Equal(t, http.StatusOK, resp.StatusCode, "body=%v", out)
		assert.Equal(t, "grading", out["state"])
	})

	t.Run("admin grading queue lists the session with its ungraded essay count", func(t *testing.T) {
		resp, out := doJSONBody(t, env, http.MethodGet, "/api/v1/admin/exams/"+examID+"/grading", nil, adminToken)
		require.Equal(t, http.StatusOK, resp.StatusCode, "body=%v", out)
		data, ok := out["data"].([]any)
		require.True(t, ok, "expected data array, got %T", out["data"])
		require.Len(t, data, 1)
		row := data[0].(map[string]any)
		assert.Equal(t, sessionID, row["session_id"])
		assert.Equal(t, studentID, row["student_id"])
		assert.Equal(t, float64(1), row["ungraded_essay_count"])
	})

	t.Run("admin session essays read returns the ungraded essay with question metadata", func(t *testing.T) {
		resp, out := doJSONBody(t, env, http.MethodGet, "/api/v1/admin/sessions/"+sessionID+"/essays", nil, adminToken)
		require.Equal(t, http.StatusOK, resp.StatusCode, "body=%v", out)
		data, ok := out["data"].([]any)
		require.True(t, ok, "expected data array, got %T", out["data"])
		require.Len(t, data, 1)
		row := data[0].(map[string]any)
		assert.Equal(t, essayQID, row["question_id"])
		assert.Equal(t, "Explain gravity", row["body"])
		assert.Equal(t, "my essay answer text", row["answer"])
		assert.Equal(t, float64(5), row["point_correct"])
		assert.Nil(t, row["score"])
		assert.Nil(t, row["graded_at"])
	})

	t.Run("grading the essay recomputes the session total by folding raw scores and re-clamps", func(t *testing.T) {
		resp, out := doJSONBody(t, env, http.MethodPost, "/api/v1/admin/sessions/"+sessionID+"/grade",
			map[string]any{"question_id": essayQID, "score": 2, "comment": "Needs more detail"}, adminToken)
		require.Equal(t, http.StatusOK, resp.StatusCode, "body=%v", out)
		assert.Equal(t, "ok", out["status"])
		// Raw persisted scores are mcq=-8 and essay=2 -> fold = -6 -> clamp(0, -6) = 0.
		// A buggy "add on top of the already-clamped submit total" (0 + 2 = 2) would fail this.
		assert.Equal(t, float64(0), out["score"], "recompute must fold raw persisted scores, not add onto the clamped submit total")

		var gradedAt *time.Time
		var score float64
		var comment string
		var gradedBy string
		err := env.pool.QueryRow(ctx,
			`SELECT graded_at, score, grader_comment, graded_by FROM exam_session_answer WHERE session_id = $1 AND question_id = $2`,
			sessionID, essayQID,
		).Scan(&gradedAt, &score, &comment, &gradedBy)
		require.NoError(t, err)
		assert.NotNil(t, gradedAt, "essay graded_at must be set after grading")
		assert.Equal(t, 2.0, score)
		assert.Equal(t, "Needs more detail", comment)
		assert.Equal(t, adminID, gradedBy)

		var sessScore float64
		err = env.pool.QueryRow(ctx, `SELECT score FROM exam_session WHERE id = $1`, sessionID).Scan(&sessScore)
		require.NoError(t, err)
		assert.Equal(t, 0.0, sessScore, "exam_session.score must persist the recomputed clamped total")
	})

	t.Run("result transitions to result state with score_pembahasan breakdown and pembahasan", func(t *testing.T) {
		resp, out := doJSONBody(t, env, http.MethodGet, "/api/v1/exam/sessions/"+sessionID+"/result", nil, studentToken)
		require.Equal(t, http.StatusOK, resp.StatusCode, "body=%v", out)
		assert.Equal(t, "result", out["state"])
		assert.Equal(t, "score_pembahasan", out["result_config"])
		// score/correct_count/empty_count carry `omitempty` on model.SessionResult, so a
		// legitimate zero value is absent from the wire payload rather than JSON `0`.
		assert.Equal(t, float64(0), jsonNum(out["score"]))
		assert.Equal(t, float64(0), jsonNum(out["correct_count"]))
		assert.Equal(t, float64(1), out["wrong_count"])
		assert.Equal(t, float64(0), jsonNum(out["empty_count"]))
		assert.Equal(t, float64(1), out["rank"], "only fully-graded submitted session for this exam")

		breakdown, ok := out["breakdown"].([]any)
		require.True(t, ok, "expected breakdown array, got %T", out["breakdown"])
		require.Len(t, breakdown, 1)
		row := breakdown[0].(map[string]any)
		assert.Equal(t, testID, row["test_id"])
		assert.Equal(t, -6.0, row["earned"], "breakdown earned is the raw unclamped fold (mcq -8 + essay 2)")
		assert.Equal(t, float64(6), row["max"], "max sums point_correct across the test (mcq 1 + essay 5)")

		pembahasan, ok := out["pembahasan"].([]any)
		require.True(t, ok, "expected pembahasan array, got %T", out["pembahasan"])
		require.Len(t, pembahasan, 1, "essay is excluded from pembahasan, only the mcq remains")
		pRow := pembahasan[0].(map[string]any)
		assert.Equal(t, mcqQID, pRow["question_id"])
		assert.Equal(t, "b", pRow["your_answer"])
		assert.Equal(t, "a", pRow["correct_answer"])
		assert.Equal(t, false, pRow["is_correct"])
	})

	t.Run("result is 404 session_not_found for a different student", func(t *testing.T) {
		otherStudentID := seedUser(t, env, "student", "active", false)
		otherToken := authToken(t, env, otherStudentID, "student")

		resp, out := doJSONBody(t, env, http.MethodGet, "/api/v1/exam/sessions/"+sessionID+"/result", nil, otherToken)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode, "body=%v", out)
		assert.Equal(t, "session_not_found", out["code"])
	})
}

// TestExam_Result_ScoreOnly_HidesBreakdownAndPembahasan covers FR-S5-22: score_only never
// includes breakdown/pembahasan, only score/counts/rank.
func TestExam_Result_ScoreOnly_HidesBreakdownAndPembahasan(t *testing.T) {
	env := newTestEnv(t)

	studentID := seedUser(t, env, "student", "active", false)
	studentToken := authToken(t, env, studentID, "student")

	testID := seedTest(t, env, "Score Only", "math", "aljabar", 30)
	mcqQID := seedQuestionWithPoints(t, env, testID, "mcq", "2+2", 1, 3, 1)
	seedMCQOptions(t, env, mcqQID, "a")

	examID := seedExam(t, env, "Score Only Paket", "score_only", nil)
	linkExamTest(t, env, examID, testID, 0)
	regID := seedExamRegistration(t, env, studentID, examID)

	sessionID := startSubmitSession(t, env, studentToken, regID, []map[string]any{
		{"question_id": mcqQID, "answer": "a"}, // correct
	})

	resp, out := doJSONBody(t, env, http.MethodGet, "/api/v1/exam/sessions/"+sessionID+"/result", nil, studentToken)
	require.Equal(t, http.StatusOK, resp.StatusCode, "body=%v", out)
	assert.Equal(t, "result", out["state"])
	assert.Equal(t, "score_only", out["result_config"])
	assert.Equal(t, float64(3), out["score"])
	assert.Equal(t, float64(1), out["correct_count"])
	assert.Equal(t, float64(1), out["rank"])

	_, hasBreakdown := out["breakdown"]
	assert.False(t, hasBreakdown, "score_only must not include a breakdown key")
	_, hasPembahasan := out["pembahasan"]
	assert.False(t, hasPembahasan, "score_only must not include a pembahasan key")
}

// TestExam_Result_HiddenAndLocked_Gates covers FR-S5-21 gates 1 and 3: hidden always wins,
// and a future result_release_at locks an otherwise-visible result.
func TestExam_Result_HiddenAndLocked_Gates(t *testing.T) {
	env := newTestEnv(t)

	testID := seedTest(t, env, "Gates", "math", "aljabar", 30)
	mcqQID := seedQuestionWithPoints(t, env, testID, "mcq", "2+2", 1, 1, 0)
	seedMCQOptions(t, env, mcqQID, "a")

	t.Run("hidden", func(t *testing.T) {
		studentID := seedUser(t, env, "student", "active", false)
		studentToken := authToken(t, env, studentID, "student")

		examID := seedExam(t, env, "Hidden Paket", "hidden", nil)
		linkExamTest(t, env, examID, testID, 0)
		regID := seedExamRegistration(t, env, studentID, examID)

		sessionID := startSubmitSession(t, env, studentToken, regID, []map[string]any{
			{"question_id": mcqQID, "answer": "a"},
		})

		resp, out := doJSONBody(t, env, http.MethodGet, "/api/v1/exam/sessions/"+sessionID+"/result", nil, studentToken)
		require.Equal(t, http.StatusOK, resp.StatusCode, "body=%v", out)
		assert.Equal(t, "hidden", out["state"])
	})

	t.Run("locked", func(t *testing.T) {
		studentID := seedUser(t, env, "student", "active", false)
		studentToken := authToken(t, env, studentID, "student")

		releaseAt := time.Now().Add(24 * time.Hour)
		examID := seedExam(t, env, "Locked Paket", "score_only", &releaseAt)
		linkExamTest(t, env, examID, testID, 0)
		regID := seedExamRegistration(t, env, studentID, examID)

		sessionID := startSubmitSession(t, env, studentToken, regID, []map[string]any{
			{"question_id": mcqQID, "answer": "a"},
		})

		resp, out := doJSONBody(t, env, http.MethodGet, "/api/v1/exam/sessions/"+sessionID+"/result", nil, studentToken)
		require.Equal(t, http.StatusOK, resp.StatusCode, "body=%v", out)
		assert.Equal(t, "locked", out["state"])
		assert.NotEmpty(t, out["result_release_at"])
	})
}

// TestExam_Result_Rank_MultipleSessionsWithTies covers FR-S5-18 end-to-end: rank is 1 +
// count of strictly-higher fully-graded submitted sessions for the same exam, and ties share
// a rank.
func TestExam_Result_Rank_MultipleSessionsWithTies(t *testing.T) {
	env := newTestEnv(t)

	testID := seedTest(t, env, "Ranked", "math", "aljabar", 30)
	q1 := seedQuestionWithPoints(t, env, testID, "mcq", "q1", 1, 1, 0)
	seedMCQOptions(t, env, q1, "a")
	q2 := seedQuestionWithPoints(t, env, testID, "mcq", "q2", 2, 1, 0)
	seedMCQOptions(t, env, q2, "a")

	examID := seedExam(t, env, "Ranked Paket", "score_only", nil)
	linkExamTest(t, env, examID, testID, 0)

	newSession := func(bothCorrect bool) (string, string) {
		studentID := seedUser(t, env, "student", "active", false)
		studentToken := authToken(t, env, studentID, "student")
		regID := seedExamRegistration(t, env, studentID, examID)
		secondAnswer := "b"
		if bothCorrect {
			secondAnswer = "a"
		}
		sessionID := startSubmitSession(t, env, studentToken, regID, []map[string]any{
			{"question_id": q1, "answer": "a"},
			{"question_id": q2, "answer": secondAnswer},
		})
		return sessionID, studentToken
	}

	topSession, topToken := newSession(true)    // score 2
	midSessionB, midTokenB := newSession(false) // score 1
	midSessionC, midTokenC := newSession(false) // score 1, tied with B
	_ = midSessionB
	_ = midSessionC

	resp, out := doJSONBody(t, env, http.MethodGet, "/api/v1/exam/sessions/"+topSession+"/result", nil, topToken)
	require.Equal(t, http.StatusOK, resp.StatusCode, "body=%v", out)
	assert.Equal(t, float64(2), out["score"])
	assert.Equal(t, float64(1), out["rank"], "top scorer has no strictly-higher sessions")

	respB, outB := doJSONBody(t, env, http.MethodGet, "/api/v1/exam/sessions/"+midSessionB+"/result", nil, midTokenB)
	require.Equal(t, http.StatusOK, respB.StatusCode, "body=%v", outB)
	assert.Equal(t, float64(1), outB["score"])
	assert.Equal(t, float64(2), outB["rank"], "one strictly-higher session (top scorer)")

	respC, outC := doJSONBody(t, env, http.MethodGet, "/api/v1/exam/sessions/"+midSessionC+"/result", nil, midTokenC)
	require.Equal(t, http.StatusOK, respC.StatusCode, "body=%v", outC)
	assert.Equal(t, float64(1), outC["score"])
	assert.Equal(t, float64(2), outC["rank"], "tied with session B, shares the same rank")
}
