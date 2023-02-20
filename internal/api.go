package main

import (
	"fmt"

	"github.com/gdamore/tcell"
	"go.uber.org/zap"
)

func showList() {
	showPage(kList)
}

func showSearch() {
	showPage(kSearch)
}

// search looks for users in Firestore with email or name starting with given part.
// Results are loaded into crntUsers uid string list.
func search(searchKey, searchValue string) {
	lgr.Info("searching for", zap.String("key", searchKey), zap.String("value", searchValue))
	if len(searchValue) < minSearchLen {
		fe.showMsg(fmt.Sprintf(errMinLen, minSearchLen))
		return
	}
	if len(actions) > 0 {
		showWarningOnce(wSearchAgain)
	}
	crntUsers = crntUsers[:0]
	fb.searchFor(searchKey, searchValue, func(newUser string) {
		crntUsers = append(crntUsers, newUser)
	})
	if len(crntUsers) == 0 {
		writeErrorStr(warnNoUsers)
	}
	sortByNameThenEmail(crntUsers)
	showErrorsIf()
}

// cancel clears unsaved permission changes.
func cancel() {
	lgr.Info("cancel")
	if len(actions) == 0 {
		showMsg(sNoChanges)
		return
	}
	actions = map[string]map[string]bool{}
	fe.layoutUsers()
}

// refresh refreshes GUI and firestore cache from iterating all firebase auth users.
func refresh() {
	lgr.Info("refresh")
	if fe.currentPage() != kList {
		showMsg(errCantRefresh)
		return
	}
	if len(actions) > 0 {
		showMsg(errActions)
		return
	}
	fb.saveList(actRefresh)
	if len(crntUsers) == 0 {
		writeErrorStr(warnNoUsers)
	}
	if showErrorsIf() {
		showMsg(sNoChanges)
	}
}

// save saves user changes if any.
func save() {
	lgr.Info("save")
	if len(actions) == 0 {
		showMsg(sNoChanges)
		return
	}
	fb.saveList(actSave)
	if showErrorsIf() {
		showMsg(sSaved)
	}
}

// showPage shows the given page.
func showPage(newPage string) {
	lgr.Info("showPage", zap.String("newPage", newPage))
	oldPage := fe.currentPage()
	if newPage == oldPage {
		return
	}

	savedUsers[oldPage] = crntUsers // saving users to the closing page
	fe.setPage(newPage)
	if us, ok := savedUsers[newPage]; ok {
		crntUsers = us
	} else {
		crntUsers = []string{}
		if newPage == kList {
			fb.saveList(actList)
			if len(crntUsers) == 0 {
				writeErrorStr(warnNoUsers)
				writeErrorStr(warnMayRefresh)
			}
		}
	}
	if len(actions) > 0 && newPage == kList {
		showWarningOnce(wActionInList)
	}
	showErrorsIf()
	fe.layoutUsers()
}

// hasPopup returns if there's an active popup window.
func hasPopup() bool {
	return len(activePopup) > 0
}

// quit exists the application.
func quit() {
	lgr.Info("quit")
	if len(actions) == 0 {
		fe.quit()
		return
	}
	showConfirm(fmt.Sprintf(warnUnsaved, len(actions)), fe.quit, nil)
}

// cmdByKey calls the adequate api function through a keyboard shortcut.
func cmdByKey(ev *tcell.EventKey) *tcell.EventKey {
	cmd, ok := shortcuts[ev.Key()]
	if !ok {
		return ev
	}
	if !hasPopup() {
		menuItems[cmd].function()
		return nil
	}
	if cmd != cmdQuit {
		return ev
	}
	hidePopup() // hide popup by esc
	return nil
}
