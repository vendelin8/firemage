package firebase

import (
	"testing"

	"firebase.google.com/go/auth"
	"github.com/stretchr/testify/assert"
	"github.com/vendelin8/firemage/internal/common"
	"github.com/vendelin8/firemage/internal/global"
	"github.com/vendelin8/firemage/internal/mock"
	testutil "github.com/vendelin8/firemage/internal/util/test"
	"go.uber.org/mock/gomock"
)

func TestNewUserFromAuth(t *testing.T) {
	tests := []struct {
		name              string
		userInCache       bool
		cachedClaims      map[string]any
		authClaims        map[string]any
		action            int
		shouldShowConfirm bool
		wantError         error
	}{
		{
			name:        "new user not in cache",
			userInCache: false,
			authClaims:  map[string]any{"admin": true, "editor": false},
			action:      actList,
		},
		{
			name:        "new user not in cache with empty claims and actRefresh",
			userInCache: false,
			authClaims:  map[string]any{},
			action:      actRefresh,
		},
		{
			name:        "new user not in cache with claims and actRefresh",
			userInCache: false,
			authClaims:  map[string]any{"admin": true},
			action:      actRefresh,
		},
		{
			name:         "user in cache with same claims",
			userInCache:  true,
			cachedClaims: map[string]any{"admin": true},
			authClaims:   map[string]any{"admin": true},
			action:       actList,
		},
		{
			name:              "user in cache with different claims",
			userInCache:       true,
			cachedClaims:      map[string]any{"admin": true},
			authClaims:        map[string]any{"admin": false, "editor": true},
			action:            actSearch,
			shouldShowConfirm: true,
		},
		{
			name:        "user with no custom claims",
			userInCache: false,
			authClaims:  map[string]any{},
			action:      actList,
		},
		{
			name:              "set permissions error",
			userInCache:       true,
			cachedClaims:      map[string]any{"admin": true},
			authClaims:        map[string]any{"admin": false},
			action:            actSave,
			shouldShowConfirm: true,
			wantError:         testutil.ErrMock,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFe := mock.NewMockFeIf(ctrl)
			common.Fe = mockFe

			mockFb := mock.NewMockFbIf(ctrl)
			common.Fb = mockFb

			testutil.InitLog()

			global.LocalUsers = make(map[string]*global.User)

			testuid := "test-uid-123"
			testEmail := "test@example.com"
			testName := "Test User"

			// Set up cached user if needed
			if tt.userInCache {
				cm, _ := common.NewClaimsMapFrom(tt.cachedClaims)
				global.LocalUsers[testuid] = &global.User{
					UID:    testuid,
					Email:  testEmail,
					Name:   testName,
					Claims: *cm,
				}
			}

			// Create auth user record
			authRecord := &auth.UserRecord{
				UserInfo: &auth.UserInfo{
					UID:         testuid,
					Email:       testEmail,
					DisplayName: testName,
				},
				CustomClaims: tt.authClaims,
			}

			privileged := make(map[string]any)
			updates := make(map[string]any)

			// Expect ShowConfirm if claims differ
			if tt.shouldShowConfirm {
				mockFb.EXPECT().
					StoreAuthClaims(gomock.Any(), testuid, gomock.Any()).
					Return(tt.wantError).
					Times(1)

				confirmCallback := func(onYes, onNo func(), ms ...string) {
					// Simulate user clicking OK
					if onYes != nil {
						onYes()
					}
				}

				mockFe.EXPECT().
					ShowConfirm(gomock.Any(), gomock.Any(), gomock.Any()).
					Do(confirmCallback).
					Times(1)
			}

			user, err := newUserFromAuth(authRecord, tt.action, privileged, updates)

			assert.ErrorIs(t, err, tt.wantError)
			assert.NotNil(t, user)
			assert.Equal(t, testuid, user.UID)
			assert.Equal(t, testEmail, user.Email)
			assert.Equal(t, testName, user.Name)

			// Verify caching behavior
			if !tt.userInCache && !tt.shouldShowConfirm {
				if tt.action != actRefresh || len(tt.authClaims) > 0 {
					assert.Equal(t, user, global.LocalUsers[testuid], "user should be cached")
				}
			}

			// Cleanup
			global.LocalUsers = make(map[string]*global.User)
		})
	}
}
