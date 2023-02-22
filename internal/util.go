package internal

import (
	"sort"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go/auth"
)

var (
	// actions contains permission updates to be saved. key is the uid, value is a map of permission
	// key and a bool value where true means adding the permission, false means removing it.
	actions = map[string]map[string]bool{}
	// savedUsers contains user id lists for all pages.
	savedUsers = map[string][]string{}
)

// User is a representation of a user downloaded from firebase auth.
type User struct {
	Name   string
	Email  string
	UID    string
	Claims map[string]bool
}

// fltrClaims removes all claims but permissions we're interested in.
func filterClaims(c map[string]any) map[string]bool {
	perms := map[string]bool{}
	for p, v := range c {
		if _, ok := permsMap[p]; ok {
			perms[p] = v.(bool)
		}
	}
	return perms
}

// newUserFromAuth converts firebase auth user to User and saves it to local cache if needed.
// Claims are filtered to the permissions we're interested in. Returns the converted User and
// if it's the same as the previous version if any, otherwise true.
func newUserFromAuth(r *auth.UserRecord, saveToCache bool) (*User, bool) {
	claims := filterClaims(r.CustomClaims)
	u, ok := localUsers[r.UID]
	if ok {
		same := compareClaims(u.Claims, claims)
		u.Claims = claims
		u.Name = r.DisplayName
		return u, same
	}
	u = &User{Name: r.DisplayName, Email: r.Email, Claims: claims, UID: r.UID}
	if saveToCache || len(claims) > 0 {
		localUsers[r.UID] = u
	}
	return u, true
}

// compareClaims compares two maps of claims if they have the same key-value pairs.
func compareClaims(a, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

func copyClaims(src map[string]bool) map[string]bool {
	dst := map[string]bool{}
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// saveListBody is the middle of the firestore transaction for listing, saving and refreshing.
// The input is a map of uid -> email users that come from the cached list of privileged users.
// It looks for new privileged users in case of refresh only by iterating through all auth users.
// On the other side, it checks all users from the cache list if they are still privileged for
// all three cases. In case of refresh and save it checks if the permissions are still the same.
// Not for listing, because it runs only at init when we download the list of permissions, so
// there's nothing to compare them with. In case of save we're merging actions into current
// permissions, updating Firebase auth claims and the cached firestore list of privileged users.
//
//nolint:all
func saveListBody(act int, res map[string]any) []firestore.Update {
	crntMap := map[string]struct{}{}
	var changed, newUsers, empty []string
	var claimCb func(r *auth.UserRecord)
	updates := make([]firestore.Update, 0, len(actions))
	manualChanges := false
	var uids map[string]struct{}
	if act != actRefresh {
		uids = map[string]struct{}{}
		for uid, email := range res {
			if _, ok := localUsers[uid]; !ok {
				localUsers[uid] = &User{Email: email.(string), UID: uid}
			}
			uids[uid] = es
		}
		for uid := range actions {
			uids[uid] = es
		}
	}
	if fe.currentPage() == lst {
		savedUsers[lst] = crntUsers
	}
	for _, uid := range savedUsers[lst] { // already downloaded users
		crntMap[uid] = es
		if act == actRefresh {
			continue
		}
		if _, ok := res[uid]; !ok {
			uids[uid] = es
		}
	}
	// this is called in all cases to check consistency
	preCb := func(r *auth.UserRecord) {
		uid := r.UID
		u, same := newUserFromAuth(r, act != actRefresh) // don't want to save all users locally
		_, inRes := res[uid]
		if len(u.Claims) == 0 {
			if !inRes {
				return
			}
			if _, ok := actions[uid]; !ok && act != actRefresh {
				empty = append(empty, u.Email)
			}
			manualChanges = true
			updates = append(updates, firestore.Update{Path: uid, Value: firestore.Delete})
			delete(res, uid)
			return
		}
		if !inRes {
			updates = append(updates, firestore.Update{Path: uid, Value: u.Email})
			res[uid] = es // no need to add email here
			manualChanges = true
		}
		if act == actList {
			return
		}
		if _, inCrnt := crntMap[uid]; !inCrnt {
			newUsers = append(newUsers, u.Email)
		}

		if !same {
			changed = append(changed, u.Email)
		}
	}
	if act == actSave {
		claimCb = func(r *auth.UserRecord) {
			preCb(r)
			uid := r.UID
			u := localUsers[uid]
			a, ok := actions[uid]
			if !ok {
				return
			}
			newClaims := copyClaims(u.Claims)
			for p, ch := range a {
				if ch { // add permission
					if r.CustomClaims == nil {
						r.CustomClaims = map[string]any{}
					}
					r.CustomClaims[p] = true
					newClaims[p] = true
				} else {
					delete(r.CustomClaims, p)
					delete(newClaims, p)
				}
			}
			if !writeErrorIf(fb.setClaims(uid, r.CustomClaims)) {
				return
			}
			u.Claims = newClaims
			delete(actions, uid)
			_, inRes := res[uid]
			if inRes == (len(u.Claims) > 0) {
				return
			}
			if inRes {
				delete(res, uid)
				updates = append(updates, firestore.Update{Path: uid, Value: firestore.Delete})
				return
			}
			updates = append(updates, firestore.Update{Path: uid, Value: r.Email})
			res[uid] = es
		}
	} else {
		claimCb = preCb
	}
	if act == actRefresh {
		fb.iterUsers(claimCb)
	} else {
		uidList := make([]auth.UserIdentifier, 0, len(uids))
		for uid := range uids {
			uidList = append(uidList, auth.UIDIdentifier{UID: uid})
		}
		if !fb.downloadClaims(uidList, claimCb) && manualChanges {
			manualChanges = false // already written
		}
	}
	savedUsers[lst] = savedUsers[lst][:0]
	for uid := range res {
		savedUsers[lst] = append(savedUsers[lst], uid)
	}
	sortByNameThenEmail(savedUsers[lst])
	if fe.currentPage() == lst {
		crntUsers = savedUsers[lst]
	}
	writeErrorList(errChanged, changed)
	writeErrorList(errEmpty, empty)
	writeErrorList(errNew, newUsers)
	if manualChanges {
		writeErrorStr(errManual)
		if act != actRefresh {
			writeErrorStr(warnMayRefresh)
		}
	}
	return updates
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
