package internal

import (
	"context"
	"errors"

	"github.com/gdamore/tcell/v2"
	"go.uber.org/zap"
)

var (
	// shortcuts defines what to call in case of a keyboard shortcut press.
	shortcuts = map[tcell.Key]int{}
	// menuItems are the default menu items.
	menuItems = map[int]*menuItem{
		cmdCancel:  {"F8", []tcell.Key{tcell.KeyF8}, "", menuCancel, false, true, cancel},
		cmdRefresh: {"F5", []tcell.Key{tcell.KeyF5}, "", menuRefresh, true, true, refresh},
		cmdSearch:  {"F2", []tcell.Key{tcell.KeyF2}, pageSrch, titles[pageSrch], false, true, showSearch},
		cmdList:    {"F3", []tcell.Key{tcell.KeyF3}, pageLst, titles[pageLst], false, true, showList},
		cmdSave:    {"F6", []tcell.Key{tcell.KeyF6}, "", menuSave, true, true, save},
		cmdQuit:    {"Esc", []tcell.Key{tcell.KeyEsc}, "", menuQuit, false, true, quit},
	}
	es          = struct{}{} // empty struct
	fb          FbIf         // firebase global instance
	fe          FeIf         // global frontend instance
	cancelF     context.CancelFunc
	activePopup string
	confPath    string
	keyPath     string
	verbose     bool
	useEmu      bool
	lgr         *zap.Logger
	localUsers  = map[string]*User{} // downloaded users
	crntUsers   []string             // currently visible users
	// localPrivileged is the downloaded privileged users.
	localPrivileged = map[string]struct{}{}
	// actions contains permission updates to be saved. key is the uid, value is a map of permission
	// key and a value. True means adding the permission, false means removing it, date means expiry.
	actions = map[string]map[string]any{}
	// savedUsers contains user id lists for all pages.
	savedUsers = map[string][]string{}
)

var (
	errEnd     = errors.New("end")
	errManual  = errors.New(errManualStr)
	errTimeout = errors.New(errTimeoutStr)
	errNoUsers = errors.New(errNoUsersStr)
	errActions = errors.New(errActionsStr)

	errNoChanges   = errors.New(errNoChangesStr)
	errConfInvalid = errors.New(errConfInvalidStr)
	errCantRefresh = errors.New(errCantRefreshStr)
)
