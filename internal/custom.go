package internal

const (
	// mainTitle will be in the top row of the app.
	mainTitle = "Firemage"
	// minSearchLen defines the minimum length of a text to search for a prefix in name
	// or email address. Less characters will give an error instead of searching.
	minSearchLen = 3
	// the following are permission keys stored in Firebase Auth claims
	admin      = "admin"
	superAdmin = "superAdmin"
	consultant = "consultant"
)

var (
	// permsMap has user table column headers for the permissions defined above.
	permsMap = map[string]string{
		admin:      "Admin",
		superAdmin: "SuperAdmin",
		consultant: "Consultant",
	}
	// allPerms defines the order of table columns of the permissions defined above.
	allPerms = []string{consultant, superAdmin, admin}
	// dateFormat shows how dates should look like.
	dateFormat = "2006-01-02"
)
