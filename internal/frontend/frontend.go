package frontend

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/vendelin8/firemage/internal/common"
	"github.com/vendelin8/firemage/internal/conf"
	"github.com/vendelin8/firemage/internal/frontend/window"
	"github.com/vendelin8/firemage/internal/lang"
	"github.com/vendelin8/firemage/internal/log"
	"github.com/vendelin8/firemage/internal/util"
	"github.com/vendelin8/tview"
	"go.uber.org/zap"
)

const ok = "OK"

const (
	claimActive = iota
	claimInactive
	claimTimed
)

type ClaimsModal struct {
	*tview.FormModal
	radio *tview.Radio
	date  *tview.InputField
	onOK  func(c common.Claim)
}

type Frontend struct {
	msg      *tview.Modal
	confirm  *tview.Modal
	claims   *ClaimsModal
	progress *tview.Modal
	pages    *tview.Pages
	menu     *tview.TextView
	filler   *tview.Box
	app      *tview.Application
	header   *tview.TextView
	userHdrs []string
	userTbl  *tview.Grid

	listPage   *tview.Flex
	searchPage *tview.Flex

	searchField *tview.InputField
	searchRadio *tview.Radio
	onShowPage  map[string]func()

	searchFieldName map[int]string
}

func (f *Frontend) CurrentPage() string {
	hs := f.menu.GetHighlights()
	if len(hs) > 0 {
		return hs[0]
	}
	return ""
}

// Quit exists the app. If there are unsaved changes, the user needs to confirm to loose them.
func (f *Frontend) Quit() {
	f.app.Stop()
}

func (f *Frontend) SetOnShow(page string, cb func()) {
	f.onShowPage[page] = cb
}

func CreateGUI() *Frontend {
	f := &Frontend{}
	f.searchRadio = tview.NewRadio(lang.SEmail, lang.SName).SetOnSetValue(func(radioValue int) {
		if radioValue == 0 {
			f.searchForEmail()
		} else {
			f.searchForName()
		}
	}).SetLabel(lang.SSearchThis).SetHorizontal(true)
	f.onShowPage = map[string]func(){}
	f.filler = tview.NewBox()
	f.app = tview.NewApplication()
	f.header = newText("")
	f.searchField = tview.NewInputField().SetFieldWidth(40)
	f.searchFieldName = map[int]string{0: "email", 1: "name"}
	f.userTbl = tview.NewGrid()
	f.menu = tview.NewTextView().SetDynamicColors(true).SetRegions(true).SetWrap(false)
	f.initUsersList()
	return f
}

func (f *Frontend) formatMenuItem(w io.Writer, menuKey, text, shortcut string, isPositive bool) {
	var color string
	switch {
	case len(menuKey) > 0:
		color = "yellow"
	case isPositive:
		color = "green"
	default:
		color = "red"
	}
	fmt.Fprintf(w, ` %s ["%s"][%s::b]%s[white::-][""]  `, shortcut, menuKey, color, text)
}

func (f *Frontend) Run() {
	err := conf.InitConf(func(menuKey, text, shortcut string, isPositive bool) {
		f.formatMenuItem(f.menu, menuKey, text, shortcut, isPositive)
	})
	log.Must("init configuration", err)

	f.initSearch()
	f.initList()
	f.pages = tview.NewPages().AddPage(lang.PageSearch, f.searchPage, true, false).
		AddPage(lang.PageList, f.listPage, true, false)
	layout := tview.NewFlex().SetDirection(tview.FlexRow).AddItem(f.header, 1, 0, false).
		AddItem(f.pages, 0, 1, true).AddItem(f.menu, 1, 0, false)
	f.app.SetInputCapture(CmdByKey)
	f.app.SetRoot(layout, true).EnableMouse(true)

	err = ShowPage(lang.PageSearch)
	log.Must("showing initial page", err)

	log.Must("run app", f.app.Run())
}

func extractMsg(ms ...string) string {
	if len(ms) > 0 {
		return ms[0]
	}
	return window.PopBuffer()
}

// ShowMsg shows the given message as a popup.
func (f *Frontend) ShowMsg(ms ...string) {
	window.ActivePopups = append(window.ActivePopups, lang.PopupMsg)
	m := extractMsg(ms...)
	if f.msg != nil {
		f.pages.ShowPage(lang.PopupMsg)
		f.msg.SetText(m)
		return
	}
	f.msg = tview.NewModal().AddButtons([]string{ok}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			window.HidePopup(lang.PopupMsg)
		})
	f.pages.AddPage(lang.PopupMsg, f.msg, true, true)
	f.msg.SetText(m)
}

// ShowConfirm shows a confirm dialog with a text, and callback functions for OK and Cancel.
func (f *Frontend) ShowConfirm(onYes, onNo func(), ms ...string) {
	m := extractMsg(ms...)
	if f.confirm != nil {
		f.confirm.SetText(m)
		f.pages.ShowPage(lang.PopupConfirm)
		return
	}
	f.confirm = tview.NewModal().AddButtons([]string{lang.SYes, lang.SNo}).
		SetDoneFunc(window.ConfirmDoneFunc(onYes, onNo))
	f.pages.AddPage(lang.PopupConfirm, f.confirm, true, true)
	f.confirm.SetText(m)
}

func (f *Frontend) ClaimButtonSetDisabled(index int, isDisabled bool) {
	f.claims.GetButton(index).SetDisabled(isDisabled)
}

func claimsFieldSetChangedFunc(text string) {
	_, err := time.Parse(common.DateFormat, text)
	isDisabled := err != nil
	common.Fe.ClaimsBtns(isDisabled)
	common.Fe.ClaimButtonSetDisabled(len(common.TimedButtons), isDisabled)
}

func (f *Frontend) ClaimsBtns(isDisabled bool) {
	for i := range len(common.TimedButtons) {
		common.Fe.ClaimButtonSetDisabled(i, isDisabled)
	}
}

func (f *Frontend) ClaimsDateSetDisabled(isDisabled bool) {
	f.claims.date.SetDisabled(isDisabled)
}

func (f *Frontend) ClaimsSetDate(value time.Time) {
	f.claims.date.SetText(value.Format(common.DateFormat))
}

func (f *Frontend) ClaimsDate() *time.Time {
	dateStr := f.claims.date.GetText()
	if len(dateStr) == 0 {
		return nil
	}

	d, err := time.Parse(common.DateFormat, dateStr)
	if err != nil {
		log.Lgr.Error("claim date parse", zap.Error(err))
		return nil
	}

	return &d
}

const (
	ClaimActive = iota
	ClaimInactive
	ClaimTimed
)

func claimsRadioSetOnSetValue(radioValue int) {
	switch radioValue {
	case ClaimActive:
		common.Fe.ClaimsDateSetDisabled(true)
		common.Fe.ClaimsBtns(true)
	case ClaimInactive:
		common.Fe.ClaimsDateSetDisabled(true)
		common.Fe.ClaimsBtns(true)
	case ClaimTimed:
		common.Fe.ClaimsDateSetDisabled(false)
		common.Fe.ClaimsBtns(false)
		if common.Fe.ClaimsDate() == nil {
			common.Fe.ClaimsSetDate(time.Now())
		}
	}
}

func (c *ClaimsModal) handleOK() {
	c.processClaimResult()
	window.HidePopup(lang.PopupClaim)
}

func (c *ClaimsModal) processClaimResult() {
	log.Lgr.Debug("processClaimResult", zap.Int("radio", c.radio.Value()))
	switch c.radio.Value() {
	case claimActive:
		c.onOK(common.Claim{Checked: true})
	case claimInactive:
		c.onOK(common.Claim{})
	case claimTimed:
		d, err := time.Parse(common.DateFormat, c.date.GetText())
		if err != nil {
			log.Lgr.Error("claim date parse", zap.Error(err))
			c.onOK(common.Claim{Checked: false})
			return
		}
		c.onOK(common.Claim{Date: &d})
	}
}

func (f *Frontend) CreateClaimChoser() {
	width, height := 50, 10
	// TODO: replace with DatePicker (struct to be created)
	dateF := tview.NewInputField().SetChangedFunc(claimsFieldSetChangedFunc)
	radio := tview.NewRadio(lang.SActive, lang.SInactive, lang.STimed).SetHorizontal(true).SetOnSetValue(claimsRadioSetOnSetValue)
	c := &ClaimsModal{radio: radio, date: dateF}
	c.FormModal = tview.NewFormModal(func(form *tview.Form) {
		form.AddFormItem(radio)
		form.AddFormItem(dateF)
		for _, items := range common.TimedButtons {
			form.AddButton(items[0], func() {
				c.incDate(items[0])
			})
		}
		form.AddButton(ok, func() {
			c.handleOK()
		})
		form.AddButton(lang.SCancel, func() {
			window.HidePopup(lang.PopupClaim)
		})
	})

	f.pages.AddPage(lang.PopupClaim, tview.NewCenter(c, width, height), true, true)
	f.claims = c
}

func (c *ClaimsModal) incDate(buttonLabel string) {
	d := common.Fe.ClaimsDate()
	if d == nil {
		now := time.Now()
		d = &now
	}

	newD := util.AddTimedDate(*d, buttonLabel)
	common.Fe.ClaimsSetDate(newD)
	c.radio.SetValue(ClaimTimed)
}

// ShowClaimChoser shows a claim chooser dialog with options for active, inactive, or timed claims.
func (f *Frontend) ShowClaimChoser(c common.Claim, onOK func(c common.Claim)) {
	var (
		claimType int
		dateStr   string
	)

	if c.Date != nil {
		claimType = ClaimTimed
		dateStr = c.Date.Format(common.DateFormat)
	} else {
		dateStr = time.Now().Format(common.DateFormat)
		claimType = ClaimInactive
	}

	creating := f.claims == nil
	if creating {
		common.Fe.CreateClaimChoser()
	}

	claims := f.claims
	claims.date.SetText(dateStr)
	// f.app.SetFocus(claims.SetFocus(0))
	radio := f.claims.radio
	radio.SetValue(claimType)
	claims.onOK = onOK
	radio.SetValue(claimTimed)

	if !creating {
		f.pages.ShowPage(lang.PopupClaim)
	}
}

// ShowConfirm shows a confirm dialog with a text, and callback functions for OK and Cancel.
func (f *Frontend) ShowProgress(ctx context.Context, cancelF context.CancelFunc, ms ...string) {
	go func() {
		<-ctx.Done()
		f.app.QueueUpdate(func() {
			window.HidePopup(lang.PopupProgress)
			f.app.ForceDraw()
		})
		f.app.Draw()
	}()

	var m string
	if len(ms) > 0 {
		m = ms[0]
	} else {
		m = lang.SWorking
	}

	if f.progress != nil {
		f.pages.ShowPage(lang.PopupProgress)
		// f.app.SetFocus(f.progress.SetFocus(0))
		f.progress.SetDoneFunc(func(_ int, _ string) {
			cancelF()
		})
		f.progress.SetText(m)
		return
	}

	f.progress = tview.NewModal().AddButtons([]string{lang.SCancel}).
		SetDoneFunc(func(_ int, _ string) {
			cancelF()
		})
	f.pages.AddPage(lang.PopupProgress, f.progress, true, true)
	f.progress.SetText(m)
}

// HidePopup hides the current popup window.
func (f *Frontend) HidePopup(popup string) {
	log.Lgr.Debug("Frontend HidePopup", zap.String("popup", popup))
	f.pages.HidePage(popup)
}

func (f *Frontend) SetPage(newPage string) {
	f.menu.Highlight(newPage).ScrollToHighlight()
	f.pages.SwitchToPage(newPage)
	f.header.SetText(fmt.Sprintf("%s - %s", lang.ShortDesc, lang.Titles[newPage]))
}

// CmdByKey calls the adequate api function through a keyboard shortcut.
func CmdByKey(ev *tcell.EventKey) *tcell.EventKey {
	cmd, ok := common.Shortcuts[ev.Key()]
	if !ok {
		return ev
	}

	if !window.HasPopup() {
		menuItem, exists := common.MenuItems[cmd]
		if !exists {
			return nil
		}

		window.ShowErrorBuffer(menuItem.Function())

		return nil
	}

	if cmd != conf.CmdQuit {
		return ev
	}

	window.HidePopup(window.ActivePopups[len(window.ActivePopups)-1]) // hide popup by esc

	return nil
}
