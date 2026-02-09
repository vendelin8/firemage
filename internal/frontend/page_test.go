package frontend

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

func TestShowPage(t *testing.T) {
	t.Run("ShowPage navigation", func(t *testing.T) {
		tests := []struct {
			name           string
			currentPage    string
			newPage        string
			savedUsers     map[string][]string
			crntUsers      []string
			actions        map[string]map[string]any
			wantErr        error
			wantUsers      []string
			wantSavedUsers map[string][]string
		}{
			{
				name:           "same page returns without changes",
				currentPage:    lang.PageSearch,
				newPage:        lang.PageSearch,
				wantErr:        nil,
				wantSavedUsers: map[string][]string{},
			},
			{
				name:        "switch from search to list with existing saved users",
				currentPage: lang.PageSearch,
				newPage:     lang.PageList,
				savedUsers: map[string][]string{
					lang.PageSearch: {"user1", "user2"},
					lang.PageList:   {"user3", "user4"},
				},
				crntUsers: []string{"current_search_users"},
				actions:   map[string]map[string]any{},
				wantErr:   nil,
				wantUsers: []string{"user3", "user4"},
				wantSavedUsers: map[string][]string{
					lang.PageSearch: {"current_search_users"},
					lang.PageList:   {"user3", "user4"},
				},
			},
			{
				name:        "switch to non-list page",
				currentPage: lang.PageList,
				newPage:     lang.PageSearch,
				savedUsers: map[string][]string{
					lang.PageList: {"user1"},
				},
				crntUsers: []string{},
				actions:   map[string]map[string]any{},
				wantErr:   nil,
				wantUsers: []string{},
				wantSavedUsers: map[string][]string{
					lang.PageList: {}},
			},
			{
				name:        "save users to old page on transition",
				currentPage: lang.PageList,
				newPage:     lang.PageSearch,
				savedUsers:  map[string][]string{},
				crntUsers:   []string{"user1", "user2", "user3"},
				actions:     map[string]map[string]any{},
				wantErr:     nil,
				wantUsers:   []string{},
				wantSavedUsers: map[string][]string{
					lang.PageList: {"user1", "user2", "user3"},
				},
			},
			{
				name:        "show warning when actions exist and switching to list page",
				currentPage: lang.PageSearch,
				newPage:     lang.PageList,
				savedUsers: map[string][]string{
					lang.PageList: {"user1"},
				},
				crntUsers: []string{},
				actions:   map[string]map[string]any{"uid1": {"admin": true}},
				wantErr:   nil,
				wantUsers: []string{"user1"},
				wantSavedUsers: map[string][]string{
					lang.PageSearch: {},
					lang.PageList:   {"user1"},
				},
			},
			{
				name:        "no warning when no actions on list page",
				currentPage: lang.PageSearch,
				newPage:     lang.PageList,
				savedUsers: map[string][]string{
					lang.PageList: {"user1"},
				},
				crntUsers: []string{},
				actions:   map[string]map[string]any{},
				wantErr:   nil,
				wantUsers: []string{"user1"},
				wantSavedUsers: map[string][]string{
					lang.PageSearch: {},
					lang.PageList:   {"user1"},
				},
			},
			{
				name:        "error when DoList fails",
				currentPage: lang.PageSearch,
				newPage:     lang.PageList,
				savedUsers:  map[string][]string{},
				crntUsers:   []string{},
				actions:     map[string]map[string]any{},
				wantErr:     testutil.ErrMock,
				wantUsers:   []string{},
				wantSavedUsers: map[string][]string{
					lang.PageSearch: {},
				},
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

				// Setup global state
				global.SavedUsers = tt.savedUsers
				global.CrntUsers = tt.crntUsers
				global.Actions = testutil.BuildActionsMap(tt.actions)

				mockFe.EXPECT().CurrentPage().Return(tt.currentPage).Times(1)

				// Setup mock expectations
				if tt.currentPage != tt.newPage {
					mockFe.EXPECT().SetPage(tt.newPage).Times(1)

					// Mock DoList when switching to list page with no cached users
					doListNeeded := tt.newPage == lang.PageList && len(tt.savedUsers[lang.PageList]) == 0
					if doListNeeded {
						mockFb.EXPECT().DoList().Return(tt.wantErr).Times(1)
					}

					// Only expect these if DoList succeeds or is not needed
					if !doListNeeded || tt.wantErr == nil {
						if len(tt.actions) > 0 {
							mockFe.EXPECT().ShowMsg(gomock.Any()).Times(1)
						}
						mockFe.EXPECT().LayoutUsers().Times(1)
					}
				}

				// Execute
				err := ShowPage(tt.newPage)

				// Assert
				assert.Equal(t, tt.wantErr, err)
				assert.Equal(t, tt.wantUsers, global.CrntUsers)
				if tt.currentPage != tt.newPage {
					assert.Equal(t, tt.wantSavedUsers, global.SavedUsers)
				}
			})
		}
	})
}
