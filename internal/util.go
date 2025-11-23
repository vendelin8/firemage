package internal

import (
	"maps"
	"sort"

	"firebase.google.com/go/auth"
)

const (
	actList = iota
	actSearch
	actRefresh
	actSave
)

// User is a representation of a user downloaded from firebase auth.
type User struct {
	Name   string
	Email  string
	UID    string
	Claims map[string]any
}

// fltrClaims removes all claims but permissions we're interested in.
func filterClaims(c map[string]any) map[string]any {
	perms := map[string]any{}
	for p, v := range c {
		if _, ok := permsMap[p]; ok {
			perms[p] = v
		}
	}
	return perms
}

// newUserFromAuth converts firebase auth user to User and saves it to local cache if needed.
// Claims are filtered to the permissions we're interested in. Returns the converted User,
// and if it differs from an already downloaded version.
func newUserFromAuth(r *auth.UserRecord, act int) (*User, map[string]any) {
	claims := filterClaims(r.CustomClaims)
	u, ok := localUsers[r.UID]
	if ok {
		u.Name = r.DisplayName

		return u, claims
	}

	u = &User{Name: r.DisplayName, Email: r.Email, Claims: claims, UID: r.UID}
	if act != actRefresh || len(claims) > 0 {
		localUsers[r.UID] = u
	}

	return u, claims
}

func sortByNameThenEmail(x []string) {
	sort.Slice(x, func(i, j int) bool {
		ui, uj := localUsers[x[i]], localUsers[x[j]]
		uiz := len(ui.Name) == 0
		ujz := len(uj.Name) == 0
		if uiz != ujz {
			return !uiz
		}
		if ui.Name < uj.Name {
			return true
		}
		if ui.Name > uj.Name {
			return false
		}
		if ui.Email < uj.Email {
			return true
		}
		return false
	})
}

func mergeIn(a, d map[string]any) {
	for k, v := range d {
		if v == nil {
			delete(a, k)
			continue
		}

		a[k] = v
	}
}

func merge(a, d map[string]any) map[string]any {
	b := maps.Clone(a)
	mergeIn(b, d)
	return b
}

func diffSym(a, b map[string]any) (map[string]any, map[string]any) {
	plus := map[string]any{}
	minus := map[string]any{}

	diffAssym(a, b, plus, minus)
	diffAssym(b, a, minus, plus)

	return plus, minus
}

func diffAssym(a, b, plus, minus map[string]any) {
	for k, bk := range b {
		ak, ok := a[k]
		if ak != bk {
			plus[k] = bk
		}

		if !ok {
			minus[k] = nil
		}
	}
}

func diffAssymHuman(a, b, plus map[string]any) {
	for k, bk := range b {
		if a[k] != bk {
			plus[k] = bk
		}
	}
}

func diffHuman(a, b map[string]any) (map[string]any, map[string]any) {
	plus := map[string]any{}
	minus := map[string]any{}

	diffAssymHuman(a, b, plus)
	diffAssymHuman(b, a, minus)

	return plus, minus
}

func differs(a, b map[string]any) bool {
	if len(a) != len(b) {
		return true
	}

	for k, v := range a {
		if b[k] != v {
			return true
		}
	}

	return false
}
