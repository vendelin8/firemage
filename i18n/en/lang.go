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

	errList   = "Lisint users failed: %w"
	errSave   = "Save failed: %w"
	errNew    = "New privileged user(s): %s ."
	errEmpty  = "Missing user(s) from privileged ones: %s ."
	errMinLen = "Insert at least %d characters!"
	errSearch = "Failed to get search results: %w"

	errRefresh    = "Refresh failed: %w"
	errChanged    = "Changed permissions or new user(s): %s ."
	errRemoved    = "The following user(s) were deleted from the system: %s ."
	errConfPath   = "Config file not found, please check application arguments: %w."
	errConfParse  = "Error while parsing config file: %s ."
	errManualStr  = "Anyone touched the claims or the database manually?"
	errTimeoutStr = "database access timeout."
	errGetFSUsers = "Failed to get users from database: %w"
	errNoUsersStr = "No such user found."
	errActionsStr = "You have to save or cancel the current changes!"

	warnMayRefresh    = "Consider a refresh."
	errCmdNotFound    = "Not found keyboard command(s): %s ."
	errKeyNotFound    = "Not found keyboard shortcut(s): %s ."
	errGetAuthUsers   = "Failed to get users from auth: %w"
	errNoChangesStr   = "No changes"
	errUpdateFSUsers  = "Failed to update users in database: %w"
	errConfInvalidStr = "Config file is invalid"
	errCantRefreshStr = "Refresh is possible only on List page!"
)

var (
	titles = map[string]string{pageSrch: sDoSearch, pageLst: "List"}
	warns  = map[int]string{
		wSearchAgain:  "Your changes stay there from your recent searches. To remove them, click on Cancel.",
		wActionInList: "Your recent changes stay there. If you added permissions while searching, you'll only see them here after Save.",
	}
)
