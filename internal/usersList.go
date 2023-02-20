package main

import (
	"github.com/vendelin8/tview"
)

var (
	localUsers = map[string]*User{}
	crntUsers  []string
)

func (f *Frontend) initUsersList() {
	colNum := len(kAllPerms) + 2
	f.userHdrs = make([]string, colNum)
	f.userHdrs[0] = sName
	f.userHdrs[1] = sEmail
	colSizes := make([]int, colNum)
	colSizes[0] = 25
	colSizes[1] = 0 // fill
	for i, perm := range kAllPerms {
		j := i + 2
		f.userHdrs[j] = kPermsMap[perm]
		colSizes[j] = len(perm) + 2
	}
	for col, text := range f.userHdrs {
		f.userTbl.AddItem(newText(text), 0, col, 1, 1, 0, 0, false)
	}
	f.userTbl.SetColumns(colSizes...)
}

// tableCB returns a checkbox to the claim table filled with the saved value.
func tableCB(i int, key string, claims map[string]bool) tview.Primitive {
	checked := false
	if cbv, ok := claims[key]; ok {
		checked = cbv
	}
	cb := tview.NewCheckbox().SetChangedFunc(func(checked bool) {
		onActionChange(checked, i, key)
	}).SetChecked(checked)
	c := tview.NewCenter(cb, 3, 1)
	if i%2 == 1 {
		cb.SetBackgroundColor(tview.Styles.PrimaryTextColor)
		cb.SetFieldTextColor(tview.Styles.PrimitiveBackgroundColor)
		c.SetBackgroundColor(tview.Styles.PrimaryTextColor)
	}
	return c
}

func onActionChange(checked bool, i int, key string) {
	uid := crntUsers[i]
	ai, ok := actions[uid]
	if !ok {
		actions[uid] = map[string]bool{key: checked}
		return
	}
	if _, ok = ai[key]; !ok {
		ai[key] = checked
		return
	}
	delete(ai, key)
	if len(ai) == 0 {
		delete(actions, uid)
	}
}

// fixedUserClaims returns user name, email and claims with applied actions.
func fixedUserClaims(uid string) (string, string, map[string]bool) {
	u := localUsers[uid]
	claims := u.Claims
	if ac, ok := actions[uid]; ok {
		claims = copyClaims(u.Claims)
		for k, v := range ac {
			claims[k] = v
		}
	}
	return u.Name, u.Email, claims
}

// layoutUsers updates current users with their permisions as checkboxes.
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
		for j, perm := range kAllPerms {
			f.userTbl.AddItem(tableCB(i, perm, claims), i+1, j+2, 1, 1, 0, 0, true)
		}
	}
	f.onShowPage[f.currentPage()]()
	f.userTbl.SetRows(rows...)
}

// newText returns a centered gui element with the given text.
func newText(text string) *tview.TextView {
	return tview.NewTextView().SetTextAlign(tview.AlignCenter).SetText(text)
}
