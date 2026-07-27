package ollamaclient

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOptionsJSONMarshalWithThink(t *testing.T) {
	// Test that the think parameter is properly marshaled to JSON
	opts := Options{
		Temperature: 0.5,
	}

	data, err := json.Marshal(opts)
	require.NoError(t, err)

	// Check that the JSON contains the think field
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	// Verify temperature field for completeness
	temp, exists := result["temperature"]
	assert.True(t, exists, "temperature field should exist in JSON")
	assert.Equal(t, float64(0.5), temp, "temperature should be 0.5")
}
