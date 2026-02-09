package firebase

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"go.uber.org/zap"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"github.com/vendelin8/firemage/internal/common"
	"github.com/vendelin8/firemage/internal/conf"
	"github.com/vendelin8/firemage/internal/frontend/window"
	"github.com/vendelin8/firemage/internal/global"
	"github.com/vendelin8/firemage/internal/lang"
	"github.com/vendelin8/firemage/internal/log"
	"github.com/vendelin8/firemage/internal/util"
)

const (
	timeout   = time.Second * 12
	downLimit = 100
)

const (
	actSearch = iota
	actList
	actSave
	actRefresh
)

var (
	ErrEnd     = errors.New("end")
	ErrTimeout = errors.New(lang.ErrTimeoutS)
	ErrMinLen  = fmt.Errorf(lang.ErrMinLen, common.MinSearchLen)
)

// Firebase implements common.FbIf for real usage.
type Firebase struct {
	cAuth  *auth.Client
	cFs    *firestore.Client
	fUsers *firestore.CollectionRef
	fSpecs *firestore.DocumentRef
}

func New() *Firebase {
	if conf.UseEmu {
		err := os.Setenv("FIRESTORE_EMULATOR_HOST", "localhost:8080")
		log.Must("setting firestore emulator env", err)
		err = os.Setenv("FIREBASE_AUTH_EMULATOR_HOST", "localhost:9099")
		log.Must("setting auth emulator env", err)
	}

	f := &Firebase{}
	opt := option.WithAuthCredentialsFile(option.ServiceAccount, conf.KeyPath)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	fba, err := firebase.NewApp(ctx, nil, opt)
	log.Must("initializing app", err)

	f.cAuth, err = fba.Auth(ctx)
	log.Must("getting Auth client", err)

	f.cFs, err = fba.Firestore(ctx)
	log.Must("getting Firestore client", err)

	f.fUsers = f.cFs.Collection("users")
	f.fSpecs = f.cFs.Collection("misc").Doc("specialUsers")

	return f
}

func (f *Firebase) Search(ctx context.Context, key, value string, cb func(uid string) error) error {
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
			return ErrTimeout
		}

		if err != nil {
			return err
		}

		if err = cb(d.Ref.ID); err != nil {
			return err
		}
	}

	return nil
}

func (f *Firebase) StoreAuthClaims(ctx context.Context, uid string, newClaims map[string]any) error {
	return f.cAuth.SetCustomUserClaims(ctx, uid, newClaims)
}

// IterUsers iterates all firebase auth users, and calls callback function with them.
func (f *Firebase) IterUsers(cb func(*auth.UserRecord) error) error {
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
			return ErrTimeout
		}
		token = pi.Token
		if len(token) == 0 {
			return ErrEnd
		}
		return nil
	}

	for {
		err := iterUser()
		if err == ErrEnd {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func (f *Firebase) GetUsers(ctx context.Context, uids []auth.UserIdentifier) (*auth.GetUsersResult, error) {
	return f.cAuth.GetUsers(ctx, uids)
}

func (f *Firebase) GetSpecs(ctx context.Context) (map[string]any, error) {
	ds, err := f.fSpecs.Get(ctx)
	if err != nil {
		return nil, err
	}

	return ds.Data(), nil
}

func (f *Firebase) RunTransaction(
	ctx context.Context,
	cb func(tr *firestore.Transaction, privileged map[string]any) error,
) error {
	ctx, cancelF := context.WithCancel(ctx)
	window.ShowProgress(ctx, cancelF)
	return f.cFs.RunTransaction(ctx, func(ctx context.Context, tr *firestore.Transaction) error {
		ds, err := tr.Get(f.fSpecs)
		if err != nil {
			return fmt.Errorf(lang.ErrGetFSUsers, err)
		}

		return cb(tr, ds.Data())
	})
}

func (f *Firebase) UpdateSpecs(tr *firestore.Transaction, updates map[string]any) error {
	s := make([]firestore.Update, 0, len(updates))

	for uid, value := range updates {
		s = append(s, firestore.Update{Path: uid, Value: value})
	}

	return tr.Update(f.fSpecs, s)
}

// Search looks for users in Firestore with email or name starting with given part.
// Results are loaded into crntUsers uid string list.
func Search(searchKey, searchValue string) error {
	log.Lgr.Debug("searching for", zap.String("key", searchKey), zap.String("value", searchValue))

	if len(searchValue) < common.MinSearchLen {
		return ErrMinLen
	}

	if len(global.Actions) > 0 {
		window.ShowWarningOnce(lang.WarnSearchAgain)
	}

	global.CrntUsers = global.CrntUsers[:0]

	err := SearchFor(searchKey, searchValue, func(newUser string) error {
		global.CrntUsers = append(global.CrntUsers, newUser)
		return nil
	})
	if err != nil {
		return fmt.Errorf(lang.ErrSearch, err)
	}

	if len(global.CrntUsers) == 0 {
		return common.ErrNoUsers
	}

	util.SortByNameThenEmail(global.CrntUsers)
	return nil
}

// SearchFor queries users with the given start of email or name, and updates user list on screen.
func SearchFor(key, value string, cb func(uid string) error) error {
	uids := []auth.UserIdentifier{}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := common.Fb.Search(ctx, key, value, func(uid string) error {
		if _, ok := global.LocalUsers[uid]; ok {
			return cb(uid)
		}
		uids = append(uids, auth.UIDIdentifier{UID: uid})
		return nil
	}); err != nil {
		return err
	}

	if len(uids) == 0 {
		return nil
	}

	return downloadClaims(uids, func(r *auth.UserRecord) error {
		if _, err := newUserFromAuth(r, actSearch, nil, nil); err != nil {
			return fmt.Errorf(lang.ErrNewUsrFrmAuth, err)
		}

		return cb(r.UID)
	})
}

// setPermissions sets given custom claims for Firebase auth user.
func setPermissions(r *auth.UserRecord, d common.ClaimsMap) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	newClaims := merge(r.CustomClaims, d)
	if err := common.Fb.StoreAuthClaims(ctx, r.UID, newClaims); err != nil {
		return fmt.Errorf("store auth claims: %w", err)
	}

	claims, err := common.NewClaimsMapFrom(r.CustomClaims)
	if err != nil {
		return fmt.Errorf("map auth claims %s => %w", d, err)
	}

	u := global.LocalUsers[r.UID]
	u.Claims = *claims

	global.LocalUsers[r.UID] = u

	return nil
}

// downloadClaims updates local auth user custom claims for the given list of users.
// Returns an optional error, and if it's critical.
func downloadClaims(uids []auth.UserIdentifier, cb func(*auth.UserRecord) error) error {
	idx, endIdx := 0, downLimit
	endIdx = min(endIdx, len(uids))
	var missing []string

	forFunc := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		rs, err := common.Fb.GetUsers(ctx, uids[idx:endIdx])
		if err != nil {
			return fmt.Errorf(lang.ErrGetAuthUsers, err)
		}

		for _, n := range rs.NotFound {
			uid := n.(auth.UIDIdentifier).UID
			missing = append(missing, global.LocalUsers[uid].Email)
		}

		for _, r := range rs.Users {
			if err = cb(r); err != nil {
				return err
			}
		}

		if endIdx == len(uids) {
			return ErrEnd
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

		if errors.Is(err, ErrEnd) {
			break
		}

		return err
	}

	if len(missing) > 0 {
		window.UseWarn()
		window.AppendListError(lang.ErrRemoved, missing, lang.ErrManualS)
		window.ShowWarn()
	}

	return nil
}

// doList downloads privileged user list for the first time.
func (f *Firebase) doList(ctx context.Context) error {
	privileged, err := f.GetSpecs(ctx)
	if err != nil {
		return fmt.Errorf(lang.ErrGetFSUsers, err)
	}

	var (
		uids  = map[string]struct{}{}
		empty []string
	)

	for uid := range privileged {
		uids[uid] = struct{}{}
		global.LocalPrivileged[uid] = struct{}{}
	}

	uidList := make([]auth.UserIdentifier, 0, len(uids))
	for uid := range uids {
		uidList = append(uidList, auth.UIDIdentifier{UID: uid})
	}

	if err = downloadClaims(uidList, func(r *auth.UserRecord) error {
		u, err := newUserFromAuth(r, actList, nil, nil)
		if err != nil {
			return fmt.Errorf(lang.ErrNewUsrFrmAuth, err)
		}

		if len(u.Claims) == 0 {
			empty = append(empty, u.Email)
			return nil
		}

		global.LocalPrivileged[u.UID] = struct{}{}

		return nil
	}); err != nil {
		return err
	}

	global.CrntUsers = make([]string, 0, len(privileged))

	for uid := range privileged {
		global.CrntUsers = append(global.CrntUsers, uid)
	}

	util.SortByNameThenEmail(global.CrntUsers)

	if window.AppendListError(lang.ErrEmpty, empty, lang.ErrManualS) {
		return window.GetError()
	}

	return nil
}

// DoList downloads privileged user list for the first time.
func (f *Firebase) DoList() error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	window.ShowProgress(ctx, cancel)

	defer cancel()

	if err := f.doList(ctx); err != nil {
		return err
	}

	if len(global.CrntUsers) == 0 {
		return fmt.Errorf("%s %s", lang.ErrNoUsersS, lang.WarnMayRefresh)
	}

	common.Fe.LayoutUsers()

	return nil
}

// DoSave saves privileged user list in a transaction Firebase auth.
func DoSave() error {
	var (
		errStoreClaims error
		clrActs        []string
	)

	updates := make(map[string]any, len(global.Actions))
	uidList := make([]auth.UserIdentifier, 0, len(global.Actions))

	for uid := range global.Actions {
		uidList = append(uidList, auth.UIDIdentifier{UID: uid})
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := common.Fb.RunTransaction(ctx, func(tr *firestore.Transaction, privileged map[string]any) error {
		errStoreClaims = downloadClaims(uidList, func(r *auth.UserRecord) error {
			_, err := newUserFromAuth(r, actSave, privileged, updates)
			if err != nil {
				return fmt.Errorf(lang.ErrNewUsrFrmAuth, err)
			}

			hasAnyValue := false
			for _, value := range util.FixedUserClaims(r.UID) {
				if !value.IsZero() {
					hasAnyValue = true
					break
				}
			}

			if hasAnyValue {
				updates[r.UID] = r.Email
			} else {
				updates[r.UID] = firestore.Delete
			}

			if err := setPermissions(r, global.Actions[r.UID]); err != nil {
				return fmt.Errorf("set permissions: %w", err)
			}

			clrActs = append(clrActs, r.UID)

			return nil
		})

		return doUpdate(updates, tr, privileged)
	}); err != nil {
		return fmt.Errorf("only store parts of your request was stored, retry saving")
	}

	if errStoreClaims == nil {
		clear(global.Actions)
		return nil
	}

	for _, a := range clrActs {
		delete(global.Actions, a)
	}

	return errStoreClaims
}

// DoRefresh downloads all users from Firebase auth and checks if there are new or changed users.
func DoRefresh() error {
	updates := make(map[string]any)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := common.Fb.RunTransaction(ctx, func(tr *firestore.Transaction, privileged map[string]any) error {
		if err := common.Fb.IterUsers(func(r *auth.UserRecord) error {
			_, err := newUserFromAuth(r, actRefresh, privileged, updates)
			if err != nil {
				return fmt.Errorf(lang.ErrNewUsrFrmAuth, err)
			}

			return nil
		}); err != nil {
			return err
		}

		return doUpdate(updates, tr, privileged)
	}); err != nil {
		return err
	}

	global.SavedUsers[lang.PageList] = global.SavedUsers[lang.PageList][:0]
	for uid := range global.LocalPrivileged {
		global.SavedUsers[lang.PageList] = append(global.SavedUsers[lang.PageList], uid)
	}

	util.SortByNameThenEmail(global.SavedUsers[lang.PageList])

	return nil
}

func doUpdate(updates map[string]any, tr *firestore.Transaction, privileged map[string]any) error {
	if len(updates) > 0 {
		if err := common.Fb.UpdateSpecs(tr, updates); err != nil {
			return fmt.Errorf(lang.ErrUpdateFSUsers, err)
		}
	}

	clear(global.LocalPrivileged)

	for uid := range privileged {
		global.LocalPrivileged[uid] = struct{}{}
	}

	return nil
}
