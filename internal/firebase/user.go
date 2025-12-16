package firebase

import (
	"fmt"
	"maps"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go/auth"
	"github.com/vendelin8/firemage/internal/common"
	"github.com/vendelin8/firemage/internal/frontend/window"
	"github.com/vendelin8/firemage/internal/global"
	"github.com/vendelin8/firemage/internal/lang"
	"github.com/vendelin8/firemage/internal/util"
)

// newUserFromAuth converts firebase auth user to User and saves it to local cache if needed.
// Claims are filtered to the permissions we're interested in. Returns the converted User,
// and if it differs from an already downloaded version.
func newUserFromAuth(r *auth.UserRecord, act int, privileged map[string]any, updates map[string]any) (*global.User, error) {
	filtered := filterClaims(r.CustomClaims)
	uid := r.UID
	u, ok := global.LocalUsers[uid]
	if !ok {
		u = &global.User{Name: r.DisplayName, Email: r.Email, Claims: filtered, UID: uid}
		if act != actRefresh || len(filtered) > 0 {
			global.LocalUsers[uid] = u
		}

		// user just saved, can't be
		return u, nil
	}

	if !differs(u.Claims, filtered) {
		return u, nil
	}

	toCompare := u.Claims
	if act == actSave {
		toCompare = util.FixedUserClaims(uid)
		if !differs(toCompare, filtered) {
			return u, nil
		}
	}

	plus, minus := diffHuman(u.Claims, filtered)

	window.UseConfirm()
	window.WriteErrorStr(fmt.Sprintf(lang.ErrPermsChangedS, u.Email))
	if len(plus) > 0 {
		window.WriteErrorStr(fmt.Sprintf(lang.WarnAddedPemsS, plus))
	}
	if len(minus) > 0 {
		window.WriteErrorStr(fmt.Sprintf(lang.WarnRemovedPemsS, minus))
	}

	hasClaims := len(u.Claims) > 0
	_, isPrivileged := privileged[uid]

	if hasClaims != isPrivileged {
		window.WriteErrorStr(lang.ErrManualS)
	}

	window.WriteErrorStr(lang.ConfirmSaveS)

	var err error

	window.ShowConfirm(func() {
		if hasClaims {
			privileged[uid] = struct{}{}

			if !isPrivileged {
				updates[uid] = u.Email
			}
		} else {
			if isPrivileged {
				updates[uid] = firestore.Delete
			}
			delete(privileged, uid)
		}

		if err = setPermissions(r, toCompare); err != nil {
			err = fmt.Errorf(lang.ErrSetPerms, err)
			return
		}

		u.Claims = toCompare
	}, func() {
		global.LocalUsers[uid] = u
	})

	return u, err
}

// fltrClaims removes all claims but permissions we're interested in.
func filterClaims(c map[string]any) common.ClaimsMap {
	var err error
	perms := *common.NewClaimsMap()
	for p, v := range c {
		if _, ok := common.PermsMap[p]; ok {
			if perms[p], err = common.NewClaimFrom(v); err != nil {
				common.Fe.ShowMsg(fmt.Sprintf("%s: filterClaims: %#v", err, c))
			}
		}
	}

	return perms
}

func merge(a map[string]any, d common.ClaimsMap) map[string]any {
	b := maps.Clone(a)
	mergeIn(b, d)
	return b
}

func mergeIn(a map[string]any, d common.ClaimsMap) {
	for k, v := range d {
		a[k] = v.ToAny()
	}
}

/*func diffSym(a, b map[string]any) (map[string]any, map[string]any) {
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
}*/

func diffAssymHuman(a, b, plus common.ClaimsMap) {
	for k, bk := range b {
		if a[k] != bk {
			plus[k] = bk
		}
	}
}

func diffHuman(a, b common.ClaimsMap) (common.ClaimsMap, common.ClaimsMap) {
	plus := common.ClaimsMap{}
	minus := common.ClaimsMap{}

	diffAssymHuman(a, b, plus)
	diffAssymHuman(b, a, minus)

	return plus, minus
}

func differs(a, b common.ClaimsMap) bool {
	if len(a) != len(b) {
		return true
	}

	for k, v := range a {
		if b[k].Differs(v) {
			return true
		}
	}

	return false
}
