package frontend

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/vendelin8/firemage/internal/common"
	"github.com/vendelin8/firemage/internal/frontend/window"
	"github.com/vendelin8/firemage/internal/global"
	"github.com/vendelin8/firemage/internal/lang"
	"github.com/vendelin8/firemage/internal/log"
	"github.com/vendelin8/firemage/internal/util"
	"github.com/vendelin8/tview"
	"go.uber.org/zap"
)

const namedCols = 2 // name and email column

func (f *Frontend) initUsersList() {
	colNum := len(common.AllPerms) + namedCols
	f.userHdrs = make([]string, colNum)
	f.userHdrs[0] = lang.SName
	f.userHdrs[1] = lang.SEmail
	colSizes := make([]int, colNum)
	colSizes[0] = 25
	colSizes[1] = 0 // fill
	for i, perm := range common.AllPerms {
		j := i + namedCols
		f.userHdrs[j] = common.PermsMap[perm]
		colSizes[j] = max(len(perm), len(common.DateFormat)) + 1 // checkbox padding
	}
	for col, text := range f.userHdrs {
		f.userTbl.AddItem(newText(text), 0, col, 1, 1, 0, 0, false)
	}
	f.userTbl.SetColumns(colSizes...)
}

// activatePopup is called when an empty checkbox is checked, or a date is clicked. It pops up
// a dialog to change value to true or a given date.
func activatePopup(i int, key string, c common.Claim) {
	window.PushPopup(lang.PopupClaim)
	common.Fe.ShowClaimChoser(c, func(c common.Claim) {
		log.Lgr.Debug("activatePopup done", zap.Int("i", i), zap.String("key", key), zap.Stringer("claim", &c))
		onActionChange(i, key, c)
	})
}

// tableCB returns a checkbox or date text to the claim table filled with the saved value.
func tableCB(i int, key string, c common.Claim) tview.Primitive {
	var (
		box tview.FormItem
		bgc = tview.Styles.PrimitiveBackgroundColor
		ftc = tview.Styles.PrimaryTextColor
	)

	if i%2 == 1 {
		bgc = tview.Styles.PrimaryTextColor
		ftc = tview.Styles.PrimitiveBackgroundColor
	}

	var tv *tview.InputField

	if c.Date == nil {
		box = createTableCheckbox(i, key, c, ftc, bgc)
	} else {
		tv = createTableDateField(i, key, c, ftc, bgc)
		box = tv
	}

	center := tview.NewCenter(box, box.GetFieldWidth(), box.GetFieldHeight())
	if i%2 == 1 {
		center.SetBackgroundColor(tview.Styles.PrimaryTextColor)
		// if tv != nil {
		// 	tv.SetFieldTextColor(tview.Styles.ContrastBackgroundColor)
		// }
	}
	return center
}

func createTableCheckbox(i int, key string, c common.Claim, ftc, bgc tcell.Color) *tview.Checkbox {
	manualChange := false
	cb := tview.NewCheckbox()
	cb.SetChecked(c.Checked).SetChangedFunc(func(checked bool) {
		onActionChange(i, key, common.Claim{Checked: checked})
		if manualChange {
			return
		}

		if !checked {
			return // unchecked, no popup
		}

		manualChange = true // to avoid re-entrance
		cb.SetChecked(!checked)
		manualChange = false
		activatePopup(i, key, common.Claim{Checked: checked})
	}).SetFieldTextColor(ftc)
	cb.SetBackgroundColor(bgc)

	return cb
}

func createTableDateField(i int, key string, c common.Claim, ftc, bgc tcell.Color) *tview.InputField {
	tv := tview.NewInputField().
		SetText(c.FormatDate()).SetFieldTextColor(ftc).SetFieldWidth(len(common.DateFormat))
	tv.SetBackgroundColor(bgc).
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			activatePopup(i, key, c)
			return nil
		}).
		SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
			if action != tview.MouseLeftClick {
				return action, event
			}

			activatePopup(i, key, c)
			return tview.MouseConsumed, nil
		})

	tv.SetDisabled(true)

	return tv
}

// onActionChange in called when a user claim is changed by checkbox or date field.
func onActionChange(i int, key string, c common.Claim) {
	uid := global.CrntUsers[i]
	current := global.LocalUsers[uid].Claims[key]
	currentVisual := util.FixedUserClaims(uid)[key]
	acts := global.Actions[uid]
	log.Lgr.Debug("onActionChange", zap.Int("i", i), zap.String("key", key), zap.Any("claim", c), zap.Any("current", current), zap.Any("currentVisual", currentVisual))

	defer func() {
		log.Lgr.Debug("defer", zap.Any("claim", c), zap.Any("current", current), zap.Any("acts", acts))
		// Check if claim differs from current visual (including when visual is nil)
		differs := currentVisual == nil || c.Differs(currentVisual)
		if differs {
			center := tableCB(i, key, c)
			common.Fe.ReplaceTableItem(i, key, center)
		}
	}()

	if current != nil && !c.Differs(current) {
		log.Lgr.Debug("deleting action")
		delete(acts, key)
		if len(acts) == 0 {
			delete(global.Actions, uid)
		}
		return
	}

	if acts == nil {
		acts = common.ClaimsMap{}
		global.Actions[uid] = acts
	}

	acts[key] = &c

	// Check for type change (boolean <-> date) and trigger layout refresh if needed
	if current != nil {
		isCurrentBoolean := current.Date == nil
		isNewBoolean := c.Date == nil
		if isCurrentBoolean != isNewBoolean {
			// Type change detected, refresh layout
			common.Fe.LayoutUsers()
		}
	}
}

// LayoutUsers updates current users with their permissions as checkboxes.
func (f *Frontend) LayoutUsers() {
	f.userTbl.ClearAfter(len(f.userHdrs))
	rows := make([]int, len(global.CrntUsers))
	for i, uid := range global.CrntUsers {
		rows[i] = 1
		n, e, claims := util.FixedUserDetails(uid)
		nt, et := newText(n), newText(e)
		if i%2 == 1 {
			nt.SetBackgroundColor(tview.Styles.PrimaryTextColor)
			nt.SetTextColor(tview.Styles.ContrastBackgroundColor)
			et.SetBackgroundColor(tview.Styles.PrimaryTextColor)
			et.SetTextColor(tview.Styles.ContrastBackgroundColor)
		}
		f.userTbl.AddItem(nt, i+1, 0, 1, 1, 0, 0, false).AddItem(et, i+1, 1, 1, 1, 0, 0, false)
		for j, perm := range common.AllPerms {
			c, ok := claims[perm]
			if !ok || c == nil {
				common.Fe.ShowMsg(fmt.Sprintf("%s: %s, %s", lang.ErrWrongDBClaimS, perm, claims))
				return
			}
			f.userTbl.AddItem(tableCB(i, perm, *c), i+1, j+namedCols, 1, 1, 0, 0, true)
		}
	}
	f.onShowPage[f.CurrentPage()]()
	f.userTbl.SetRows(rows...)
}

func (f *Frontend) ReplaceTableItem(i int, key string, p tview.Primitive) {
	for j, perm := range common.AllPerms {
		if perm == key {
			f.userTbl.ReplaceItemAt(p, i+1, j+namedCols)
			break
		}
	}
}

// newText returns a centered gui element with the given text.
func newText(text string) *tview.TextView {
	return tview.NewTextView().SetTextAlign(tview.AlignCenter).SetText(text)
}
