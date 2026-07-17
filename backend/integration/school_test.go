package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"akademi-bimbel/internal/service"

	"github.com/stretchr/testify/require"
)

// authTokenWithSchool returns a JWT with a schoolID claim and writes the
// session key to Redis, mirroring authToken's contract.
func authTokenWithSchool(t *testing.T, env *testEnv, userID, role, schoolID string) string {
	t.Helper()
	ctx := context.Background()
	caps := service.Capabilities(role)
	tokenStr, jti, err := env.signer.SignAccess(userID, role, &schoolID, caps)
	require.NoError(t, err)
	err = env.rdb.Set(ctx, "session:access:"+jti, userID, 15*time.Minute).Err()
	require.NoError(t, err)
	return tokenStr
}

func TestSchoolCRUD_Integration(t *testing.T) {
	env := newTestEnv(t)

	// 1. Seed a school
	var schoolID string
	err := env.pool.QueryRow(t.Context(),
		`INSERT INTO school (name, code, npsn, school_types, alamat, status)
		 VALUES ($1, $2, $3, $4, $5, 'active') RETURNING id`,
		"SMAN Test", "smantest", "20000000", []string{"SMA"}, "Jl. Test No.1",
	).Scan(&schoolID)
	require.NoError(t, err)
	require.NotEmpty(t, schoolID)

	// 2. Create an admin_school account bound to the school
	superUserID := seedUser(t, env, "super_admin", "active", false)
	superToken := authToken(t, env, superUserID, "super_admin")
	createBody := map[string]interface{}{
		"email":     "schooladmin@test.com",
		"name":      "School Admin",
		"role":      "admin_school",
		"password":  "password123",
		"school_id": schoolID,
	}
	b, _ := json.Marshal(createBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/system/accounts", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+superToken)
	rec := httptest.NewRecorder()
	env.server.Config.Handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)

	var createResp map[string]interface{}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&createResp))
	createdAdminID := createResp["id"].(string)
	require.NotEmpty(t, createdAdminID)

	// 3. Register a student as the school admin
	adminToken := authTokenWithSchool(t, env, createdAdminID, "admin_school", schoolID)
	studentBody := map[string]interface{}{
		"name": "Test Student",
		"nis":  "12345",
	}
	b, _ = json.Marshal(studentBody)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/admin/students", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	rec = httptest.NewRecorder()
	env.server.Config.Handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)

	var regResp map[string]interface{}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&regResp))
	require.NotEmpty(t, regResp["temp_password"])
	require.Equal(t, "smantest_12345", regResp["username"])
	studentID := regResp["id"].(string)

	// 4. Student can log in with username + temp_password (FR-STU-10)
	loginBody := map[string]string{
		"identifier": "smantest_12345",
		"password":   regResp["temp_password"].(string),
	}
	b, _ = json.Marshal(loginBody)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	env.server.Config.Handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	// 5. Credential reissue returns a new password
	req = httptest.NewRequest(http.MethodGet,
		"/api/v1/admin/students/"+studentID+"/credentials", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	rec = httptest.NewRecorder()
	env.server.Config.Handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var credResp map[string]interface{}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&credResp))
	require.NotEqual(t, regResp["temp_password"], credResp["temp_password"],
		"reissue should return a different password")
}

// TestAdminCreateSchool_OmittedSchoolTypes_Integration reproduces the
// FR-SCH-02 blocker: omitting school_types (a spec-optional field) must not
// 500 — the NOT NULL column has no default applied when an explicit NULL is
// inserted, so nil []string must be coerced to []string{} before the INSERT.
func TestAdminCreateSchool_OmittedSchoolTypes_Integration(t *testing.T) {
	env := newTestEnv(t)

	superUserID := seedUser(t, env, "super_admin", "active", false)
	superToken := authToken(t, env, superUserID, "super_admin")

	body := map[string]interface{}{
		"name": "SMAN Omitted Types",
		"code": "smanomitted",
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/schools", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+superToken)
	rec := httptest.NewRecorder()
	env.server.Config.Handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())

	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	types, _ := resp["school_types"].([]interface{})
	require.Empty(t, types, "school_types should default to an empty array, not null")
}

// TestAdminCreateSchool_ResponseStatusActive_Integration reproduces the
// FR-SCH-02 blocker where a successful create response reported status:""
// instead of "active", even though the DB row persisted correctly.
func TestAdminCreateSchool_ResponseStatusActive_Integration(t *testing.T) {
	env := newTestEnv(t)

	superUserID := seedUser(t, env, "super_admin", "active", false)
	superToken := authToken(t, env, superUserID, "super_admin")

	body := map[string]interface{}{
		"name":         "SMAN Status Check",
		"code":         "smanstatuschk",
		"school_types": []string{"SMA"},
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/schools", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+superToken)
	rec := httptest.NewRecorder()
	env.server.Config.Handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code, "body: %s", rec.Body.String())

	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	require.Equal(t, "active", resp["status"])
}

func TestSchoolCodeChange_Integration(t *testing.T) {
	env := newTestEnv(t)

	// Seed a school
	var schoolID string
	err := env.pool.QueryRow(t.Context(),
		`INSERT INTO school (name, code, npsn, school_types, alamat, status)
		 VALUES ($1, $2, $3, $4, $5, 'active') RETURNING id`,
		"Code Change School", "codechg", "20000001", []string{"SMA"}, "Jl. Test",
	).Scan(&schoolID)
	require.NoError(t, err)

	// Register a student
	alsUserID := seedUser(t, env, "admin_school", "active", false)
	adminToken := authTokenWithSchool(t, env, alsUserID, "admin_school", schoolID)
	studentBody := map[string]string{"name": "Stu", "nis": "chgtest"}
	b, _ := json.Marshal(studentBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/students", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	rec := httptest.NewRecorder()
	env.server.Config.Handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	// Change code with students — should succeed (lock removed)
	suUserID := seedUser(t, env, "super_admin", "active", false)
	superToken := authToken(t, env, suUserID, "super_admin")
	updateBody := map[string]string{"code": "newcodechg"}
	b, _ = json.Marshal(updateBody)
	req = httptest.NewRequest(http.MethodPut, "/api/v1/admin/schools/"+schoolID, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+superToken)
	rec = httptest.NewRecorder()
	env.server.Config.Handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestRowScoping_Integration(t *testing.T) {
	env := newTestEnv(t)

	// Two schools
	var schoolA, schoolB string
	env.pool.QueryRow(t.Context(),
		`INSERT INTO school (name, code, npsn, school_types, alamat, status)
		 VALUES ($1, $2, $3, $4, $5, 'active') RETURNING id`,
		"School A", "schoola", "20000002", []string{"SMA"}, "Jl. A",
	).Scan(&schoolA)
	env.pool.QueryRow(t.Context(),
		`INSERT INTO school (name, code, npsn, school_types, alamat, status)
		 VALUES ($1, $2, $3, $4, $5, 'active') RETURNING id`,
		"School B", "schoolb", "20000003", []string{"SMA"}, "Jl. B",
	).Scan(&schoolB)

	// Admin A registers a student
	adminAUserID := seedUser(t, env, "admin_school", "active", false)
	tokenA := authTokenWithSchool(t, env, adminAUserID, "admin_school", schoolA)
	studentBody := map[string]string{"name": "Student A", "nis": "a001"}
	b, _ := json.Marshal(studentBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/students", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenA)
	rec := httptest.NewRecorder()
	env.server.Config.Handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)
	var regResp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&regResp)
	studentID := regResp["id"].(string)

	// Admin B tries to access Admin A's student → 404
	adminBUserID := seedUser(t, env, "admin_school", "active", false)
	tokenB := authTokenWithSchool(t, env, adminBUserID, "admin_school", schoolB)
	req = httptest.NewRequest(http.MethodPatch, "/api/v1/admin/students/"+studentID,
		bytes.NewReader([]byte(`{"status":"deactivated"}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenB)
	rec = httptest.NewRecorder()
	env.server.Config.Handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNotFound, rec.Code)
}
