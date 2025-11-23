package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	userA = &User{UID: "a", Name: "Aaa", Email: "a@b.c"}
	userB = &User{UID: "b", Name: "Bbb", Email: "b@b.c"}
	userC = &User{UID: "c", Name: "Aaa", Email: "c@b.c"}
	userE = &User{UID: "e", Name: "", Email: "e@b.c"}
	userF = &User{UID: "f", Name: "", Email: "f@b.c"}
)

func TestSortByNameThenEmail(t *testing.T) {
	cases := []struct {
		name  string
		users []*User
		want  []*User
	}{
		{
			name:  "different_names",
			users: []*User{userA, userB},
		},
		{
			name:  "different_names_wrong_order",
			users: []*User{userB, userA},
			want:  []*User{userA, userB},
		},
		{
			name:  "same_name",
			users: []*User{userA, userC},
		},
		{
			name:  "same_name_wrong_order",
			users: []*User{userC, userA},
			want:  []*User{userA, userC},
		},
		{
			name:  "same_name_empty",
			users: []*User{userA, userC, userE},
		},
		{
			name:  "same_name_empty_wrong_order",
			users: []*User{userE, userC, userA},
			want:  []*User{userA, userC, userE},
		},
		{
			name:  "same_name_empty_wrong_order2",
			users: []*User{userE, userA, userC},
			want:  []*User{userA, userC, userE},
		},
		{
			name:  "all",
			users: []*User{userA, userC, userB, userE, userF},
		},
		{
			name:  "all_wrong_order",
			users: []*User{userF, userC, userE, userA, userB},
			want:  []*User{userA, userC, userB, userE, userF},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			list := make([]string, len(c.users))
			if c.want == nil {
				c.want = make([]*User, len(c.users))
				copy(c.want, c.users)
			}
			for i, u := range c.users {
				list[i] = u.UID
				localUsers[u.UID] = u
			}
			sortByNameThenEmail(list)
			for i, li := range list {
				require.Equal(t, c.want[i].UID, li)
			}
		})
		localUsers = map[string]*User{}
	}
}

func TestDiff(t *testing.T) {
	cases := []struct {
		name      string
		a         map[string]any
		b         map[string]any
		wantPlus  map[string]any
		wantMinus map[string]any
	}{
		{
			name:      "empty",
			wantPlus:  map[string]any{},
			wantMinus: map[string]any{},
		},
		{
			name: "add_to_empty",
			a: map[string]any{
				"a": 0,
				"c": 3,
			},
			b: map[string]any{},
			wantPlus: map[string]any{
				"a": nil,
				"c": nil,
			},
			wantMinus: map[string]any{
				"a": 0,
				"c": 3,
			},
		},
		{
			name: "overwrite_one",
			a: map[string]any{
				"a": 0,
				"c": 3,
			},
			b: map[string]any{
				"a": 0,
				"c": 2,
			},
			wantPlus: map[string]any{
				"c": 2,
			},
			wantMinus: map[string]any{
				"c": 3,
			},
		},
		{
			name: "overwrite_all",
			a: map[string]any{
				"a": 0,
				"c": 3,
			},
			b: map[string]any{
				"a": 1,
				"c": 2,
			},
			wantPlus: map[string]any{
				"a": 1,
				"c": 2,
			},
			wantMinus: map[string]any{
				"a": 0,
				"c": 3,
			},
		},
		{
			name: "all_in_one",
			a: map[string]any{
				"a": 0,
				"c": 3,
			},
			b: map[string]any{
				"a": 2,
				"b": 1,
				"c": 3,
			},
			wantPlus: map[string]any{
				"a": 2,
				"b": 1,
			},
			wantMinus: map[string]any{
				"a": 0,
				"b": nil,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			plus, minus := diffSym(c.a, c.b)
			require.Equal(t, plus, c.wantPlus)
			require.Equal(t, minus, c.wantMinus)

			b := merge(c.a, plus)
			require.Equal(t, c.b, b)

			a := merge(c.b, minus)
			require.Equal(t, c.a, a)
		})
	}
}
