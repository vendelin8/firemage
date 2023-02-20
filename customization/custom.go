package main

const (
	// mainTitle will be in the top row of the app.
	mainTitle = "FireMage"
	// minSearchLen defines the minimum length of a text to search for a prefix in name
	// or email address. Less characters will give an error instead of searching.
	minSearchLen = 3
	// the following are permission keys stored in Firebase Auth claims
	kAdmin      = "admin"
	kConsultant = "consultant"
)

var (
	// kPermsMap has user table column headers for the permissions defined above.
	kPermsMap = map[string]string{
		kAdmin:      "Admin",
		kConsultant: "Consultant",
	}
	// kAllPerms defines the order of table columns of the permissions defined above.
	kAllPerms = []string{kConsultant, kAdmin}
)
