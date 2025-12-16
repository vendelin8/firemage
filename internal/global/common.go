package global

import "github.com/vendelin8/firemage/internal/common"

// User represents a local cached user record
type User struct {
	UID    string
	Email  string
	Name   string
	Claims common.ClaimsMap
}
