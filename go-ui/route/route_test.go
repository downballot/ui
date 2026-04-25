package route

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoute(t *testing.T) {
	t.Run("Trivial route", func(t *testing.T) {
		r, err := Parse("/")
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "^/$", r.Regexp())

		t.Run("Does not match empty path", func(t *testing.T) {
			match, variables := r.Match("")
			assert.False(t, match)
			assert.Empty(t, variables)
		})
		t.Run("Does not match relative path", func(t *testing.T) {
			match, variables := r.Match("hello")
			assert.False(t, match)
			assert.Empty(t, variables)
		})
		t.Run("Does match exact path", func(t *testing.T) {
			match, variables := r.Match("/")
			assert.True(t, match)
			assert.Empty(t, variables)
		})
		t.Run("Does not match deeper path", func(t *testing.T) {
			match, variables := r.Match("/hello")
			assert.False(t, match)
			assert.Empty(t, variables)
		})
	})
	t.Run("Exact route", func(t *testing.T) {
		r, err := Parse("/home/user1")
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "^/home/user1$", r.Regexp())

		t.Run("Does not match empty path", func(t *testing.T) {
			match, variables := r.Match("")
			assert.False(t, match)
			assert.Empty(t, variables)
		})
		t.Run("Does not match relative path", func(t *testing.T) {
			match, variables := r.Match("home/user1")
			assert.False(t, match)
			assert.Empty(t, variables)
		})
		t.Run("Does match exact path", func(t *testing.T) {
			match, variables := r.Match("/home/user1")
			assert.True(t, match)
			assert.Empty(t, variables)
		})
		t.Run("Does not match deeper path", func(t *testing.T) {
			match, variables := r.Match("/home/user1/subfolder")
			assert.False(t, match)
			assert.Empty(t, variables)
		})
	})
	t.Run("Single variable", func(t *testing.T) {
		r, err := Parse("/home/:user")
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "^/home/([^/]+)$", r.Regexp())

		t.Run("Does not match empty path", func(t *testing.T) {
			match, variables := r.Match("")
			assert.False(t, match)
			assert.Empty(t, variables)
		})
		t.Run("Does not match relative path", func(t *testing.T) {
			match, variables := r.Match("home/user1")
			assert.False(t, match)
			assert.Empty(t, variables)
		})
		t.Run("Does match exact path", func(t *testing.T) {
			match, variables := r.Match("/home/user1")
			assert.True(t, match)
			assert.Equal(t, map[string]string{"user": "user1"}, variables)
		})
		t.Run("Does not match deeper path", func(t *testing.T) {
			match, variables := r.Match("/home/user1/subfolder")
			assert.False(t, match)
			assert.Empty(t, variables)
		})
	})
	t.Run("Multiple variables", func(t *testing.T) {
		r, err := Parse("/home/:user/:folder")
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "^/home/([^/]+)/([^/]+)$", r.Regexp())

		t.Run("Does not match empty path", func(t *testing.T) {
			match, variables := r.Match("")
			assert.False(t, match)
			assert.Empty(t, variables)
		})
		t.Run("Does not match relative path", func(t *testing.T) {
			match, variables := r.Match("home/user1")
			assert.False(t, match)
			assert.Empty(t, variables)
		})
		t.Run("Does not match prefix path", func(t *testing.T) {
			match, variables := r.Match("/home/user1")
			assert.False(t, match)
			assert.Empty(t, variables)
		})
		t.Run("Does match exact path", func(t *testing.T) {
			match, variables := r.Match("/home/user1/folder1")
			assert.True(t, match)
			assert.Equal(t, map[string]string{"user": "user1", "folder": "folder1"}, variables)
		})
		t.Run("Does not match deeper path", func(t *testing.T) {
			match, variables := r.Match("/home/user1/folder1/subfolder")
			assert.False(t, match)
			assert.Empty(t, variables)
		})
	})
}
