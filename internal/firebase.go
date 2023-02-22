package internal

import (
	"context"
	"errors"
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

const (
	actList = iota // saveList possible actions
	actRefresh
	actSave
)

var (
	es      = struct{}{} // empty struct
	fb      FbIf
	cancelF context.CancelFunc
)

// FbIf is an interface to be able to mock Firebase functionality.
type FbIf interface {
	searchFor(key, value string, cb func(uid string))
	setClaims(uid string, newClaims map[string]any) error
	downloadClaims(uids []auth.UserIdentifier, cb func(*auth.UserRecord)) bool
	iterUsers(cb func(*auth.UserRecord))
	saveList(act int)
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
	fba, err := firebase.NewApp(getCtx(), nil, opt)
	must("initializing app", err)
	f.cAuth, err = fba.Auth(getCtx())
	must("getting Auth client", err)
	f.cFs, err = fba.Firestore(getCtx())
	must("getting Firestore client", err)
	f.fUsers = f.cFs.Collection("users")
	f.fSpecs = f.cFs.Collection("misc").Doc("specialUsers")
	return f
}

// searchFor queries users with the given start of email or name, and updates user list on screen.
func (f *Firebase) searchFor(key, value string, cb func(uid string)) {
	ds := f.fUsers.Where(key, ">=", value).Where(key, "<", value+"\uf8ff").Documents(getCtx())
	defer ds.Stop()
	uids := []auth.UserIdentifier{}
	for {
		d, err := ds.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if errors.Is(err, context.DeadlineExceeded) {
			writeErrorStr(errTimeout)
			break
		}
		if !writeErrorIf(err) {
			continue
		}
		uid := d.Ref.ID
		if _, ok := localUsers[uid]; ok {
			cb(uid)
			continue
		}
		uids = append(uids, auth.UIDIdentifier{UID: uid})
	}
	if len(uids) > 0 {
		f.downloadClaims(uids, func(r *auth.UserRecord) {
			newUserFromAuth(r, true)
			cb(r.UID)
		})
	}
}

// setClaims sets given custom claims for Firebase auth user.
func (f *Firebase) setClaims(uid string, newClaims map[string]any) error {
	return f.cAuth.SetCustomUserClaims(getCtx(), uid, newClaims)
}

// iterUsers iterates all firebase auth users, and calls callback function with privileged ones.
func (f *Firebase) iterUsers(cb func(*auth.UserRecord)) {
	var token string
	for {
		it := f.cAuth.Users(getCtx(), token)
		r, err := it.Next()
		pi := it.PageInfo()
		for !errors.Is(err, iterator.Done) && !errors.Is(err, context.DeadlineExceeded) {
			if err != nil {
				writeErrorStr(err.Error())
				continue
			}
			cb(r.UserRecord)
			if pi.Remaining() == 0 {
				break
			}
			r, err = it.Next()
		}
		if errors.Is(err, context.DeadlineExceeded) {
			writeErrorStr(errTimeout)
			return
		}
		token = pi.Token
		if len(token) == 0 {
			break
		}
	}
}

// downloadClaims updates local auth user custom claims for the given list of users.
func (f *Firebase) downloadClaims(uids []auth.UserIdentifier, cb func(*auth.UserRecord)) bool {
	idx, endIdx := 0, downLimit
	if endIdx > len(uids) {
		endIdx = len(uids)
	}
	var missing []string
	for {
		rs, err := f.cAuth.GetUsers(getCtx(), uids[idx:endIdx])
		if !writeErrorIf(err) {
			return false
		}
		for _, n := range rs.NotFound {
			uid := n.(auth.UIDIdentifier).UID
			missing = append(missing, localUsers[uid].Email)
		}
		for _, r := range rs.Users {
			cb(r)
		}
		if endIdx == len(uids) {
			break
		}
		idx, endIdx = endIdx, endIdx+downLimit
		if endIdx > len(uids) {
			endIdx = len(uids)
		}
	}

	return writeErrorList(errRemoved, missing, errManual, warnMayRefresh)
}

// saveList saves privileged user list in a transaction. Can be called with 3 possible actions:
// list downloads privileged user list for the first time, refresh downloads all users from
// Firebase auth and checks if there are new users, save saves user changes to Firebase auth.
func (f *Firebase) saveList(act int) {
	writeErrorIf(f.cFs.RunTransaction(getCtx(), func(ctx context.Context, tr *firestore.Transaction) error {
		ds, err := tr.Get(f.fSpecs)
		if !writeErrorIf(err) {
			return nil
		}
		res := ds.Data()
		updates := saveListBody(act, res)
		if len(updates) == 0 {
			return nil
		}
		err = tr.Update(f.fSpecs, updates)
		writeErrorIf(err)
		return err
	}))
}

func getCtx() context.Context {
	var ctx context.Context
	ctx, cancelF = context.WithTimeout(context.Background(), timeout)
	return ctx
}
