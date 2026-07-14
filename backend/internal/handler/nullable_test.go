package handler

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNullable_distinguishes_absent_null_and_value(t *testing.T) {
	type body struct {
		A Nullable[int] `json:"a"`
	}

	// Key absent entirely.
	var absent body
	require.NoError(t, json.Unmarshal([]byte(`{}`), &absent))
	assert.False(t, absent.A.Set, "absent key must leave Set false")

	// Key present, explicit null.
	var explicitNull body
	require.NoError(t, json.Unmarshal([]byte(`{"a": null}`), &explicitNull))
	assert.True(t, explicitNull.A.Set, "explicit null must leave Set true")
	assert.False(t, explicitNull.A.Valid, "explicit null must leave Valid false")

	// Key present, real value.
	var withValue body
	require.NoError(t, json.Unmarshal([]byte(`{"a": 5}`), &withValue))
	assert.True(t, withValue.A.Set)
	assert.True(t, withValue.A.Valid)
	assert.Equal(t, 5, withValue.A.Value)
}
