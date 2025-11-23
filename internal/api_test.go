package internal

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"runtime/debug"
	"strings"
	"testing"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go/auth"
	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/require"
)

const (
	someClaimKey  = "someClaim"
	someClaimVal  = "claimVal"
	errNetworkStr = "something didn't pass through"
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

type mockConfirmAction struct {
	action bool
	msg    string
}

// FrontFake implements FeIf to mock GUI operations in unit tests.
type FrontFake struct {
	currPage string
	messages []string
	stopped  bool
	t        *testing.T

	confirmActions []mockConfirmAction
}

func (f *FrontFake) run()                     {}
func (f *FrontFake) hidePopup()               {}
func (f *FrontFake) setOnShow(string, func()) {}
func (f *FrontFake) layoutUsers()             {}

func (f *FrontFake) showConfirm(m string, okFunc, cancelFunc func()) {
	if len(f.confirmActions) == 0 {
		f.t.Fatalf("an unexpected confirm dialog: %s", m)
	}

	a := f.confirmActions[0]
	copy(f.confirmActions, f.confirmActions[1:])

	require.Equal(f.t, m, a.msg)
	if a.action {
		okFunc()
	} else {
		cancelFunc()
	}
}

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
	fmt.Println("XXXnetworkOK", ok, f.netErr, string(debug.Stack()))
	if !ok {
		return errNetwork
	}
	return nil
}

func (f *FireFake) search(ctx context.Context, key, value string, cb func(uid string) error) error {
	var cmp func(*auth.UserRecord) bool
	if key == "email" {
		cmp = func(r *auth.UserRecord) bool {
			return strings.HasPrefix(r.Email, value)
		}
	} else { // name
		cmp = func(r *auth.UserRecord) bool {
			return strings.HasPrefix(r.DisplayName, value)
		}
	}
	for _, r := range f.users {
		if cmp(r) {
			newUserFromAuth(copyUserRecord(r), actSearch) // TODO: check if changed
			if err := cb(r.UID); err != nil {
				return err
			}
		}
	}
	return nil
}

func (f *FireFake) storeAuthClaims(ctx context.Context, uid string, newClaims map[string]any) error {
	fmt.Println("XXXstoreAuthClaims", uid, newClaims, f.netErr)
	if err := f.networkOK(); err != nil {
		return err
	}

	r := f.users[uid]
	if r.CustomClaims == nil {
		r.CustomClaims = map[string]any{}
	}
	maps.Copy(r.CustomClaims, newClaims)

	return nil
}

// iterUsers iterates all firebase auth users, and calls callback function with them.
func (f *FireFake) iterUsers(cb func(*auth.UserRecord) error) error {
	for _, u := range f.users {
		if err := cb(copyUserRecord(u)); err != nil {
			return err
		}
	}

	return nil
}

func (f *FireFake) getUsers(ctx context.Context, uids []auth.UserIdentifier) (*auth.GetUsersResult, error) {
	users := make([]*auth.UserRecord, len(uids))

	for i, uidi := range uids {
		users[i] = f.users[uidi.(auth.UIDIdentifier).UID]
	}

	return &auth.GetUsersResult{
		Users: users,
	}, nil
}

func (f *FireFake) getSpecs(ctx context.Context) (map[string]any, error) {
	anyMap := make(map[string]any, len(f.specUsers))

	for key, value := range f.specUsers {
		anyMap[key] = value
	}

	return anyMap, nil
}

func (f *FireFake) runTransaction(ctx context.Context, cb func(tr *firestore.Transaction, privileged map[string]any) error) error {
	privileged, _ := f.getSpecs(ctx)

	return cb(nil, privileged)
}

func (f *FireFake) updateSpecs(tr *firestore.Transaction, updates map[string]any) error {
	if err := f.networkOK(); err != nil {
		return err
	}
	for uid, change := range updates {
		if change == firestore.Delete {
			delete(f.specUsers, uid)
			fmt.Println("delete specUsers", uid)
		} else {
			f.specUsers[uid] = change.(string)
			fmt.Println("add specUsers", uid, change)
		}
	}

	return nil
}

func setup(t *testing.T) func() {
	logSync := initLogger()
	fire = &FireFake{}
	fb = fire
	front = &FrontFake{t: t}
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
		actions = map[string]map[string]any{}
		localUsers = map[string]*User{}
		warns = copyMapIntString(wrnCp)
		hidePopup()
	}
}

// checkMsg checks the result of every api function. The inner error buffer must be empty, we should
// have all error in a single "popup message".
func checkMsg(t *testing.T, descr string, wantErr ...string) {
	defer func() {
		front.messages = front.messages[:0]
	}()

	require.Emptyf(t, b.String(), "%s buffer", descr)

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

func checkList(t *testing.T, wantErr error) {
	err := showList()
	require.ErrorIs(t, err, wantErr)

	specMap := make(map[string]string, len(crntUsers))
	for _, u := range crntUsers {
		specMap[u] = localUsers[u].Email
	}

	require.Equal(t, specMap, fire.specUsers, "local/remote users")
	// checkMsg(t, "list", errs...)
}

// TestSearch tests search functionality.
func TestSearch(t *testing.T) {
	defer setup(t)()

	cases := []struct {
		name    string
		key     string
		value   string
		users   []*User
		want    []*User
		preF    func()
		wantErr error
		wantMsg string
	}{
		{
			name:    "not_found",
			key:     "email",
			value:   "notfound",
			users:   []*User{userA, userC, userB, userE, userF},
			wantErr: errNoUsers,
		},
		{
			name:  "too_short_input",
			key:   "email",
			value: "a",
			users: []*User{userA, userC, userB, userE, userF},
			preF: func() {
				// to trigger showing warning, but only the next one because
				//  too short message returns immediately
				actions[userC.UID] = map[string]any{perm0: true}
			},
			wantErr: fmt.Errorf(errMinLen, minSearchLen),
		},
		{
			name:    "found A by email",
			key:     "email",
			value:   "a@b",
			users:   []*User{userA, userC, userB, userE, userF},
			want:    []*User{userA},
			wantMsg: wrnCp[wSearchAgain],
		},
		{
			name:  "found A-s by name",
			key:   "name",
			value: "Aaa",
			users: []*User{userA, userC, userB, userE, userF},
			want:  []*User{userA, userC},
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

			err := search(c.key, c.value)
			require.Equal(t, c.wantErr, err)

			if len(c.wantMsg) > 0 {
				checkMsg(t, "wanted mssage", c.wantMsg)
			} else {
				checkMsg(t, "wanted mssage")
			}

			var current []string

			if len(c.want) > 0 {
				current = make([]string, len(c.want))
				for i, c := range c.want {
					current[i] = c.UID
				}
			}

			require.Equal(t, crntUsers, current)
		})
	}
}

// TestSaveCancel checks save and cancel functionality.
func TestSaveCancel(t *testing.T) {
	defer setup(t)()

	checkCancel := func(errWant error) {
		err := cancel()
		require.ErrorIs(t, err, errWant, "cancel")
	}

	checkSave := func(netErr int, errWant error, msgWant string) {
		fire.netErr = netErr
		err := save()
		if errWant == nil {
			require.NoError(t, err, "save")
		} else {
			require.ErrorIs(t, err, errWant, "save")
		}

		if len(msgWant) > 0 {
			checkMsg(t, "save", msgWant)
		} else {
			checkMsg(t, "save")
		}
	}

	cancelSaveNoChange := func() {
		checkCancel(errNoChanges)
		checkSave(0, errNoChanges, "")
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
					userA.UID: userToRecord(userA),
					userB.UID: userToRecord(userB),
					userC.UID: userToRecord(userC),
					userE.UID: userToRecord(userE),
					userF.UID: userToRecord(userF),
				}
				fire.users[userA.UID].CustomClaims = map[string]any{someClaimKey: someClaimVal,
					perm0: true}
				fire.users[userB.UID].CustomClaims = map[string]any{perm0: true}
				fire.specUsers = map[string]string{userA.UID: userA.Email, userB.UID: userB.Email}
				checkList(t, nil)
			},
		},
		{
			name:  "cancel_no_change",
			check: cancelSaveNoChangeBothPages,
		},
		{
			name: "cancel",
			check: func() {
				actions[userC.UID] = map[string]any{perm0: true}
				checkCancel(nil)
				require.Empty(t, actions, "actions after cancel")
			},
		},
		{
			name:  "double_cancel",
			check: cancelSaveNoChangeBothPages,
		},
		{
			name: "action_can't_refresh",
			check: func() {
				actions[userA.UID] = map[string]any{perm0: false}
				actions[userC.UID] = map[string]any{perm0: true}
				showSearch()

				wantErr := wrnCp[wSearchAgain]
				search("email", "f@b")
				checkMsg(t, "search", wantErr)

				err := refresh()
				require.Equal(t, errCantRefresh, err, "can't refresh on search")

				wantErr = wrnCp[wActionInList]
				err = showList()
				require.NoError(t, err)
				checkMsg(t, "open list with actions", wantErr)

				err = refresh()
				require.Equal(t, errActions, err, "can't refresh with active actions")
			},
		},
		{
			name: "save_partial_fail",
			check: func() {
				checkSave(2, errNetwork, "")
				require.Len(t, actions, 1)
			},
		},
		{
			name: "save_OK",
			check: func() {
				checkSave(3, nil, sSaved)
				require.Empty(t, actions, "all actions saved")
			},
		},
		{
			name: "save_inconsistent",
			check: func() {
				delete(fire.specUsers, userC.UID)
				fire.users[userB.UID].CustomClaims[perm1] = true
				actions[userB.UID] = map[string]any{perm0: true}
				checkSave(0, nil, fmt.Sprintf("The user's premissions have been changed since loading from the database. Email:\n%s", userE.Email))
				require.Empty(t, actions, "all actions saved")
			},
		},
		{
			name: "list_after_save",
			check: func() {
				fire.netErr = 100
				wantSpecUsers := map[string]struct{}{userC.UID: es, userB.UID: es, userE.UID: es}
				wantUsers := map[string]*auth.UserRecord{
					userA.UID: userToRecord(userA),
					userB.UID: userToRecord(userB),
					userC.UID: userToRecord(userC),
					userE.UID: userToRecord(userE),
					userF.UID: userToRecord(userF),
				}
				wantUsers[userA.UID].CustomClaims = map[string]any{someClaimKey: someClaimVal}
				wantUsers[userB.UID].CustomClaims = map[string]any{perm0: true}
				wantUsers[userC.UID].CustomClaims = map[string]any{perm0: true}
				wantUsers[userE.UID].CustomClaims = map[string]any{perm1: true}

				require.Len(t, fire.specUsers, len(wantSpecUsers), "user cache list on the server")
				for uid := range wantSpecUsers {
					if _, ok := fire.specUsers[uid]; !ok {
						t.Errorf("current user missing from server user cache list '%s'", uid)
					}
				}

				require.Len(t, fire.users, len(wantUsers), "users on the server")
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
				actions[userC.UID] = map[string]any{perm1: true}

				front.confirmActions = append(front.confirmActions, mockConfirmAction{
					action: true,
					msg:    fmt.Sprintf("The user's premissions have been changed since loading from the database. Email:\n%s", userE.Email),
				}, mockConfirmAction{
					action: true,
					msg:    fmt.Sprintf("The user's premissions have been changed since loading from the database. Email:\n%s", userB.Email),
				})

				checkSave(0, nil, sSaved)
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
	defer setup(t)()

	checkRefresh := func(wantErr error) {
		err := refresh()
		require.ErrorIs(t, err, wantErr)
		// checkMsg(t, "refresh", errs...)
	}

	checkRefreshNoUsers := func() {
		checkRefresh(errNoUsers)
	}

	cases := []struct {
		name  string
		check func()
	}{
		{
			name: "initial_empty_list",
			check: func() {
				checkList(t, errNoUsers)
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
					userA.UID: userToRecord(userA),
					userB.UID: userToRecord(userB),
					userC.UID: userToRecord(userC),
					userE.UID: userToRecord(userE),
					userF.UID: userToRecord(userF),
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
				front.confirmActions = append(front.confirmActions, mockConfirmAction{
					action: true,
					msg:    fmt.Sprintf("The user's premissions have been changed since loading from the database. Email:\n%s", userA.Email),
				}, mockConfirmAction{
					action: true,
					msg:    fmt.Sprintf("The user's premissions have been changed since loading from the database. Email:\n%s", userB.Email),
				})

				checkRefresh(nil)
			},
		},
		{
			name: "refresh_no_changes",
			check: func() {
				checkRefresh(errNoChanges)
			},
		},
		{
			name: "inconsistent_missing_from_cache_list",
			check: func() {
				fire.users[userA.UID].CustomClaims = map[string]any{someClaimKey: someClaimVal}
				fire.users[userE.UID].CustomClaims = map[string]any{perm0: true}
				fire.users[userB.UID].CustomClaims = map[string]any{perm1: true}

				front.confirmActions = append(front.confirmActions, mockConfirmAction{
					action: true,
					msg:    fmt.Sprintf("The user's premissions have been changed since loading from the database. Email:\n%s", userE.Email),
				})

				checkRefresh(nil)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			c.check()
		})
	}
}

// TestApi tests all api functions through keyboard shortcuts
func TestApi(t *testing.T) {
	defer setup(t)()
	loadConf(func(shortcut, menuKey, text string, isPositive bool) {}, nil)
	defer func() {
		shortcuts = map[tcell.Key]int{}
	}()

	cases := []struct {
		name    string
		pre     func()
		post    func()
		cmd     int
		key     tcell.Key
		wantMsg string
		keep    bool
		noHide  bool
	}{
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
			cmd:     cmdCancel,
			wantMsg: errNoChanges.Error(),
		},
		{
			name: "not_found_key",
			key:  tcell.KeyPause,
			keep: true,
		},
		{
			name:    "save_no_changes",
			cmd:     cmdSave,
			wantMsg: errNoChanges.Error(),
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
				actions[userA.UID] = map[string]any{perm1: true}
			},
			cmd:     cmdRefresh,
			noHide:  true,
			wantMsg: errCantRefresh.Error(),
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
			name:    "move_to_list_page",
			cmd:     cmdList,
			wantMsg: wrnCp[wActionInList],
		},
		{
			name:    "can't_refresh_because_actions",
			cmd:     cmdRefresh,
			wantMsg: errActions.Error(),
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
			if c.pre != nil {
				c.pre()
			}
			key := c.key
			if key == 0 {
				key = menuItems[c.cmd].keys[0]
			}
			event := tcell.NewEventKey(key, 0, 0)
			res := cmdByKey(event)
			checkMsg(t, "pressing menu shortcut", c.wantMsg)
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
		})
	}
}
