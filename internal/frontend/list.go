package frontend

import (
	"github.com/vendelin8/firemage/internal/global"
	"github.com/vendelin8/firemage/internal/lang"
	"github.com/vendelin8/tview"
)

func (f *Frontend) initList() {
	f.SetOnShow(lang.PageList, func() {
		f.listPage.ResizeItemAt(0, len(global.CrntUsers)+1, 0)
	})
	f.listPage = tview.NewFlex().SetDirection(tview.FlexRow).AddItem(f.userTbl, 1, 0, true).
		AddItem(f.filler, 0, 1, false)
}
