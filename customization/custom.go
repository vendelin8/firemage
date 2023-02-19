package main

const (
	mainTitle    = "NapOwner"
	kAdmin       = "admin"
	kConsultant  = "consultant"
	minSearchLen = 3
)

var (
	kPermsMap = map[string]string{
		kAdmin:      "Admin",
		kConsultant: "Konzulens",
	}
	kAllPerms = []string{kConsultant, kAdmin}
)
