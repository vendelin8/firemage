package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vendelin8/firemage/internal/global"
)

func TestSortByNameThenEmail(t *testing.T) {
	t.Run("sorts users by name then email", func(t *testing.T) {
		// Setup
		global.LocalUsers = map[string]*global.User{
			"uid1": {UID: "uid1", Name: "Alice", Email: "alice@example.com"},
			"uid2": {UID: "uid2", Name: "Bob", Email: "bob@example.com"},
			"uid3": {UID: "uid3", Name: "Alice", Email: "aliceb@example.com"},
		}
		uids := []string{"uid2", "uid3", "uid1"}

		SortByNameThenEmail(uids)

		assert.Equal(t, []string{"uid1", "uid3", "uid2"}, uids)
	})

	t.Run("sorts users with same name by email", func(t *testing.T) {
		global.LocalUsers = map[string]*global.User{
			"uid1": {UID: "uid1", Name: "Alice", Email: "alice.b@example.com"},
			"uid2": {UID: "uid2", Name: "Alice", Email: "alice.a@example.com"},
			"uid3": {UID: "uid3", Name: "Alice", Email: "alice.c@example.com"},
		}
		uids := []string{"uid1", "uid3", "uid2"}

		SortByNameThenEmail(uids)

		assert.Equal(t, []string{"uid2", "uid1", "uid3"}, uids)
	})

	t.Run("users with names come before users without names", func(t *testing.T) {
		global.LocalUsers = map[string]*global.User{
			"uid1": {UID: "uid1", Name: "", Email: "z@example.com"},
			"uid2": {UID: "uid2", Name: "Bob", Email: "b@example.com"},
			"uid3": {UID: "uid3", Name: "", Email: "a@example.com"},
		}
		uids := []string{"uid3", "uid2", "uid1"}

		SortByNameThenEmail(uids)

		// Bob should come first (has name), then the two without names
		assert.Equal(t, []string{"uid2", "uid3", "uid1"}, uids)
	})

	t.Run("sorts multiple users without names by email", func(t *testing.T) {
		global.LocalUsers = map[string]*global.User{
			"uid1": {UID: "uid1", Name: "", Email: "z@example.com"},
			"uid2": {UID: "uid2", Name: "", Email: "a@example.com"},
			"uid3": {UID: "uid3", Name: "", Email: "m@example.com"},
		}
		uids := []string{"uid1", "uid3", "uid2"}

		SortByNameThenEmail(uids)

		assert.Equal(t, []string{"uid2", "uid3", "uid1"}, uids)
	})

	t.Run("sorts mixed users correctly", func(t *testing.T) {
		global.LocalUsers = map[string]*global.User{
			"uid1": {UID: "uid1", Name: "Charlie", Email: "charlie@example.com"},
			"uid2": {UID: "uid2", Name: "", Email: "z@example.com"},
			"uid3": {UID: "uid3", Name: "Alice", Email: "alice@example.com"},
			"uid4": {UID: "uid4", Name: "", Email: "a@example.com"},
			"uid5": {UID: "uid5", Name: "Bob", Email: "bob@example.com"},
		}
		uids := []string{"uid2", "uid5", "uid4", "uid1", "uid3"}

		SortByNameThenEmail(uids)

		// Expected order: Alice, Bob, Charlie (with names), then a@, z@ (without names)
		assert.Equal(t, []string{"uid3", "uid5", "uid1", "uid4", "uid2"}, uids)
	})

	t.Run("handles empty slice", func(t *testing.T) {
		global.LocalUsers = map[string]*global.User{}
		uids := []string{}

		// Should not panic
		SortByNameThenEmail(uids)

		assert.Equal(t, []string{}, uids)
	})

	t.Run("handles single user", func(t *testing.T) {
		global.LocalUsers = map[string]*global.User{
			"uid1": {UID: "uid1", Name: "Alice", Email: "alice@example.com"},
		}
		uids := []string{"uid1"}

		SortByNameThenEmail(uids)

		assert.Equal(t, []string{"uid1"}, uids)
	})

	t.Run("handles users with empty names correctly", func(t *testing.T) {
		global.LocalUsers = map[string]*global.User{
			"uid1": {UID: "uid1", Name: "Alice", Email: "alice@example.com"},
			"uid2": {UID: "uid2", Name: "", Email: "bob@example.com"},
		}
		uids := []string{"uid2", "uid1"}

		SortByNameThenEmail(uids)

		// Alice (with name) should come before Bob (without name)
		assert.Equal(t, []string{"uid1", "uid2"}, uids)
	})

	t.Run("is stable sort for users with identical names and emails", func(t *testing.T) {
		global.LocalUsers = map[string]*global.User{
			"uid1": {UID: "uid1", Name: "Alice", Email: "alice@example.com"},
			"uid2": {UID: "uid2", Name: "Alice", Email: "alice@example.com"},
		}
		uids := []string{"uid1", "uid2"}

		SortByNameThenEmail(uids)

		// Both have same name and email, so order should be stable (uid1, uid2)
		assert.Equal(t, []string{"uid1", "uid2"}, uids)
	})
}
