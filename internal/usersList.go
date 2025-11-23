package internal

import (
	"maps"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/vendelin8/tview"
)

const namedCols = 2 // name and email column

func (f *Frontend) initUsersList() {
	colNum := len(allPerms) + namedCols
	f.userHdrs = make([]string, colNum)
	f.userHdrs[0] = sName
	f.userHdrs[1] = sEmail
	colSizes := make([]int, colNum)
	colSizes[0] = 25
	colSizes[1] = 0 // fill
	for i, perm := range allPerms {
		j := i + namedCols
		f.userHdrs[j] = permsMap[perm]
		colSizes[j] = len(perm) + 2 // checkbox padding
	}
	for col, text := range f.userHdrs {
		f.userTbl.AddItem(newText(text), 0, col, 1, 1, 0, 0, false)
	}
	f.userTbl.SetColumns(colSizes...)
}

// activatePopup is called when an empty checkbox is checked. It pops up a dialog
// to change value to true or a given date.
func activatePopup(i int, key string, checked bool, date *time.Time) {
	// TODO: show a popup with a checkbox and a date picker
	// fill in the checkbox or the date by the parameter
	// buttons with "add ..." based on custom template
}

// tableCB returns a checkbox to the claim table filled with the saved value.
func tableCB(i int, key string, claims map[string]any) tview.Primitive {
	checked := false
	var date *time.Time
	if cbv, ok := claims[key]; ok {
		if b, ok := cbv.(bool); ok {
			checked = b
		} else {
			dateVal := cbv.(time.Time)
			date = &dateVal
		}
	}

	var (
		box tview.FormItem
		bgc = tview.Styles.PrimitiveBackgroundColor
		ftc = tview.Styles.PrimaryTextColor
	)

	if i%2 == 1 {
		bgc = tview.Styles.PrimaryTextColor
		ftc = tview.Styles.PrimitiveBackgroundColor
	}

	if date == nil {
		cb := tview.NewCheckbox().SetChangedFunc(func(checked bool) {
			if checked {
				activatePopup(i, key, false, nil)
			} else {
				onActionChange(i, key, checked, nil)
			}
		}).SetChecked(checked).
			SetFieldTextColor(ftc)
		cb.SetBackgroundColor(bgc)
		box = cb
	} else {
		tv := tview.NewTextView().
			SetText(date.Format(dateFormat)).
			SetTextColor(ftc)
		tv.SetBackgroundColor(bgc).
			SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				activatePopup(i, key, false, date)
				return nil
			}).SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
			activatePopup(i, key, false, date)
			return action, nil
		})
		box = tv
	}

	c := tview.NewCenter(box, box.GetFieldWidth(), box.GetFieldHeight())
	if i%2 == 1 {
		c.SetBackgroundColor(tview.Styles.PrimaryTextColor)
	}
	return c
}

// onActionChange in called when a user claim is changed by checkbox or date field.
func onActionChange(i int, key string, checked bool, date *time.Time) {
	uid := crntUsers[i]
	current := localUsers[uid].Claims[key]
	acts, ok := actions[uid]

	if date != nil && current == *date || (current != nil) == checked {
		delete(acts, key)
		if len(acts) == 0 {
			delete(actions, uid)
		}
		return
	}

	if !ok && len(acts) == 0 {
		acts = map[string]any{}
		actions[uid] = acts
	}

	if date != nil {
		acts[key] = *date
		return
	}

	acts[key] = checked
}

// fixedUserClaims returns user name, email and claims with applied actions.
func fixedUserClaims(uid string) (string, string, map[string]any) {
	u := localUsers[uid]
	claims := u.Claims
	if ac, ok := actions[uid]; ok {
		claims = maps.Clone(u.Claims)
		for k, v := range ac {
			claims[k] = v
		}
	}
	return u.Name, u.Email, claims
}

// layoutUsers updates current users with their permissions as checkboxes.
func (f *Frontend) layoutUsers() {
	f.userTbl.ClearAfter(len(f.userHdrs))
	rows := make([]int, len(crntUsers))
	for i, uid := range crntUsers {
		rows[i] = 1
		n, e, claims := fixedUserClaims(uid)
		nt, et := newText(n), newText(e)
		if i%2 == 1 {
			nt.SetBackgroundColor(tview.Styles.PrimaryTextColor)
			nt.SetTextColor(tview.Styles.ContrastBackgroundColor)
			et.SetBackgroundColor(tview.Styles.PrimaryTextColor)
			et.SetTextColor(tview.Styles.ContrastBackgroundColor)
		}
		f.userTbl.AddItem(nt, i+1, 0, 1, 1, 0, 0, false).AddItem(et, i+1, 1, 1, 1, 0, 0, false)
		for j, perm := range allPerms {
			f.userTbl.AddItem(tableCB(i, perm, claims), i+1, j+namedCols, 1, 1, 0, 0, true)
		}
	}
	f.onShowPage[f.currentPage()]()
	f.userTbl.SetRows(rows...)
}

// newText returns a centered gui element with the given text.
func newText(text string) *tview.TextView {
	return tview.NewTextView().SetTextAlign(tview.AlignCenter).SetText(text)
}
