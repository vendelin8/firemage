package main

import (
	"fmt"

	"github.com/vendelin8/tview"
)

const (
	kMsg    = "msg"
	kCnfrm  = "confirm"
	kSearch = "search"
	kList   = "list"
)

var (
	fe          FeIf
	activePopup string
)

// FeIf is an interface to be able to mock tview GUI functionality.
type FeIf interface {
	run()
	currentPage() string
	setPage(string)
	setOnShow(string, func())
	showMsg(m string)
	showConfirm(m string, okFunc, cancelFunc func())
	hidePopup()
	quit()
}

type Frontend struct {
	msg, confirm    *tview.Modal
	pages           *tview.Pages
	menu            *tview.TextView
	onShowPage      map[string]func()
	filler          *tview.Box
	app             *tview.Application
	header          *tview.TextView
	listPage        *tview.Flex
	searchPage      *tview.Flex
	searchField     *tview.InputField
	searchFieldName map[int]string
	searchRadio     *tview.Radio
	userHdrs        []string
	userTbl         *tview.Grid
}

func (f *Frontend) currentPage() string {
	hs := f.menu.GetHighlights()
	if len(hs) > 0 {
		return hs[0]
	}
	return ""
}

// quit exists the app. If there are unsaved changes, the user needs to confirm to loose them.
func (f *Frontend) quit() {
	f.app.Stop()
}

func (f *Frontend) setOnShow(page string, cb func()) {
	f.onShowPage[page] = cb
}

func createGUI() *Frontend {
	f := &Frontend{}
	f.searchRadio = tview.NewRadio(sEmail, sName).SetOnSetValue(func(radioValue int) {
		if radioValue == 0 {
			f.searchForEmail()
		} else {
			f.searchForName()
		}
	}).SetLabel(sSearchThis).SetHorizontal(true)
	f.onShowPage = map[string]func(){}
	f.filler = tview.NewBox()
	f.app = tview.NewApplication()
	f.header = newText("")
	f.searchField = tview.NewInputField().SetFieldWidth(40)
	f.searchFieldName = map[int]string{0: kEmail, 1: kName}
	f.userTbl = tview.NewGrid()
	f.pages = tview.NewPages().AddPage(kSearch, f.searchPage, true, false).
		AddPage(kList, f.listPage, true, false)

	f.menu = tview.NewTextView().SetDynamicColors(true).SetRegions(true).SetWrap(false)
	initConf(func(shortcut, menuKey, text string, isPositive bool) {
		var color string
		switch {
		case len(menuKey) > 0:
			color = "yellow"
		case isPositive:
			color = "green"
		default:
			color = "red"
		}
		fmt.Fprintf(f.menu, ` %s ["%s"][%s::b]%s[white::-][""]  `, shortcut, menuKey, color, text)
	})
	// fmt.Fprintf(f.menu, ` F2 ["%s"][yellow::b]%s[white::-][""]  `, kSearch, titles[kSearch])
	// fmt.Fprintf(f.menu, ` F3 ["%s"][yellow::b]%s[white::-][""]  `, kList, titles[kList])
	// fmt.Fprintf(f.menu, ` F5 [""][green::b]%s[white::-][""]  `, menuRefresh)
	// fmt.Fprintf(f.menu, ` F6 [""][green::b]%s[white::-][""]  `, menuSave)
	// fmt.Fprintf(f.menu, ` F8 [""][red::b]%s[white::-][""]  `, menuCancel)
	// fmt.Fprintf(f.menu, ` Esc [""][red::b]%s[white::-][""]  `, menuQuit)
	f.initSearch()
	f.initList()
	f.initUsersList()
	return f
}

func (f *Frontend) run() {
	layout := tview.NewFlex().SetDirection(tview.FlexRow).AddItem(f.header, 1, 0, false).
		AddItem(f.pages, 0, 1, true).AddItem(f.menu, 1, 0, false)
	f.app.SetInputCapture(cmdByKey)
	showPage(kSearch)
	must("run app", f.app.SetRoot(layout, true).EnableMouse(true).Run())
}

// showMsg shows the given message as a popup.
func (f *Frontend) showMsg(m string) {
	if f.msg != nil {
		f.pages.ShowPage(kMsg)
		f.msg.SetText(m)
		return
	}
	width, height := 50, 10 // TODO: resize based on text size
	f.msg = tview.NewModal().AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			hidePopup()
		})
	f.pages.AddPage(kMsg, tview.NewCenter(f.msg, width, height), true, true)
	f.msg.SetText(m)
}

// showConfirm shows a confirm dialog with a text, and callback functions for OK and Cancel.
func (f *Frontend) showConfirm(m string, okFunc, cancelFunc func()) {
	if f.confirm != nil {
		f.pages.ShowPage(kCnfrm)
		f.app.SetFocus(f.confirm.SetFocus(0))
		f.confirm.SetText(m)
		return
	}
	width, height := 50, 10 // TODO: resize based on text size
	f.confirm = tview.NewModal().AddButtons([]string{sYes, sNo}).
		SetDoneFunc(confirmDoneFunc(okFunc, cancelFunc))
	f.pages.AddPage(kCnfrm, tview.NewCenter(f.confirm, width, height), true, true)
	f.confirm.SetText(m)
}

// hidePopup hides the current popup window.
func (f *Frontend) hidePopup() {
	f.pages.HidePage(activePopup)
}

func (f *Frontend) setPage(newPage string) {
	f.menu.Highlight(newPage).ScrollToHighlight()
	f.pages.SwitchToPage(newPage)
	f.header.SetText(fmt.Sprintf("%s - %s", mainTitle, titles[newPage]))
	f.layoutUsers(newPage)
}
