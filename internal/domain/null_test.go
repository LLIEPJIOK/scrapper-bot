package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNull_MarshalJSON(t *testing.T) {
	t.Parallel()

	t.Run("valid bool", func(t *testing.T) {
		t.Parallel()

		nullBool := domain.NewNull(true)
		data, err := json.Marshal(nullBool)
		require.NoError(t, err, "should marshal valid null")
		assert.Equal(t, "true", string(data), "should marshal to true")
	})

	t.Run("valid string", func(t *testing.T) {
		t.Parallel()

		nullString := domain.NewNull("test")
		data, err := json.Marshal(nullString)
		require.NoError(t, err, "should marshal valid null")
		assert.Equal(t, `"test"`, string(data), "should marshal to test")
	})

	t.Run("invalid bool", func(t *testing.T) {
		t.Parallel()

		nullBool := domain.Null[bool]{}
		data, err := json.Marshal(nullBool)
		require.NoError(t, err, "should marshal invalid null")
		assert.Equal(t, "null", string(data), "should marshal to null")
	})

	t.Run("invalid string", func(t *testing.T) {
		t.Parallel()

		nullString := domain.Null[string]{}
		data, err := json.Marshal(nullString)
		require.NoError(t, err, "should marshal invalid null")
		assert.Equal(t, "null", string(data), "should marshal to null")
	})
}

func TestNull_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	t.Run("valid bool", func(t *testing.T) {
		t.Parallel()

		var nullBool domain.Null[bool]
		err := json.Unmarshal([]byte("true"), &nullBool)
		require.NoError(t, err, "should unmarshal valid bool")
		assert.True(t, nullBool.Valid, "should be valid")
		assert.True(t, nullBool.Value, "should be true")
	})

	t.Run("valid string", func(t *testing.T) {
		t.Parallel()

		var nullString domain.Null[string]
		err := json.Unmarshal([]byte(`"test"`), &nullString)
		require.NoError(t, err, "should unmarshal valid string")
		assert.True(t, nullString.Valid, "should be valid")
		assert.Equal(t, "test", nullString.Value, "should be test")
	})

	t.Run("invalid bool", func(t *testing.T) {
		t.Parallel()

		var nullBool domain.Null[bool]
		err := json.Unmarshal([]byte("null"), &nullBool)
		require.NoError(t, err, "should unmarshal null")
		assert.False(t, nullBool.Valid, "should be invalid")
		assert.False(t, nullBool.Value, "should be false")
	})

	t.Run("invalid string", func(t *testing.T) {
		t.Parallel()

		var nullString domain.Null[string]
		err := json.Unmarshal([]byte("null"), &nullString)
		require.NoError(t, err, "should unmarshal null")
		assert.False(t, nullString.Valid, "should be invalid")
		assert.Equal(t, "", nullString.Value, "should be empty string")
	})

	t.Run("error bool", func(t *testing.T) {
		t.Parallel()

		var nullBool domain.Null[bool]
		err := json.Unmarshal([]byte("invalid"), &nullBool)
		require.Error(t, err, "should return error for invalid JSON")
		assert.Error(t, err)
	})

	t.Run("error string", func(t *testing.T) {
		t.Parallel()

		var nullString domain.Null[string]
		err := json.Unmarshal([]byte("invalid"), &nullString)
		require.Error(t, err, "should return error for invalid JSON")
		assert.Error(t, err)
	})
}

func TestNewNull(t *testing.T) {
	t.Parallel()

	t.Run("bool", func(t *testing.T) {
		t.Parallel()

		nullBool := domain.NewNull(true)
		assert.True(t, nullBool.Valid, "should be valid")
		assert.True(t, nullBool.Value, "should be true")
	})

	t.Run("string", func(t *testing.T) {
		t.Parallel()

		nullString := domain.NewNull("test")
		assert.True(t, nullString.Valid, "should be valid")
		assert.Equal(t, "test", nullString.Value, "should be test")
	})
}
