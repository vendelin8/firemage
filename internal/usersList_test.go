package main

import (
	"testing"
)

func TestOnActionChange(t *testing.T) {
	defer setup()()

	const p0, p1 = "perm0", "perm1"
	crntUsers = []string{userA.UID, userB.UID, userE.UID}
	localUsers = map[string]*User{userA.UID: userA, userB.UID: userB, userE.UID: userE}
	localUsers[userA.UID].Claims = map[string]bool{p0: true}
	localUsers[userB.UID].Claims = map[string]bool{p1: true}
	localUsers[userE.UID].Claims = map[string]bool{p0: true, p1: true}
	cases := []struct {
		name        string
		key         string
		i           int
		checked     bool
		wantActions map[string]map[string]bool
		wantClaims  map[string]map[string]bool
	}{
		{
			name:        "check_b",
			key:         p0,
			i:           1,
			checked:     true,
			wantActions: map[string]map[string]bool{userB.UID: map[string]bool{p0: true}},
			wantClaims: map[string]map[string]bool{userA.UID: map[string]bool{p0: true},
				userB.UID: map[string]bool{p0: true, p1: true},
				userE.UID: map[string]bool{p0: true, p1: true}},
		},
		{
			name:    "uncheck_e",
			key:     p1,
			i:       2,
			checked: false,
			wantActions: map[string]map[string]bool{userB.UID: map[string]bool{p0: true},
				userE.UID: map[string]bool{p1: false}},
			wantClaims: map[string]map[string]bool{userA.UID: map[string]bool{p0: true},
				userB.UID: map[string]bool{p0: true, p1: true},
				userE.UID: map[string]bool{p0: true, p1: false}},
		},
		{
			name:    "uncheck_b",
			key:     p1,
			i:       1,
			checked: false,
			wantActions: map[string]map[string]bool{userB.UID: map[string]bool{p0: true, p1: false},
				userE.UID: map[string]bool{p1: false}},
			wantClaims: map[string]map[string]bool{userA.UID: map[string]bool{p0: true},
				userB.UID: map[string]bool{p0: true, p1: false},
				userE.UID: map[string]bool{p0: true, p1: false}},
		},
		{
			name:    "cancel_b",
			key:     p0,
			i:       1,
			checked: false,
			wantActions: map[string]map[string]bool{userE.UID: map[string]bool{p1: false},
				userB.UID: map[string]bool{p1: false}},
			wantClaims: map[string]map[string]bool{userA.UID: map[string]bool{p0: true},
				userB.UID: map[string]bool{p1: false},
				userE.UID: map[string]bool{p0: true, p1: false}},
		},
		{
			name:        "cancel_e",
			key:         p1,
			i:           2,
			checked:     true,
			wantActions: map[string]map[string]bool{userB.UID: map[string]bool{p1: false}},
			wantClaims: map[string]map[string]bool{userA.UID: map[string]bool{p0: true},
				userB.UID: map[string]bool{p1: false}, userE.UID: map[string]bool{p0: true}},
		},
		{
			name:    "cancel_b2",
			key:     p1,
			i:       1,
			checked: true,
			wantClaims: map[string]map[string]bool{userA.UID: map[string]bool{p0: true},
				userB.UID: map[string]bool{p1: true}, userE.UID: map[string]bool{p0: true, p1: true}},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			onActionChange(c.checked, c.i, c.key)
			checkLen(t, "actions", len(c.wantActions), len(actions), c.wantActions, actions)
			for uid, wantChanges := range c.wantActions {
				act := actions[uid]
				checkLen(t, "actions sub", len(wantChanges), len(act), wantChanges, act)
				name, email, claims := fixedUserClaims(uid)
				lu := localUsers[uid]
				if name != lu.Name {
					t.Errorf("name doesn't match expected %s != %s result", lu.Name, name)
				}
				if email != lu.Email {
					t.Errorf("email doesn't match expected %s != %s result", lu.Email, email)
				}
				wantClaims := c.wantClaims[uid]
				if !compareClaims(claims, wantClaims) {
					t.Errorf("claims doesn't match for user %s expected %#v != %#v result",
						uid, wantClaims, claims)
				}
				for perm, wantVal := range wantChanges {
					actVal := act[perm]
					if actVal != wantVal {
						t.Errorf("value doesn't match expected %t != %t result", wantVal, actVal)
					}
				}
			}
		})
	}
}
