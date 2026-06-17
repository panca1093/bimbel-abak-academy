package integration_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTestEnvBoots(t *testing.T) {
	env := newTestEnv(t)

	resp := env.doJSON(t, http.MethodGet, "/api/v1/health", nil, "")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]string
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Equal(t, "ok", body["status"])
	assert.Equal(t, "ok", body["postgres"])
	assert.Equal(t, "ok", body["redis"])
}
