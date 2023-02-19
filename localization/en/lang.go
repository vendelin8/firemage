package main

const (
	menuRefresh    = "Refresh"
	menuSave       = "Save"
	menuCancel     = "Cancel"
	menuQuit       = "Quit"
	shortDesc      = "Firebase auth admin"
	longDesc       = "firemage is a CLI tool to manage Firebase Auth Claims, written in Golang"
	descKey        = "Google service account file path"
	descConf       = "config file path"
	descDebug      = "to print debug info"
	descEmul       = "if use local firebase emulator"
	sName          = "Name"
	sEmail         = "Email"
	sSearchThis    = "Search this:"
	sDoSearch      = "Search"
	sYes           = "Yes"
	sNo            = "No"
	sSaved         = "Saved"
	sNoChanges     = "No changes"
	cShortcuts     = "keyboardShortcuts"
	warnUnsaved    = "You have %d unsaved actions. Are you sure to quit?"
	warnNoUsers    = "No such user found."
	errCantRefresh = "Refresh is possible only on List page!"
	errActions     = "You have to save or cancel the current changes!"
	errMinLen      = "Insert at least %d characters!"
	warnMayRefresh = "Consider a refresh."
	errManual      = "Anyone touched the claims or the database manually?"
	errNew         = "New privileged user(s): %s ."
	errChanged     = "Changed permissions or new user(s): %s ."
	errEmpty       = "Missing user(s) from privileged ones: %s ."
	errTimeout     = "Database access timeout."
	errRemoved     = "The following user(s) were deleted from the system: %s ."
	errConfPath    = "Config file not found, please check application arguments: %s ."
	errConfParse   = "Error while parsing config file: %s ."
	errCmdNotFound = "Not found keyboard command(s): %s ."
	errKeyNotFound = "Not found keyboard shortcut(s): %s ."
)

var (
	titles = map[string]string{kSearch: sDoSearch, kList: "List"}
	warns  = map[int]string{
		wSearchAgain:  "Your changes stay there from your recent searches. To remove them, click on Cancel.",
		wActionInList: "Your recent changes stay there. If you added permissions while searching, you'll only see them here after Save.",
	}
)
