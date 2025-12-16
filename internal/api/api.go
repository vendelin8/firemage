package api

import (
	"errors"
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/vendelin8/firemage/internal/common"
	"github.com/vendelin8/firemage/internal/conf"
	"github.com/vendelin8/firemage/internal/firebase"
	"github.com/vendelin8/firemage/internal/frontend"
	"github.com/vendelin8/firemage/internal/frontend/window"
	"github.com/vendelin8/firemage/internal/global"
	"github.com/vendelin8/firemage/internal/lang"
)

var (
	ErrActions     = errors.New(lang.ErrActionsS)
	ErrNoChanges   = errors.New(lang.ErrNoChangesS)
	ErrCantRefresh = errors.New(lang.ErrCantRefreshS)
)

func InitMenu() {
	common.MenuItems = map[int]common.MenuItem{
		conf.CmdCancel:  {Shortcut: "F8", Keys: []tcell.Key{tcell.KeyF8}, MenuKey: "", Text: lang.MenuCancel, Positive: false, IsDef: true, Function: cancel},
		conf.CmdRefresh: {Shortcut: "F5", Keys: []tcell.Key{tcell.KeyF5}, MenuKey: "", Text: lang.MenuRefresh, Positive: true, IsDef: true, Function: refresh},
		conf.CmdSearch:  {Shortcut: "F2", Keys: []tcell.Key{tcell.KeyF2}, MenuKey: lang.PageSearch, Text: lang.Titles[lang.PageSearch], Positive: false, IsDef: true, Function: showSearch},
		conf.CmdList:    {Shortcut: "F3", Keys: []tcell.Key{tcell.KeyF3}, MenuKey: lang.PageList, Text: lang.Titles[lang.PageList], Positive: false, IsDef: true, Function: showList},
		conf.CmdSave:    {Shortcut: "F6", Keys: []tcell.Key{tcell.KeyF6}, MenuKey: "", Text: lang.MenuSave, Positive: true, IsDef: true, Function: save},
		conf.CmdQuit:    {Shortcut: "Esc", Keys: []tcell.Key{tcell.KeyEsc}, MenuKey: "", Text: lang.MenuQuit, Positive: false, IsDef: true, Function: window.Quit},
	}
}

func showList() error {
	return frontend.ShowPage(lang.PageList)
}

func showSearch() error {
	return frontend.ShowPage(lang.PageSearch)
}

// cancel clears unsaved permission changes.
func cancel() error {
	if len(global.Actions) == 0 {
		return ErrNoChanges
	}
	global.Actions = map[string]common.ClaimsMap{}
	common.Fe.LayoutUsers()
	return nil
}

// refresh refreshes GUI and firestore cache from iterating all firebase auth users.
func refresh() error {
	if common.Fe.CurrentPage() != lang.PageList {
		return ErrCantRefresh
	}
	if len(global.Actions) > 0 {
		return ErrActions
	}

	// Run refresh operation asynchronously so UI can redraw the progress popup
	go func() {
		if err := firebase.DoRefresh(); err != nil {
			window.ShowErrorBuffer(fmt.Errorf(lang.ErrRefresh, err))
			return
		}
		if len(global.CrntUsers) == 0 {
			window.ShowErrorBuffer(common.ErrNoUsers)
		}
	}()

	return nil
}

// save saves user changes if any.
func save() error {
	if len(global.Actions) == 0 {
		return ErrNoChanges
	}

	// Run save operation asynchronously so UI can redraw the progress popup
	go func() {
		if err := firebase.DoSave(); err != nil {
			window.ShowErrorBuffer(fmt.Errorf(lang.ErrSave, err))
			return
		}
		common.Fe.ShowMsg(lang.SSaved)
	}()

	return nil
}
