package internal

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

const (
	timeout   = time.Second * 12
	downLimit = 100 // firebase auth batch download limit
)

// FbIf is an interface to be able to mock Firebase functionality.
type FbIf interface {
	search(ctx context.Context, key, value string, cb func(uid string) error) error
	storeAuthClaims(ctx context.Context, uid string, newClaims map[string]any) error
	iterUsers(cb func(*auth.UserRecord) error) error
	getUsers(ctx context.Context, uids []auth.UserIdentifier) (*auth.GetUsersResult, error)
	getSpecs(ctx context.Context) (map[string]any, error)
	updateSpecs(tr *firestore.Transaction, updates map[string]any) error
	runTransaction(ctx context.Context, cb func(tr *firestore.Transaction, privileged map[string]any) error) error
}

// Firebase implements FbIf for real usage.
type Firebase struct {
	cAuth  *auth.Client
	cFs    *firestore.Client
	fUsers *firestore.CollectionRef
	fSpecs *firestore.DocumentRef // privileged user list
}

func NewFirebase() *Firebase {
	if useEmu {
		err := os.Setenv("FIRESTORE_EMULATOR_HOST", "localhost:8080")
		must("setting firestore emulator env", err)
		err = os.Setenv("FIREBASE_AUTH_EMULATOR_HOST", "localhost:9099")
		must("setting auth emulator env", err)
	}

	f := &Firebase{}
	opt := option.WithCredentialsFile(keyPath)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	fba, err := firebase.NewApp(ctx, nil, opt)
	must("initializing app", err)

	f.cAuth, err = fba.Auth(ctx)
	must("getting Auth client", err)

	f.cFs, err = fba.Firestore(ctx)
	must("getting Firestore client", err)

	f.fUsers = f.cFs.Collection("users")
	f.fSpecs = f.cFs.Collection("misc").Doc("specialUsers")

	return f
}

func (f *Firebase) search(ctx context.Context, key, value string, cb func(uid string) error) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ds := f.fUsers.Where(key, ">=", value).Where(key, "<", value+"\uf8ff").Documents(ctx)
	defer ds.Stop()

	for {
		d, err := ds.Next()
		if errors.Is(err, iterator.Done) {
			break
		}

		if errors.Is(err, context.DeadlineExceeded) {
			return errTimeout
		}

		if err != nil {
			return err // will be wrapped in a parent search func
		}

		if err = cb(d.Ref.ID); err != nil {
			return err
		}
	}

	return nil
}

func (f *Firebase) storeAuthClaims(ctx context.Context, uid string, newClaims map[string]any) error {
	return f.cAuth.SetCustomUserClaims(ctx, uid, newClaims)
}

// iterUsers iterates all firebase auth users, and calls callback function with them.
func (f *Firebase) iterUsers(cb func(*auth.UserRecord) error) error {
	var token string

	iterUser := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		it := f.cAuth.Users(ctx, token)
		r, err := it.Next()
		pi := it.PageInfo()
		for !errors.Is(err, iterator.Done) && !errors.Is(err, context.DeadlineExceeded) {
			if err != nil {
				return err
			}

			if err = cb(r.UserRecord); err != nil {
				return err
			}

			if pi.Remaining() == 0 {
				break
			}

			r, err = it.Next()
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return errTimeout
		}
		token = pi.Token
		if len(token) == 0 {
			return errEnd
		}
		return nil
	}

	for {
		err := iterUser()
		if err == errEnd {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func (f *Firebase) getUsers(ctx context.Context, uids []auth.UserIdentifier) (*auth.GetUsersResult, error) {
	return f.cAuth.GetUsers(ctx, uids)
}

func (f *Firebase) getSpecs(ctx context.Context) (map[string]any, error) {
	ds, err := f.fSpecs.Get(ctx)
	if err != nil {
		return nil, err
	}

	return ds.Data(), nil
}

func (f *Firebase) runTransaction(ctx context.Context, cb func(tr *firestore.Transaction, privileged map[string]any) error) error {
	return f.cFs.RunTransaction(ctx, func(ctx context.Context, tr *firestore.Transaction) error {
		ds, err := tr.Get(f.fSpecs)
		if err != nil {
			return fmt.Errorf(errGetFSUsers, err)
		}

		return cb(tr, ds.Data())
	})
}

func (f *Firebase) updateSpecs(tr *firestore.Transaction, updates map[string]any) error {
	s := make([]firestore.Update, 0, len(updates))

	for uid, value := range updates {
		s = append(s, firestore.Update{Path: uid, Value: value})
	}

	return tr.Update(f.fSpecs, s)
}

// searchFor queries users with the given start of email or name, and updates user list on screen.
func searchFor(key, value string, cb func(uid string) error) error {
	var err error

	uids := []auth.UserIdentifier{}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err = fb.search(ctx, key, value, func(uid string) error {
		if _, ok := localUsers[uid]; ok {
			if err = cb(uid); err != nil {
				return err
			}

			return nil
		}
		uids = append(uids, auth.UIDIdentifier{UID: uid})

		return nil
	}); err != nil {
		return err
	}

	if len(uids) > 0 {
		downloadClaims(uids, func(r *auth.UserRecord) error {
			newUserFromAuth(r, actSearch) // TODO: check if changed
			return cb(r.UID)
		})
	}

	return nil
}

// setPermissions sets given custom claims for Firebase auth user.
func setPermissions(r *auth.UserRecord, d map[string]any) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	newClaims := merge(r.CustomClaims, d)
	if err := fb.storeAuthClaims(ctx, r.UID, newClaims); err != nil {
		return fmt.Errorf("store auth claims: %w", err) // TODO: translate
	}

	localUsers[r.UID].Claims = r.CustomClaims

	return nil
}

// downloadClaims updates local auth user custom claims for the given list of users.
func downloadClaims(uids []auth.UserIdentifier, cb func(*auth.UserRecord) error) error {
	idx, endIdx := 0, downLimit
	if endIdx > len(uids) {
		endIdx = len(uids)
	}
	var missing []string

	forFunc := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		rs, err := fb.getUsers(ctx, uids[idx:endIdx])
		if err != nil {
			return fmt.Errorf(errGetAuthUsers, err)
		}

		for _, n := range rs.NotFound {
			uid := n.(auth.UIDIdentifier).UID
			missing = append(missing, localUsers[uid].Email)
		}

		for _, r := range rs.Users {
			if err = cb(r); err != nil {
				return err
			}
		}

		if endIdx == len(uids) {
			return errEnd
		}

		idx, endIdx = endIdx, endIdx+downLimit
		if endIdx > len(uids) {
			endIdx = len(uids)
		}
		return nil
	}

	for {
		err := forFunc()
		if err == nil {
			continue
		}

		if errors.Is(err, errEnd) {
			break
		}

		return err
	}

	// this error shouldn't break any ongoing process, but need to be shown to the user
	writeErrorListShow(errRemoved, missing, errManualStr, warnMayRefresh)

	return nil
}

// list downloads privileged user list for the first time.
func doList() error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	privileged, err := fb.getSpecs(ctx)
	if err != nil {
		return fmt.Errorf(errGetFSUsers, err)
	}

	var (
		uids  = map[string]struct{}{}
		empty []string
	)

	for uid := range privileged {
		uids[uid] = es
		localPrivileged[uid] = es
	}

	uidList := make([]auth.UserIdentifier, 0, len(uids))
	for uid := range uids {
		uidList = append(uidList, auth.UIDIdentifier{UID: uid})
	}

	if err := downloadClaims(uidList, func(r *auth.UserRecord) error {
		u, claims := newUserFromAuth(r, actList)
		if len(claims) == 0 {
			empty = append(empty, u.Email)
			return nil
		}

		localPrivileged[u.UID] = es

		return nil
	}); err != nil {
		return err
	}
	crntUsers = make([]string, 0, len(privileged))
	for uid := range privileged {
		crntUsers = append(crntUsers, uid)
	}
	sortByNameThenEmail(crntUsers)

	return writeErrorListGet(errEmpty, empty, errManualStr, warnMayRefresh)
}

// doSave saves privileged user list in a transaction Firebase auth.
func doSave() error {
	var (
		errStoreClaims error
		clrActs        []string
	)

	updates := make(map[string]any, len(actions)) // saving them to firestore
	uidList := make([]auth.UserIdentifier, 0, len(actions))

	for uid := range actions { // iterating users with actions is the main purpose here
		uidList = append(uidList, auth.UIDIdentifier{UID: uid})
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := fb.runTransaction(ctx, func(tr *firestore.Transaction, privileged map[string]any) error {
		fmt.Println("XXXdoSave", uidList)
		errStoreClaims = downloadClaims(uidList, func(r *auth.UserRecord) error {
			err := checkChanged(r, actSave, privileged, updates)
			if err != nil {
				fmt.Println("XXXdoSave checkChanged err != nil", err)
				return err
			}

			// if len(newClaims) == 0 {
			// 	fmt.Println("XXXdoSave checkChanged len(newClaims)")
			// 	return nil
			// }

			for _, value := range actions[r.UID] {
				if value == false {
					updates[r.UID] = firestore.Delete
				} else {
					updates[r.UID] = r.Email
				}
			}

			if err := fb.storeAuthClaims(ctx, r.UID, actions[r.UID]); err != nil {
				fmt.Println("fb.storeAuthClaims ERROR", err)
				return fmt.Errorf("store auth claims: %w", err) // TODO: translate
			}

			clrActs = append(clrActs, r.UID)
			fmt.Println("clrActs = append(clrActs, r.UID)", clrActs)

			return nil
		})

		fmt.Println("XXXdoSave errStoreClaims:", errStoreClaims)

		return doUpdate(updates, tr, privileged)
	}); err != nil {
		return fmt.Errorf("only store parts of your request was stored, retry saving") // TODO: translate
	}

	if errStoreClaims == nil {
		clear(actions)
		return nil
	}

	fmt.Println("XXXE?!", clrActs)
	for _, a := range clrActs {
		delete(actions, a)
	}
	fmt.Println("XXXE?!", actions)

	return errStoreClaims
}

// doRefresh downloads all users from Firebase auth and checks if there are new or changed users.
func doRefresh() error {
	var err error

	updates := make(map[string]any, len(actions)) // saving them to firestore
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err = fb.runTransaction(ctx, func(tr *firestore.Transaction, privileged map[string]any) error {
		if err = fb.iterUsers(func(r *auth.UserRecord) error {
			return checkChanged(r, actRefresh, privileged, updates) // don't want to save all users locally
		}); err != nil {
			return err
		}

		return doUpdate(updates, tr, privileged)
	}); err != nil {
		return err
	}

	savedUsers[pageLst] = savedUsers[pageLst][:0]
	for uid := range localPrivileged {
		savedUsers[pageLst] = append(savedUsers[pageLst], uid)
	}

	sortByNameThenEmail(savedUsers[pageLst]) // TODO: need to redraw?

	return nil
}

// checkChanged checks if the user has been changed since the last load
//   - if cloud auth isPrivileged != cloud hasClaims
//     show manualChanges error in the current confirm
//   - if
//     the user is privileged here, but has no claims in the cloud OR
//     the claims don't match
//     ask if we store the claims back
//   - if the user is not privileged here but has claims in the cloud
//     ask if we delete cloud claims
func checkChanged(r *auth.UserRecord, act int, privileged map[string]any, updates map[string]any) error {
	uid := r.UID
	u, filteredClaims := newUserFromAuth(r, act)

	if filteredClaims == nil || !differs(u.Claims, filteredClaims) {
		fmt.Printf("checkChanged no diff %#v; %#v, %t\n", u, filteredClaims, differs(u.Claims, filteredClaims))
		return nil
	}

	hasClaims := len(u.Claims) > 0
	_, isPrivileged := privileged[uid]

	compareClaims := u.Claims
	if act == actSave {
		_, _, compareClaims = fixedUserClaims(r.UID)
		if !differs(compareClaims, filteredClaims) {
			fmt.Printf("checkChanged no diff2 %#v\n", u)
			return nil // when saving, we ignore if db already contains our changes
		}
	}

	plus, minus := diffHuman(u.Claims, filteredClaims)

	fmt.Printf("manual error %#v: +%#v; -%#v; hasClaims(%t) ?= (%t)isPrivileged\n", u, plus, minus, hasClaims, isPrivileged)
	writeErrorStr("The user's premissions have been changed since loading from the database. Email:") // TODO: Translate
	writeErrorStr(u.Email)
	if len(plus) > 0 {
		writeErrorStr("Added permissions:") // TODO: Translate
		writeErrorStr(fmt.Sprintf("%v", plus))
	}
	if len(minus) > 0 {
		writeErrorStr("Removed permissions:") // TODO: Translate
		writeErrorStr(fmt.Sprintf("%v", minus))
	}

	if hasClaims != isPrivileged {
		writeErrorStr(errManualStr)
	}

	writeErrorStr("Click OK to save yours to the database or Cancel to download them.") // TODO: Translate

	var err error

	fe.showConfirm(getErrorStr(), func() {
		if hasClaims {
			privileged[uid] = es

			if !isPrivileged {
				updates[uid] = u.Email
			}
		} else {
			updates[uid] = firestore.Delete
			delete(privileged, uid)
		}

		if err = setPermissions(r, compareClaims); err != nil {
			err = fmt.Errorf("set permissions: %w", err) // TODO: translate
			return
		}

		u.Claims = compareClaims
	}, func() {
		localUsers[uid] = u
		compareClaims = nil
	})

	return err
}

func doUpdate(updates map[string]any, tr *firestore.Transaction, privileged map[string]any) error {
	fmt.Println("doUpdate", updates)
	if len(updates) > 0 {
		if err := fb.updateSpecs(tr, updates); err != nil {
			return fmt.Errorf(errUpdateFSUsers, err)
		}
	}

	clear(localPrivileged)

	for uid := range privileged {
		localPrivileged[uid] = es
	}

	return nil
}
