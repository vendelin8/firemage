//go:generate go tool go.uber.org/mock/mockgen -package=mock -source=./firebase.go -destination=../mock/mock_firebase.go
package common

import (
	"context"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go/auth"
)

// FbIf is an interface to be able to mock Firebase functionality
type FbIf interface {
	Search(ctx context.Context, key, value string, cb func(uid string) error) error
	StoreAuthClaims(ctx context.Context, uid string, newClaims map[string]any) error
	IterUsers(cb func(*auth.UserRecord) error) error
	GetUsers(ctx context.Context, uids []auth.UserIdentifier) (*auth.GetUsersResult, error)
	GetSpecs(ctx context.Context) (map[string]any, error)
	UpdateSpecs(tr *firestore.Transaction, updates map[string]any) error
	RunTransaction(ctx context.Context, cb func(tr *firestore.Transaction, privileged map[string]any) error) error
	DoList() error
}
