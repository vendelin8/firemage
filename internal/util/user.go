package util

import (
	"maps"
	"sort"

	"github.com/vendelin8/firemage/internal/common"
	"github.com/vendelin8/firemage/internal/global"
	"github.com/vendelin8/firemage/internal/log"
	"go.uber.org/zap"
)

// fixedUserClaims returns user claims with applied actions.
func fixedUserClaims(u *global.User) common.ClaimsMap {
	claims := u.Claims
	if ac, ok := global.Actions[u.UID]; ok {
		log.Lgr.Debug("fixedUserClaims", zap.String("uid", u.UID), zap.Any("actions", ac))
		claims = maps.Clone(u.Claims)
		maps.Copy(claims, ac)
	}
	return claims
}

// FixedUserDetails returns user name, email and claims with applied actions.
func FixedUserDetails(uid string) (string, string, common.ClaimsMap) {
	u := global.LocalUsers[uid]
	return u.Name, u.Email, fixedUserClaims(u)
}

// FixedUserClaims returns user name, email and claims with applied actions.
func FixedUserClaims(uid string) common.ClaimsMap {
	u := global.LocalUsers[uid]
	return fixedUserClaims(u)
}

func SortByNameThenEmail(x []string) {
	sort.Slice(x, func(i, j int) bool {
		ui, uj := global.LocalUsers[x[i]], global.LocalUsers[x[j]]
		uin, ujn := ui.Name, uj.Name
		uiz := len(uin) == 0
		if uiz != (len(ujn) == 0) {
			return !uiz
		}
		if uin < ujn {
			return true
		}
		if uin > ujn {
			return false
		}
		if ui.Email < uj.Email {
			return true
		}
		return false
	})
}
