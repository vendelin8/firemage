package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/vendelin8/firemage/internal/common"
	"github.com/vendelin8/firemage/internal/global"
	"github.com/vendelin8/firemage/internal/lang"
	"github.com/vendelin8/firemage/internal/mock"
	testutil "github.com/vendelin8/firemage/internal/util/test"
)

func TestCancel(t *testing.T) {
	cleanup := testutil.InitLog()
	defer cleanup()

	tests := []struct {
		name         string
		setupActions map[string]map[string]any
		wantError    error
	}{
		{
			name:         "cancel with pending actions",
			setupActions: map[string]map[string]any{"uid1": {"admin": true}},
		},
		{
			name:         "cancel with single action",
			setupActions: map[string]map[string]any{"uid1": {"admin": true}},
		},
		{
			name:         "cancel with no pending actions",
			setupActions: map[string]map[string]any{},
			wantError:    ErrNoChanges,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFe := mock.NewMockFeIf(ctrl)
			common.Fe = mockFe

			global.Actions = testutil.BuildActionsMap(tt.setupActions)

			if tt.wantError == nil {
				mockFe.EXPECT().LayoutUsers().Times(1)
			}

			err := cancel()
			assert.Equal(t, tt.wantError, err)
			assert.Equal(t, 0, len(global.Actions))
		})
	}
}

func TestRefresh(t *testing.T) {
	cleanup := testutil.InitLog()
	defer cleanup()

	tests := []struct {
		name            string
		setupUsers      []string
		setupActions    map[string]map[string]any
		wantCurrentPage string
		wantError       error
	}{
		{
			name:            "refresh on list page with users and no actions",
			setupUsers:      []string{"uid1", "uid2"},
			wantCurrentPage: lang.PageList,
			wantError:       nil,
		},
		{
			name:            "refresh not on list page",
			setupUsers:      []string{"uid1"},
			wantCurrentPage: lang.PageSearch,
			wantError:       ErrCantRefresh,
		},
		{
			name:            "refresh with pending actions",
			setupUsers:      []string{"uid1"},
			setupActions:    map[string]map[string]any{"uid1": {"admin": true}},
			wantCurrentPage: lang.PageList,
			wantError:       ErrActions,
		},
		{
			name:            "refresh with no users",
			wantCurrentPage: lang.PageList,
			wantError:       nil,
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

			mockFe.EXPECT().CurrentPage().Return(tt.wantCurrentPage).Times(1)

			if tt.wantError == nil {
				mockFb.EXPECT().RunTransaction(gomock.Any(), gomock.Any()).Return(nil).Times(1)

				if len(tt.setupUsers) == 0 {
					mockFe.EXPECT().ShowMsg(gomock.Any()).Times(1)
				}
			}

			global.CrntUsers = tt.setupUsers
			global.Actions = testutil.BuildActionsMap(tt.setupActions)

			err := refresh()
			assert.Equal(t, tt.wantError, err)
			global.Actions = map[string]common.ClaimsMap{}
		})
	}
}

func TestSave(t *testing.T) {
	cleanup := testutil.InitLog()
	defer cleanup()

	tests := []struct {
		name         string
		setupActions map[string]map[string]any
		wantError    bool
	}{
		{
			name:         "save with single action",
			setupActions: map[string]map[string]any{"uid1": {"admin": true}},
		},
		{
			name:         "save with multiple actions",
			setupActions: map[string]map[string]any{"uid1": {"admin": true}, "uid2": {"moderator": false}},
		},
		{
			name:         "save with no pending actions",
			setupActions: map[string]map[string]any{},
			wantError:    true,
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

			if !tt.wantError {
				mockFe.EXPECT().ShowMsg(gomock.Any()).Times(1)
				// Mock RunTransaction for successful save
				mockFb.EXPECT().
					RunTransaction(gomock.Any(), gomock.Any()).
					Return(nil).
					Times(1)
			}

			global.Actions = testutil.BuildActionsMap(tt.setupActions)
			err := save()

			if tt.wantError {
				assert.Error(t, err)
				assert.Equal(t, ErrNoChanges, err)
			} else {
				assert.NoError(t, err)
			}

			global.Actions = map[string]common.ClaimsMap{}
		})
	}
}

func TestShowListAndSearch(t *testing.T) {
	cleanup := testutil.InitLog()
	defer cleanup()

	tests := []struct {
		name        string
		page        string
		currentPage string
		savedUsers  map[string][]string
		crntUsers   []string
		setup       func(mockFb *mock.MockFbIf)
		wantError   error
	}{
		{
			name:        "show list with cached saved users",
			page:        lang.PageList,
			currentPage: lang.PageSearch,
			savedUsers:  map[string][]string{lang.PageList: {"user1", "user2"}},
			crntUsers:   []string{"search_user1"},
		},
		{
			name:        "show list with empty cached saved users",
			page:        lang.PageList,
			currentPage: lang.PageSearch,
			savedUsers:  map[string][]string{lang.PageList: {}},
			crntUsers:   []string{"search_user1"},
		},
		{
			name:        "show search from list page with users",
			page:        lang.PageSearch,
			currentPage: lang.PageList,
			savedUsers:  map[string][]string{},
			crntUsers:   []string{"list_user1", "list_user2"},
		},
		{
			name:        "show search with no previous users",
			page:        lang.PageSearch,
			currentPage: lang.PageList,
			savedUsers:  map[string][]string{},
			crntUsers:   []string{},
		},
		{
			name:        "show list saves current users to old page before switching",
			page:        lang.PageList,
			currentPage: lang.PageSearch,
			savedUsers:  map[string][]string{lang.PageList: {"saved_user1", "saved_user2"}},
			crntUsers:   []string{"current_search_user"},
		},
		{
			name:        "show list with mock error",
			page:        lang.PageList,
			currentPage: lang.PageSearch,
			savedUsers:  map[string][]string{},
			crntUsers:   []string{},
			setup: func(mockFb *mock.MockFbIf) {
				mockFb.EXPECT().DoList().Return(testutil.ErrMock).Times(1)
			},
			wantError: testutil.ErrMock,
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
			mockFe.EXPECT().CurrentPage().Return(tt.currentPage).Times(1)
			mockFe.EXPECT().SetPage(tt.page).Times(1)

			if tt.setup != nil {
				tt.setup(mockFb)
			}

			global.CrntUsers = tt.crntUsers
			global.SavedUsers = tt.savedUsers
			global.LocalPrivileged = make(map[string]struct{})
			global.LocalUsers = make(map[string]*global.User)

			// LayoutUsers should only be called on success
			if tt.wantError == nil {
				mockFe.EXPECT().LayoutUsers().Times(1)
			}

			var err error
			if tt.page == lang.PageList {
				err = showList()
			} else {
				err = showSearch()
			}

			assert.ErrorIs(t, err, tt.wantError)
		})
	}
}
