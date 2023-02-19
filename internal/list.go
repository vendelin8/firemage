package main

import "github.com/vendelin8/tview"

func (f *Frontend) initList() {
	fe.setOnShow(kList, func() {
		f.listPage.ResizeItemAt(0, len(crntUsers)+1, 0)
	})
	f.listPage = tview.NewFlex().SetDirection(tview.FlexRow).AddItem(f.userTbl, 1, 0, true).
		AddItem(f.filler, 0, 1, false)
}
