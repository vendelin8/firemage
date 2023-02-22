package internal

import (
	"errors"
	"fmt"
	"runtime/debug"
	"strings"
	"testing"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go/auth"
	"github.com/gdamore/tcell"
)

const (
	someClaimKey  = "someClaim"
	someClaimVal  = "claimVal"
	errNetworkStr = "Something didn't pass through"
)

var (
	fire       *FireFake
	front      *FrontFake
	perm0      = "admin"
	perm1      = "consultant"
	wrnCp      map[int]string
	errNetwork = errors.New(errNetworkStr)
)

func init() {
	wrnCp = copyMapIntString(warns)
	// verbose = true
}

func copyMapIntString(src map[int]string) map[int]string {
	dst := map[int]string{}
	for key, value := range src {
		dst[key] = value
	}
	return dst
}

// FrontFake implements FeIf to mock GUI operations in unit tests.
type FrontFake struct {
	currPage string
	messages []string
	stopped  bool
}

func (f *FrontFake) run()                                            {}
func (f *FrontFake) showConfirm(m string, okFunc, cancelFunc func()) {}
func (f *FrontFake) hidePopup()                                      {}
func (f *FrontFake) setOnShow(string, func())                        {}
func (f *FrontFake) layoutUsers()                                    {}

func (f *FrontFake) setPage(newPage string) {
	f.currPage = newPage
}

func (f *FrontFake) currentPage() string {
	return f.currPage
}

func (f *FrontFake) quit() {
	f.stopped = true
}

func (f *FrontFake) showMsg(m string) {
	f.messages = append(f.messages, m)
}

func (u *User) String() string {
	return fmt.Sprintf("{%s %s %s %#v}", u.Name, u.Email, u.UID, u.Claims)
}

func userToRecord(u *User) *auth.UserRecord {
	return &auth.UserRecord{UserInfo: &auth.UserInfo{DisplayName: u.Name, UID: u.UID, Email: u.Email}}
}

func copyUserRecord(r *auth.UserRecord) *auth.UserRecord {
	var newClaims map[string]any
	if len(r.CustomClaims) > 0 {
		newClaims = map[string]any{}
		for k, v := range r.CustomClaims {
			newClaims[k] = v
		}
	}
	return &auth.UserRecord{UserInfo: r.UserInfo, CustomClaims: newClaims}
}

// FireFake implements FbIf to mock Firebase operations in unit tests.
type FireFake struct {
	users     map[string]*auth.UserRecord
	specUsers map[string]string
	netErr    int // simulating network error when == 1; > 1 decreases in every step
}

// networkOK may add a simulated network error when the counter is 1.
func (f *FireFake) networkOK() error {
	ok := f.netErr != 1
	f.netErr--
	if !ok {
		return errNetwork
	}
	return nil
}

func (f *FireFake) searchFor(key, value string, cb func(newUser string)) {
	var cmp func(*auth.UserRecord) bool
	if key == email {
		cmp = func(r *auth.UserRecord) bool {
			return strings.HasPrefix(r.Email, value)
		}
	} else {
		cmp = func(r *auth.UserRecord) bool {
			return strings.HasPrefix(r.DisplayName, value)
		}
	}
	for _, r := range f.users {
		if cmp(r) {
			newUserFromAuth(copyUserRecord(r), true)
			cb(r.UID)
		}
	}
}

func (f *FireFake) setClaims(uid string, newClaims map[string]any) error {
	if err := f.networkOK(); err != nil {
		return err
	}
	f.users[uid].CustomClaims = newClaims
	return nil
}

func (f *FireFake) downloadClaims(uids []auth.UserIdentifier, cb func(*auth.UserRecord)) bool {
	if !writeErrorIf(f.networkOK()) {
		return false
	}
	missing := []string{}
	for _, u := range uids {
		uid := u.(auth.UIDIdentifier).UID
		if u, ok := f.users[uid]; ok {
			cb(copyUserRecord(u))
			continue
		}
		missing = append(missing, localUsers[uid].Email)
	}
	return writeErrorList(errRemoved, missing, errManual, warnMayRefresh)
}

func (f *FireFake) iterUsers(cb func(*auth.UserRecord)) {
	for _, u := range f.users {
		cb(copyUserRecord(u))
	}
}

// saveList saves privileged user cache list in a simulated transaction.
func (f *FireFake) saveList(act int) {
	res := map[string]any{}
	for uid, email := range f.specUsers {
		res[uid] = email
	}
	updates := saveListBody(act, res)
	for _, update := range updates {
		if update.Value == firestore.Delete {
			delete(f.specUsers, update.Path)
		} else {
			f.specUsers[update.Path] = update.Value.(string)
		}
	}
}

func setup() func() {
	logSync := initLogger()
	fire = &FireFake{}
	fb = fire
	front = &FrontFake{}
	fe = front
	permsMap = map[string]string{
		perm0: "Admin",
		perm1: "Consultant",
	}
	allPerms = []string{perm0, perm1}

	return func() {
		logSync()
		savedUsers = map[string][]string{}
		crntUsers = crntUsers[:0]
		actions = map[string]map[string]bool{}
		localUsers = map[string]*User{}
		warns = copyMapIntString(wrnCp)
		hidePopup()
	}
}

func checkLen(t *testing.T, descr string, wantL, resultL int, want, result any) {
	if wantL != resultL {
		t.Fatalf("%s: expected length %d != %d result length; %s VS %s\n%s", descr, wantL, resultL,
			want, result, debug.Stack())
	}
}

func checkItemAt(t *testing.T, descr, want, result string, index int) {
	if result != want {
		t.Errorf("%s at index %d expected %s != %s result", descr, index, want, result)
	}
}

func checkEmpty(t *testing.T, descr string, l int, target any) {
	if l > 0 {
		t.Errorf("%s should be empty but it's %#v\n%s", descr, target, debug.Stack())
	}
}

// checkMsg checks the result of every api function. The inner error buffer must be empty, we should
// have all error in a single "popup message".
func checkMsg(t *testing.T, descr string, wantErr ...string) {
	defer func() {
		front.messages = front.messages[:0]
	}()
	checkEmpty(t, descr+" buffer", b.Len(), b.String())
	if len(wantErr) == 0 {
		if len(front.messages) == 0 {
			return
		}
		t.Errorf("%s: messages should be empty but it's %#v", descr, front.messages)
	}
	if len(front.messages) == 0 && len(wantErr) == 0 {
		return
	}
	want := strings.Join(wantErr, "\n")
	if len(front.messages) == 0 {
		t.Errorf("%s: result should be '%s' but it's empty\n%s", descr, want, debug.Stack())
		return
	}
	if len(front.messages) > 1 {
		t.Fatalf("%s: only one message should be, but it's %d: %#v\n%s", descr,
			len(front.messages), front.messages, debug.Stack())
	}
	result := front.messages[0]
	if result == want {
		return
	}
	t.Errorf("%s: result should be '%s' but it's '%s'\n%s", descr, want, result, debug.Stack())
}

func checkList(t *testing.T, errs ...string) {
	showList()
	checkLen(t, "local/remote users", len(crntUsers), len(fire.specUsers), crntUsers, fire.specUsers)
	for _, uid := range crntUsers {
		if _, ok := fire.specUsers[uid]; !ok {
			t.Errorf("current user missing from server user cache list '%s'", uid)
		}
	}
	checkMsg(t, "list", errs...)
}

// TestSearch searches test functionality.
func TestSearch(t *testing.T) {
	defer setup()()

	cases := []struct {
		name  string
		key   string
		value string
		users []*User
		want  []*User
		preF  func()
		errs  []string
	}{
		{
			name:  "not found",
			key:   email,
			value: "notfound",
			users: []*User{userA, userA2, userB, userE, userE2},
			errs:  []string{warnNoUsers},
		},
		{
			name:  "too short input",
			key:   email,
			value: "a",
			users: []*User{userA, userA2, userB, userE, userE2},
			preF: func() {
				// to trigger showing warning, but only the next one because
				//  too short message returns immediately
				actions[userA2.UID] = map[string]bool{perm0: true}
			},
			errs: []string{fmt.Sprintf(errMinLen, minSearchLen)},
		},
		{
			name:  "found A by email",
			key:   email,
			value: "a@b",
			users: []*User{userA, userA2, userB, userE, userE2},
			want:  []*User{userA},
			errs:  []string{wrnCp[wSearchAgain]},
		},
		{
			name:  "found A-s by name",
			key:   name,
			value: "Aaa",
			users: []*User{userA, userA2, userB, userE, userE2},
			want:  []*User{userA, userA2},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.preF != nil {
				c.preF()
			}
			fire.users = map[string]*auth.UserRecord{}
			for _, u := range c.users {
				fire.users[u.UID] = userToRecord(u)
			}
			search(c.key, c.value)
			checkLen(t, "current users", len(c.want), len(crntUsers), c.want, crntUsers)
			for i, u := range c.want {
				checkItemAt(t, "current user", crntUsers[i], u.UID, i)
			}
			checkMsg(t, "search", c.errs...)
		})
	}
}

// TestSaveCancel checks save and cancel functionality.
func TestSaveCancel(t *testing.T) {
	defer setup()()

	checkCancel := func(errs ...string) {
		cancel()
		checkMsg(t, "cancel", errs...)
	}

	checkSave := func(netErr int, errs ...string) {
		fire.netErr = netErr
		save()
		checkMsg(t, "save", errs...)
	}

	cancelSaveNoChange := func() {
		checkCancel(sNoChanges)
		checkSave(0, sNoChanges)
	}

	cancelSaveNoChangeBothPages := func() {
		showList()
		cancelSaveNoChange()
		showSearch()
		cancelSaveNoChange()
	}

	cases := []struct {
		name  string
		check func()
	}{
		{
			name: "initial",
			check: func() {
				fire.users = map[string]*auth.UserRecord{
					userA.UID:  userToRecord(userA),
					userB.UID:  userToRecord(userB),
					userA2.UID: userToRecord(userA2),
					userE.UID:  userToRecord(userE),
					userE2.UID: userToRecord(userE2),
				}
				fire.users[userA.UID].CustomClaims = map[string]any{someClaimKey: someClaimVal,
					perm0: true}
				fire.users[userB.UID].CustomClaims = map[string]any{perm0: true}
				fire.specUsers = map[string]string{userA.UID: userA.Email, userB.UID: userB.Email}
				checkList(t)
			},
		},
		{
			name:  "cancel_no_change",
			check: cancelSaveNoChangeBothPages,
		},
		{
			name: "cancel",
			check: func() {
				actions[userA2.UID] = map[string]bool{perm0: true}
				checkCancel()
				checkEmpty(t, "actions after cancel", len(actions), actions)
			},
		},
		{
			name:  "double_cancel",
			check: cancelSaveNoChangeBothPages,
		},
		{
			name: "action_can't_refresh",
			check: func() {
				actions[userA2.UID] = map[string]bool{perm0: true}
				actions[userA.UID] = map[string]bool{perm0: false}
				showSearch()
				wantErr := wrnCp[wSearchAgain]
				search(email, "f@b")
				checkMsg(t, "search", wantErr)
				refresh()
				checkMsg(t, "can't refresh on search", errCantRefresh)
				wantErr = wrnCp[wActionInList]
				showList()
				checkMsg(t, "open list with actions", wantErr)
				refresh()
				checkMsg(t, "can't refresh with active actions", errActions)
			},
		},
		{
			name: "save_partial_fail",
			check: func() {
				checkSave(2, errNetworkStr)
				if len(actions) != 1 {
					t.Errorf("action should be 1 but it's %d: %#v", len(actions), actions)
				}
			},
		},
		{
			name: "save_OK",
			check: func() {
				checkSave(0, sSaved)
				checkEmpty(t, "all actions saved", len(actions), actions)
			},
		},
		{
			name: "save_but_missing",
			check: func() {
				delete(fire.specUsers, userA2.UID)
				actions[userA2.UID] = map[string]bool{perm0: true}
				actions[userE.UID] = map[string]bool{perm1: true}
				checkSave(0, errManual, warnMayRefresh)
				checkEmpty(t, "all actions saved", len(actions), actions)
			},
		},
		{
			name: "list_after_save",
			check: func() {
				wantSpecUsers := map[string]struct{}{userA2.UID: es, userB.UID: es, userE.UID: es}
				wantUsers := map[string]*auth.UserRecord{
					userA.UID:  userToRecord(userA),
					userB.UID:  userToRecord(userB),
					userA2.UID: userToRecord(userA2),
					userE.UID:  userToRecord(userE),
					userE2.UID: userToRecord(userE2),
				}
				wantUsers[userA.UID].CustomClaims = map[string]any{someClaimKey: someClaimVal}
				wantUsers[userB.UID].CustomClaims = map[string]any{perm0: true}
				wantUsers[userA2.UID].CustomClaims = map[string]any{perm0: true}
				wantUsers[userE.UID].CustomClaims = map[string]any{perm1: true}
				checkLen(t, "user cache list on the server", len(wantSpecUsers),
					len(fire.specUsers), wantSpecUsers, fire.specUsers)
				for uid := range wantSpecUsers {
					if _, ok := fire.specUsers[uid]; !ok {
						t.Errorf("current user missing from server user cache list '%s'", uid)
					}
				}
				checkLen(t, "users on the server", len(wantUsers), len(fire.users),
					wantUsers, fire.users)
				for uid := range wantUsers {
					if _, ok := fire.users[uid]; !ok {
						t.Errorf("current user missing from server '%s'", uid)
					}
				}
			},
		},
		{
			name:  "final_cancel_no_change",
			check: cancelSaveNoChangeBothPages,
		},
		{
			name: "inconsistent_missing auth",
			check: func() {
				delete(fire.users, userE.UID)
				fire.users[userB.UID].CustomClaims = nil
				actions[userA2.UID] = map[string]bool{perm1: true}
				checkSave(0, fmt.Sprintf(errRemoved, userE.Email), errManual, warnMayRefresh,
					fmt.Sprintf(errEmpty, userB.Email))
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			c.check()
		})
	}
}

// TestRefresh checks refresh functionality.
func TestRefresh(t *testing.T) {
	defer setup()()

	checkRefresh := func(errs ...string) {
		refresh()
		checkMsg(t, "refresh", errs...)
	}

	checkRefreshNoUsers := func() {
		checkRefresh(warnNoUsers)
	}

	cases := []struct {
		name  string
		check func()
	}{
		{
			name: "initial_empty_list",
			check: func() {
				checkList(t, warnNoUsers, warnMayRefresh)
			},
		},
		{
			name:  "initial_refresh",
			check: checkRefreshNoUsers,
		},
		{
			name: "initial_data",
			check: func() {
				fire.users = map[string]*auth.UserRecord{
					userA.UID:  userToRecord(userA),
					userB.UID:  userToRecord(userB),
					userA2.UID: userToRecord(userA2),
					userE.UID:  userToRecord(userE),
					userE2.UID: userToRecord(userE2),
				}
				fire.users[userA.UID].CustomClaims = map[string]any{someClaimKey: someClaimVal,
					perm0: true}
				fire.users[userB.UID].CustomClaims = map[string]any{perm0: true}
				fire.specUsers = map[string]string{userA.UID: userA.Email, userB.UID: userB.Email}
			},
		},
		{
			name: "refresh_with_data",
			check: func() {
				checkRefresh(fmt.Sprintf(errNew,
					strings.Join([]string{userA.Email, userB.Email}, ", ")))
			},
		},
		{
			name: "refresh_no_changes",
			check: func() {
				checkRefresh(sNoChanges)
			},
		},
		{
			name: "inconsistent_missing_from_cache_list",
			check: func() {
				fire.users[userA.UID].CustomClaims = map[string]any{someClaimKey: someClaimVal}
				fire.users[userE.UID].CustomClaims = map[string]any{perm0: true}
				fire.users[userB.UID].CustomClaims = map[string]any{perm1: true}
				checkRefresh(fmt.Sprintf(errChanged, userB.Email), fmt.Sprintf(errNew, userE.Email),
					errManual)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			c.check()
		})
	}
}

type apiTestCase struct {
	name   string
	pre    func()
	post   func()
	cmd    int
	key    tcell.Key
	wErr   []string
	keep   bool
	noHide bool
}

func checkAPI(t *testing.T, c apiTestCase) {
	if c.pre != nil {
		c.pre()
	}
	key := c.key
	if key == 0 {
		key = menuItems[c.cmd].keys[0]
	}
	event := tcell.NewEventKey(key, 0, 0)
	res := cmdByKey(event)
	checkMsg(t, "pressing menu shortcut", c.wErr...)
	if (res == nil) == c.keep {
		if c.keep {
			t.Fatal("key should have NOT been consumed but it was")
		}
		t.Fatal("key should have been consumed but it wasn't")
	}
	if c.post != nil {
		c.post()
	}
	if !c.noHide {
		hidePopup()
	}
}

// TestApi tests all api functions through keyboard shortcuts
func TestApi(t *testing.T) {
	defer setup()()
	loadConf(func(shortcut, menuKey, text string, isPositive bool) {}, nil)
	defer func() {
		shortcuts = map[tcell.Key]int{}
	}()

	cases := []apiTestCase{
		{
			name: "initial_list",
			pre: func() {
				fire.users = map[string]*auth.UserRecord{
					userA.UID: userToRecord(userA),
					userB.UID: userToRecord(userB),
				}
				fire.users[userA.UID].CustomClaims = map[string]any{perm0: true}
				fire.specUsers = map[string]string{userA.UID: userA.Email}
			},
			cmd:  cmdCancel,
			wErr: []string{sNoChanges},
		},
		{
			name: "not_found_key",
			key:  tcell.KeyPause,
			keep: true,
		},
		{
			name: "save_no_changes",
			cmd:  cmdSave,
			wErr: []string{sNoChanges},
		},
		{
			name: "quit_OK",
			pre: func() {
				if front.stopped {
					t.Fatal("should NOT have been stopped but was")
				}
			},
			cmd: cmdQuit,
			post: func() {
				if !front.stopped {
					t.Fatal("should have been stopped but wasn't")
				}
				front.stopped = false
			},
		},
		{
			name: "can't_refresh_on_search",
			pre: func() {
				actions[userA.UID] = map[string]bool{perm1: true}
			},
			cmd:    cmdRefresh,
			noHide: true,
			wErr:   []string{errCantRefresh},
		},
		{
			name: "can't_move_to_list_page_because of popup",
			cmd:  cmdList,
			keep: true,
			post: func() {
				if !hasPopup() {
					t.Fatal("should have popup but doesn't")
				}
			},
		},
		{
			name: "move_to_list_page",
			cmd:  cmdList,
			wErr: []string{wrnCp[wActionInList]},
		},
		{
			name: "can't_refresh_because_actions",
			cmd:  cmdRefresh,
			wErr: []string{errActions},
		},
		{
			name: "move_to_list_search",
			cmd:  cmdSearch,
		},
		{
			name: "quit_can't_because_actions",
			pre: func() {
				if front.stopped {
					t.Fatal("should NOT have been stopped but was")
				}
			},
			cmd:    cmdQuit,
			noHide: true,
		},
		{
			name: "hide_popup_with_esc",
			pre: func() {
				if front.stopped {
					t.Fatal("should NOT have been stopped but was")
				}
				if !hasPopup() {
					t.Fatal("should have popup but doesn't")
				}
			},
			cmd: cmdQuit,
			post: func() {
				if front.stopped {
					t.Fatal("should NOT have been stopped but was")
				}
				if hasPopup() {
					t.Fatal("should NOT have popup but does")
				}
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			checkAPI(t, c)
		})
	}
}
