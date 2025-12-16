package global

import "github.com/vendelin8/firemage/internal/common"

var (
	LocalUsers = map[string]*User{} // downloaded users
	CrntUsers  []string             // currently visible users

	// LocalPrivileged is the downloaded privileged users.
	LocalPrivileged = map[string]struct{}{}

	// actions contains pending permission updates to be saved. key is the uid, value is a map of permission
	// key and a value. True means adding the permission, false means removing it, date means expiry.
	Actions = map[string]common.ClaimsMap{}

	// SavedUsers contains user id lists for all pages.
	SavedUsers = map[string][]string{}
)
