//go:generate mockgen -package=mock -source=./frontend.go -destination=../mock/mock_frontend.go
package common

import (
	"context"
	"time"

	"github.com/vendelin8/tview"
)

// FeIf is an interface to be able to mock tview GUI functionality.
type FeIf interface {
	Run()
	CurrentPage() string
	SetPage(string)
	SetOnShow(string, func())
	ShowMsg(ms ...string)
	ShowConfirm(onYes, onNo func(), ms ...string)
	ShowProgress(ctx context.Context, cancelFunc context.CancelFunc, ms ...string)
	ClaimButtonSetDisabled(index int, isDisabled bool)
	HidePopup(popup string)
	LayoutUsers()
	Quit()

	ShowClaimChoser(i int, key string, c Claim)
	CreateClaimChoser()
	ClaimsDateSetDisabled(bool)
	ClaimsBtns(bool)
	ClaimsSetDate(time.Time)
	ClaimsDate() *time.Time
	ReplaceTableItem(i int, key string, p tview.Primitive)
}
