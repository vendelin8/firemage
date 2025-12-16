package frontend

import (
	"github.com/vendelin8/firemage/internal/common"
	"github.com/vendelin8/firemage/internal/frontend/window"
	"github.com/vendelin8/firemage/internal/global"
	"github.com/vendelin8/firemage/internal/lang"
)

// initPages initializes pages that need it.
func initPages(page string) error {
	if page == lang.PageList {
		return common.Fb.DoList()
	}
	return nil
}

// ShowPage shows the given page.
func ShowPage(newPage string) error {
	oldPage := common.Fe.CurrentPage()
	if newPage == oldPage {
		return nil
	}

	global.SavedUsers[oldPage] = global.CrntUsers // saving users to the closing page
	common.Fe.SetPage(newPage)
	if us, ok := global.SavedUsers[newPage]; ok {
		global.CrntUsers = us
	} else {
		global.CrntUsers = []string{}
		if err := initPages(newPage); err != nil {
			return err
		}
	}
	if len(global.Actions) > 0 && newPage == lang.PageList {
		window.ShowWarningOnce(lang.WarnActionInList)
	}
	common.Fe.LayoutUsers()
	return nil
}
