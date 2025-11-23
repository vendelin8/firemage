package internal

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOnActionChange(t *testing.T) {
	defer setup(t)()

	now := time.Now()
	const p0, p1 = "perm0", "perm1"
	crntUsers = []string{userA.UID, userB.UID, userE.UID}
	localUsers = map[string]*User{userA.UID: userA, userB.UID: userB, userE.UID: userE}
	localUsers[userA.UID].Claims = map[string]any{p0: true}
	localUsers[userB.UID].Claims = map[string]any{p1: true}
	localUsers[userE.UID].Claims = map[string]any{p0: true, p1: true}
	cases := []struct {
		name        string
		key         string
		i           int
		checked     bool
		date        *time.Time
		wantActions map[string]map[string]any
		wantClaims  map[string]map[string]any
	}{
		{
			name:        "check_b",
			key:         p0,
			i:           1,
			checked:     true,
			wantActions: map[string]map[string]any{userB.UID: {p0: true}},
			wantClaims: map[string]map[string]any{
				userA.UID: {p0: true},
				userB.UID: {p0: true, p1: true},
				userE.UID: {p0: true, p1: true},
			},
		},
		{
			name: "set_date_e",
			key:  p1,
			i:    2,
			date: &now,
			wantActions: map[string]map[string]any{
				userB.UID: {p0: true},
				userE.UID: {p1: now},
			},
			wantClaims: map[string]map[string]any{userA.UID: {p0: true},
				userB.UID: {p0: true, p1: true},
				userE.UID: {p0: true, p1: now},
			},
		},
		{
			name:    "uncheck_e",
			key:     p1,
			i:       2,
			checked: false,
			wantActions: map[string]map[string]any{
				userB.UID: {p0: true},
				userE.UID: {p1: false},
			},
			wantClaims: map[string]map[string]any{
				userA.UID: {p0: true},
				userB.UID: {p0: true, p1: true},
				userE.UID: {p0: true, p1: false},
			},
		},
		{
			name:    "uncheck_b",
			key:     p1,
			i:       1,
			checked: false,
			wantActions: map[string]map[string]any{
				userB.UID: {p0: true, p1: false},
				userE.UID: {p1: false},
			},
			wantClaims: map[string]map[string]any{
				userA.UID: {p0: true},
				userB.UID: {p0: true, p1: false},
				userE.UID: {p0: true, p1: false},
			},
		},
		{
			name:    "cancel_b",
			key:     p0,
			i:       1,
			checked: false,
			wantActions: map[string]map[string]any{
				userB.UID: {p1: false},
				userE.UID: {p1: false},
			},
			wantClaims: map[string]map[string]any{
				userA.UID: {p0: true},
				userB.UID: {p1: false},
				userE.UID: {p0: true, p1: false},
			},
		},
		{
			name:        "cancel_e",
			key:         p1,
			i:           2,
			checked:     true,
			wantActions: map[string]map[string]any{userB.UID: {p1: false}},
			wantClaims: map[string]map[string]any{
				userA.UID: {p0: true},
				userB.UID: {p1: false},
				userE.UID: {p0: true},
			},
		},
		{
			name:    "cancel_b2",
			key:     p1,
			i:       1,
			checked: true,
			wantClaims: map[string]map[string]any{
				userA.UID: {p0: true},
				userB.UID: {p1: true},
				userE.UID: {p0: true, p1: true},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			onActionChange(c.i, c.key, c.checked, c.date)
			require.Len(t, actions, len(c.wantActions), "actions")
			for uid, wantChanges := range c.wantActions {
				act := actions[uid]
				require.Len(t, act, len(wantChanges), "actions sub")
				name, email, claims := fixedUserClaims(uid)
				lu := localUsers[uid]
				require.Equal(t, lu.Name, name)
				require.Equal(t, lu.Email, email)
				wantClaims := c.wantClaims[uid]
				require.Equalf(t, wantClaims, claims, "claims doesn't match for user %s", uid)

				for perm, wantVal := range wantChanges {
					require.Equal(t, wantVal, act[perm])
				}
			}
		})
	}
}
