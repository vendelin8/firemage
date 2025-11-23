package internal

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"go.uber.org/zap"
)

func showList() error {
	showPage(pageLst)
	return nil
}

func showSearch() error {
	showPage(pageSrch)
	return nil
}

// search looks for users in Firestore with email or name starting with given part.
// Results are loaded into crntUsers uid string list.
func search(searchKey, searchValue string) error {
	lgr.Info("searching for", zap.String("key", searchKey), zap.String("value", searchValue))
	if len(searchValue) < minSearchLen {
		return fmt.Errorf(errMinLen, minSearchLen)
	}
	if len(actions) > 0 {
		showWarningOnce(wSearchAgain)
	}
	crntUsers = crntUsers[:0]
	err := searchFor(searchKey, searchValue, func(newUser string) error {
		crntUsers = append(crntUsers, newUser)
		return nil
	})
	if err != nil {
		return fmt.Errorf(errSearch, err)
	}
	if len(crntUsers) == 0 {
		return errNoUsers
	}
	sortByNameThenEmail(crntUsers)
	return nil
}

// cancel clears unsaved permission changes.
func cancel() error {
	lgr.Info("cancel")
	if len(actions) == 0 {
		return errNoChanges
	}
	actions = map[string]map[string]any{}
	fe.layoutUsers()
	return nil
}

// refresh refreshes GUI and firestore cache from iterating all firebase auth users.
func refresh() error {
	lgr.Info("refresh")
	if fe.currentPage() != pageLst {
		return errCantRefresh
	}
	if len(actions) > 0 {
		return errActions
	}
	if err := doRefresh(); err != nil {
		return fmt.Errorf(errRefresh, err)
	}
	if len(crntUsers) == 0 {
		return errNoUsers
	}

	return errNoChanges
}

// save saves user changes if any.
func save() error {
	lgr.Info("save")
	if len(actions) == 0 {
		return errNoChanges
	}

	if err := doSave(); err != nil {
		return fmt.Errorf(errSave, err)
	}

	fe.showMsg(sSaved)

	return nil
}

// showPage shows the given page.
func showPage(newPage string) error {
	lgr.Info("showPage", zap.String("newPage", newPage))
	oldPage := fe.currentPage()
	if newPage == oldPage {
		return nil
	}

	savedUsers[oldPage] = crntUsers // saving users to the closing page
	fe.setPage(newPage)
	if us, ok := savedUsers[newPage]; ok {
		crntUsers = us
	} else {
		crntUsers = []string{}
		if newPage == pageLst {
			if err := doList(); err != nil {
				return fmt.Errorf(errList, err)
			}
			if len(crntUsers) == 0 {
				return fmt.Errorf("%s %s", errNoUsersStr, warnMayRefresh)
			}
		}
	}
	if len(actions) > 0 && newPage == pageLst {
		showWarningOnce(wActionInList)
	}
	fe.layoutUsers()
	return nil
}

// hasPopup returns if there's an active popup window.
func hasPopup() bool {
	return len(activePopup) > 0
}

// quit exists the application.
func quit() error {
	lgr.Info("quit")
	if len(actions) == 0 {
		fe.quit()
		return nil
	}

	showConfirm(fmt.Sprintf(warnUnsaved, len(actions)), fe.quit, nil)
	return nil
}

// cmdByKey calls the adequate api function through a keyboard shortcut.
func cmdByKey(ev *tcell.EventKey) *tcell.EventKey {
	cmd, ok := shortcuts[ev.Key()]
	if !ok {
		return ev
	}
	if !hasPopup() {
		if err := menuItems[cmd].function(); err != nil {
			fe.showMsg(err.Error())
		}
		return nil
	}
	if cmd != cmdQuit {
		return ev
	}
	hidePopup() // hide popup by esc
	return nil
}
