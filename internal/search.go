package main

import (
	"fmt"

	"github.com/vendelin8/tview"
)

const (
	kName  = "name"
	kEmail = "email"
)

func (f *Frontend) initSearch() {
	f.searchForEmail()
	f.app.SetFocus(f.searchField)

	form := tview.NewForm().AddFormItem(f.searchRadio).AddFormItem(f.searchField)
	form.AddButton(sDoSearch, func() {
		search(f.searchFieldName[f.searchRadio.Value()], f.searchField.GetText())
		fe.layoutUsers()
	})

	f.setOnShow(kSearch, func() {
		f.searchPage.ResizeItemAt(1, len(crntUsers)+1, 0)
	})
	f.searchPage = tview.NewFlex().SetDirection(tview.FlexRow).AddItem(form, 7, 0, true).
		AddItem(f.userTbl, 1, 0, true).AddItem(f.filler, 0, 1, false)
}

func (f *Frontend) searchForEmail() {
	f.searchField.SetLabel(fmt.Sprintf("%s: ", sEmail))
}

func (f *Frontend) searchForName() {
	f.searchField.SetLabel(fmt.Sprintf("%s: ", sName))
}
