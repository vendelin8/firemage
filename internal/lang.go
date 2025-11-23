package internal

const (
	menuRefresh = "Refresh"
	menuSave    = "Save"
	menuCancel  = "Cancel"
	menuQuit    = "Quit"
	shortDesc   = "Firebase auth admin"
	longDesc    = "firemage is a CLI tool to manage Firebase Auth Claims, written in Golang"
	descKey     = "Google service account file path"
	descConf    = "config file path"
	descDebug   = "to print debug info"
	descEmul    = "if use local firebase emulator"
	sName       = "Name"
	sEmail      = "Email"
	sSearchThis = "Search this:"
	sDoSearch   = "Search"
	sYes        = "Yes"
	sNo         = "No"
	sSaved      = "Saved"
	cShortcuts  = "keyboardShortcuts"
	warnUnsaved = "You have %d unsaved actions. Are you sure to quit?"

	errList   = "listing users failed: %w"
	errSave   = "save failed: %w"
	errNew    = "new privileged user(s): %s"
	errEmpty  = "missing user(s) from privileged ones: %s"
	errMinLen = "insert at least %d characters"
	errSearch = "failed to get search results: %w"

	errRefresh    = "refresh failed: %w"
	errChanged    = "changed permissions or new user(s): %s"
	errRemoved    = "the following user(s) were deleted from the system: %s"
	errConfPath   = "config file not found, please check application arguments: %w"
	errConfParse  = "error while parsing config file: %s"
	errManualStr  = "anyone touched the claims or the database manually?"
	errTimeoutStr = "database access timeout"
	errGetFSUsers = "failed to get users from database: %w"
	errNoUsersStr = "no such user found"
	errActionsStr = "you have to save or cancel the current changes"

	warnMayRefresh    = "consider a refresh"
	errCmdNotFound    = "not found keyboard command(s): %s"
	errKeyNotFound    = "not found keyboard shortcut(s): %s"
	errGetAuthUsers   = "failed to get users from auth: %w"
	errNoChangesStr   = "no changes"
	errUpdateFSUsers  = "failed to update users in database: %w"
	errConfInvalidStr = "config file is invalid"
	errCantRefreshStr = "refresh is possible only on List page"
)

var (
	titles = map[string]string{pageSrch: sDoSearch, pageLst: "List"}
	warns  = map[int]string{
		wSearchAgain:  "Your changes stay there from your recent searches. To remove them, click on Cancel.",
		wActionInList: "Your recent changes stay there. If you added permissions while searching, you'll only see them here after Save.",
	}
)
