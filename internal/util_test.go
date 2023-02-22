package internal

import (
	"testing"
)

var (
	userA  = &User{UID: "a", Name: "Aaa", Email: "a@b.c"}
	userB  = &User{UID: "b", Name: "Bbb", Email: "b@b.c"}
	userA2 = &User{UID: "c", Name: "Aaa", Email: "c@b.c"}
	userE  = &User{UID: "e", Name: "", Email: "e@b.c"}
	userE2 = &User{UID: "f", Name: "", Email: "f@b.c"}
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
			users: []*User{userA, userA2},
		},
		{
			name:  "same_name_wrong_order",
			users: []*User{userA2, userA},
			want:  []*User{userA, userA2},
		},
		{
			name:  "same_name_empty",
			users: []*User{userA, userA2, userE},
		},
		{
			name:  "same_name_empty_wrong_order",
			users: []*User{userE, userA2, userA},
			want:  []*User{userA, userA2, userE},
		},
		{
			name:  "same_name_empty_wrong_order2",
			users: []*User{userE, userA, userA2},
			want:  []*User{userA, userA2, userE},
		},
		{
			name:  "all",
			users: []*User{userA, userA2, userB, userE, userE2},
		},
		{
			name:  "all_wrong_order",
			users: []*User{userE2, userA2, userE, userA, userB},
			want:  []*User{userA, userA2, userB, userE, userE2},
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
				checkItemAt(t, "uid", c.want[i].UID, li, i)
			}
		})
		localUsers = map[string]*User{}
	}
}
