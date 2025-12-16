package firebase

import (
	"context"
	"testing"

	"firebase.google.com/go/auth"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/vendelin8/firemage/internal/common"
	"github.com/vendelin8/firemage/internal/global"
	"github.com/vendelin8/firemage/internal/lang"
	"github.com/vendelin8/firemage/internal/mock"
	testutil "github.com/vendelin8/firemage/internal/util/test"
)

func TestSearchFor(t *testing.T) {
	cleanup := testutil.InitLog()
	defer cleanup()

	tests := []struct {
		name        string
		searchKey   string
		searchValue string
		setup       func(results []string, mockFb *mock.MockFbIf, mockFe *mock.MockFeIf)
		actions     map[string]map[string]any
		results     []string
		cbErr       error
		wantError   error
	}{
		{
			name:        "with valid email",
			searchKey:   "email",
			searchValue: "test@example.com",
		},
		{
			name:        "with mock error",
			searchKey:   "email",
			searchValue: "test@",
			actions:     map[string]map[string]any{"uid1": {"admin": true}},
			results:     []string{"uid1"},
			cbErr:       testutil.ErrMock,
			wantError:   testutil.ErrMock,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockFb := mock.NewMockFbIf(ctrl)
			mockFe := mock.NewMockFeIf(ctrl)
			common.Fb = mockFb
			common.Fe = mockFe

			defer func() {
				ctrl.Finish()
				global.Actions = map[string]common.ClaimsMap{}
				global.CrntUsers = []string{}
				global.LocalUsers = make(map[string]*global.User)
				lang.Warns = make(map[int]string)
			}()

			// Setup global state
			global.Actions = testutil.BuildActionsMap(tt.actions)
			global.CrntUsers = []string{}
			global.LocalUsers = make(map[string]*global.User)

			mockFb.EXPECT().Search(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, _, _ string, cb func(uid string) error) error {
					for _, uid := range tt.results {
						// Only populate if not already pre-populated
						if _, exists := global.LocalUsers[uid]; !exists {
							global.LocalUsers[uid] = &global.User{
								UID:   uid,
								Email: uid + "@example.com",
								Name:  "User " + uid,
							}
						}
						if err := cb(uid); err != nil {
							return err
						}
					}
					return nil
				}).
				Times(1)

			// Execute test
			err := SearchFor(tt.searchKey, tt.searchValue, func(uid string) error {
				return tt.cbErr
			})

			// Verify results
			assert.ErrorIs(t, err, tt.wantError)
		})
	}
}

func TestSearch(t *testing.T) {
	cleanup := testutil.InitLog()
	defer cleanup()

	type testCase struct {
		name        string
		searchKey   string
		searchValue string
		setup       func(results []string, mockFb *mock.MockFbIf, mockFe *mock.MockFeIf)
		actions     map[string]map[string]any
		results     []string
		wantOrder   []string
		wantError   error
	}

	setupMock := func(results []string, mockFb *mock.MockFbIf, mockFe *mock.MockFeIf) {
		mockFb.EXPECT().Search(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, _, _ string, cb func(uid string) error) error {
				for _, uid := range results {
					// Only populate if not already pre-populated
					if _, exists := global.LocalUsers[uid]; !exists {
						global.LocalUsers[uid] = &global.User{
							UID:   uid,
							Email: uid + "@example.com",
							Name:  "User " + uid,
						}
					}
					if err := cb(uid); err != nil {
						return err
					}
				}
				return nil
			}).
			Times(1)
	}

	tests := []testCase{
		{
			name:        "with valid results",
			searchKey:   "email",
			searchValue: "test@example.com",
			setup:       setupMock,
			results:     []string{"uid1", "uid2"},
		},
		{
			name:        "with value too short",
			searchKey:   "email",
			searchValue: "t",
			wantError:   ErrMinLen,
		},
		{
			name:        "returns no results",
			searchKey:   "email",
			searchValue: "notfound@example.com",
			setup:       setupMock,
			wantError:   common.ErrNoUsers,
		},
		{
			name:        "shows warning when actions pending",
			searchKey:   "email",
			searchValue: "test@",
			setup: func(results []string, mockFb *mock.MockFbIf, mockFe *mock.MockFeIf) {
				setupMock(results, mockFb, mockFe)
				lang.Warns = map[int]string{
					lang.WarnSearchAgain: "There are unsaved changes",
				}
				mockFe.EXPECT().ShowMsg(gomock.Any()).Times(1)

			},
			actions: map[string]map[string]any{"uid1": {"admin": true}},
			results: []string{"uid1"},
		},
		{
			name:        "results are sorted by name then email",
			searchKey:   "email",
			searchValue: "example",
			results:     []string{"uid1", "uid2", "uid3"},
			setup: func(results []string, mockFb *mock.MockFbIf, mockFe *mock.MockFeIf) {
				setupMock(results, mockFb, mockFe)
				global.LocalUsers = map[string]*global.User{
					"uid1": {UID: "uid1", Email: "alice@example.com", Name: "Charlie"},
					"uid2": {UID: "uid2", Email: "bob@example.com", Name: "Alice"},
					"uid3": {UID: "uid3", Email: "charlie@example.com", Name: "Alice"},
				}
			},
			wantOrder: []string{"uid2", "uid3", "uid1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockFb := mock.NewMockFbIf(ctrl)
			mockFe := mock.NewMockFeIf(ctrl)
			common.Fb = mockFb
			common.Fe = mockFe

			defer func() {
				ctrl.Finish()
				global.Actions = map[string]common.ClaimsMap{}
				global.CrntUsers = []string{}
				global.LocalUsers = make(map[string]*global.User)
				lang.Warns = make(map[int]string)
			}()

			// Setup global state
			global.Actions = testutil.BuildActionsMap(tt.actions)
			global.CrntUsers = []string{}
			global.LocalUsers = make(map[string]*global.User)

			if tt.setup != nil {
				tt.setup(tt.results, mockFb, mockFe)
			}

			// Execute test
			err := Search(tt.searchKey, tt.searchValue)

			// Verify results
			assert.ErrorIs(t, err, tt.wantError)
			if tt.wantError == nil && len(tt.results) > 0 {
				if len(tt.wantOrder) > 0 {
					assert.Equal(t, tt.wantOrder, global.CrntUsers)
				} else {
					assert.Equal(t, tt.results, global.CrntUsers)
				}
			}
		})
	}
}

func TestDownloadClaims(t *testing.T) {
	cleanup := testutil.InitLog()
	defer cleanup()

	tests := []struct {
		name      string
		userCount int
		wantError error
	}{
		{
			name:      "successful download with single user",
			userCount: 1,
			wantError: nil,
		},
		{
			name:      "successful download with multiple users",
			userCount: 3,
			wantError: nil,
		},
		{
			name:      "error from GetUsers",
			wantError: testutil.ErrMock,
		},
		{
			name:      "empty user list",
			userCount: 0,
			wantError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFb := mock.NewMockFbIf(ctrl)
			common.Fb = mockFb

			global.LocalUsers = make(map[string]*global.User)

			if tt.wantError != nil {
				mockFb.EXPECT().
					GetUsers(gomock.Any(), gomock.Any()).
					Return(nil, testutil.ErrMock).
					Times(1)
			} else {
				userRecords := make([]*auth.UserRecord, tt.userCount)
				for i := 0; i < tt.userCount; i++ {
					uid := "uid" + string(rune(i+1+48))
					email := "user" + string(rune(i+1+48)) + "@test.com"
					name := "User " + string(rune(i+1+48))

					userRecords[i] = &auth.UserRecord{
						UserInfo: &auth.UserInfo{
							UID:         uid,
							Email:       email,
							DisplayName: name,
						},
						CustomClaims: map[string]interface{}{"admin": true},
					}
				}

				mockFb.EXPECT().
					GetUsers(gomock.Any(), gomock.Any()).
					Return(&auth.GetUsersResult{Users: userRecords}, nil).
					Times(1)
			}

			uids := make([]auth.UserIdentifier, tt.userCount)
			for i := 0; i < tt.userCount; i++ {
				uid := "uid" + string(rune(i+1+48))
				uids[i] = auth.UIDIdentifier{UID: uid}
				global.LocalUsers[uid] = &global.User{UID: uid}
			}

			callCount := 0
			err := downloadClaims(uids, func(r *auth.UserRecord) error {
				callCount++
				return nil
			})

			assert.ErrorIs(t, err, tt.wantError)
			assert.Equal(t, tt.userCount, callCount)

			global.LocalUsers = make(map[string]*global.User)
		})
	}
}

func TestNewUserFromAuthClaimFiltering(t *testing.T) {
	tests := []struct {
		name           string
		authClaims     map[string]any
		expectedClaims common.ClaimsMap
		action         int
	}{
		{
			name:       "filters out non-permission claims",
			authClaims: map[string]any{"admin": true, "custom_data": "should_be_filtered", "editor": false},
			action:     actList,
		},
		{
			name:       "handles empty claims",
			authClaims: map[string]any{},
			action:     actList,
		},
		{
			name:       "handles nil claims",
			authClaims: nil,
			action:     actList,
		},
		{
			name:       "actSearch action",
			authClaims: map[string]any{"admin": true},
			action:     actSearch,
		},
		{
			name:       "actList action",
			authClaims: map[string]any{"admin": true},
			action:     actList,
		},
		{
			name:       "actSave action",
			authClaims: map[string]any{"admin": true},
			action:     actSave,
		},
		{
			name:       "actRefresh action",
			authClaims: map[string]any{"admin": true},
			action:     actRefresh,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFe := mock.NewMockFeIf(ctrl)
			common.Fe = mockFe
			testutil.InitLog()

			global.LocalUsers = make(map[string]*global.User)
			testuid := "test-uid-" + tt.name

			authRecord := &auth.UserRecord{
				UserInfo: &auth.UserInfo{
					UID:         testuid,
					Email:       "test@example.com",
					DisplayName: "Test User",
				},
				CustomClaims: tt.authClaims,
			}

			privileged := make(map[string]any)
			updates := make(map[string]any)

			user, err := newUserFromAuth(authRecord, tt.action, privileged, updates)

			assert.NoError(t, err)
			assert.NotNil(t, user)
			assert.Equal(t, testuid, user.UID)

			// Verify that claims were filtered (only known permissions kept)
			for k := range user.Claims {
				_, isKnown := common.PermsMap[k]
				assert.True(t, isKnown, "claim %s should be a known permission", k)
			}

			// Cleanup
			global.LocalUsers = make(map[string]*global.User)
		})
	}
}
