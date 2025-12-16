package frontend

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vendelin8/firemage/internal/common"
	"github.com/vendelin8/firemage/internal/global"
	"github.com/vendelin8/firemage/internal/log"
	"github.com/vendelin8/firemage/internal/mock"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func init() {
	// Initialize logger for tests
	if log.Lgr == nil {
		log.Lgr, _ = zap.NewDevelopment()
	}
}

func claimsMapFromAny(data map[string]any) common.ClaimsMap {
	cm := make(common.ClaimsMap)
	for key, value := range data {
		// Handle time.Time values by converting to formatted string first
		var claimValue any = value
		if t, ok := value.(time.Time); ok {
			claimValue = t.Format(common.DateFormat)
		}
		if c, err := common.NewClaimFrom(claimValue); err == nil {
			cm[key] = c
		}
	}
	return cm
}

// TestOnActionChange is a comprehensive table-driven test for the onActionChange function
func TestOnActionChange(t *testing.T) {
	date1 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	date2 := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	date3 := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name            string
		currentClaims   map[string]any
		initialActions  map[string]any
		key             string
		checked         bool
		datePtr         *time.Time
		wantLayoutUsers int
		wantActions     map[string]any
	}{
		{
			name:            "RevertBooleanToTrue",
			currentClaims:   map[string]any{"admin": true},
			initialActions:  map[string]any{"admin": false},
			key:             "admin",
			checked:         true,
			datePtr:         nil,
			wantLayoutUsers: 0,
			wantActions:     nil,
		},
		{
			name:            "RevertBooleanFalse",
			currentClaims:   map[string]any{"admin": false},
			initialActions:  map[string]any{"admin": true},
			key:             "admin",
			checked:         false,
			datePtr:         nil,
			wantLayoutUsers: 0,
			wantActions:     nil,
		},
		{
			name:            "DateChange",
			currentClaims:   map[string]any{"expiry": date1},
			initialActions:  map[string]any{},
			key:             "expiry",
			checked:         false,
			datePtr:         &date3,
			wantLayoutUsers: 0,
			wantActions:     map[string]any{"expiry": date3},
		},
		{
			name:            "DateRevertToCurrent",
			currentClaims:   map[string]any{"expiry": date1},
			initialActions:  map[string]any{"expiry": date2},
			key:             "expiry",
			checked:         false,
			datePtr:         &date1,
			wantLayoutUsers: 0,
			wantActions:     nil,
		},
		{
			name:            "PartialRevert",
			currentClaims:   map[string]any{"admin": true, "superadmin": true},
			initialActions:  map[string]any{"admin": false, "superadmin": false},
			key:             "admin",
			checked:         true,
			datePtr:         nil,
			wantLayoutUsers: 0,
			wantActions:     map[string]any{"superadmin": false},
		},
		{
			name:            "TypeChangeFromBoolToDate",
			currentClaims:   map[string]any{"permission": true},
			initialActions:  map[string]any{},
			key:             "permission",
			checked:         false,
			datePtr:         &date2,
			wantLayoutUsers: 1,
			wantActions:     map[string]any{"permission": date2},
		},
		{
			name:            "TypeChangeFromDateToBool",
			currentClaims:   map[string]any{"permission": date1},
			initialActions:  map[string]any{},
			key:             "permission",
			checked:         false,
			datePtr:         nil,
			wantLayoutUsers: 1,
			wantActions:     map[string]any{"permission": false},
		},
		{
			name:            "StoreBooleanTrue",
			currentClaims:   map[string]any{"admin": false},
			initialActions:  map[string]any{},
			key:             "admin",
			checked:         true,
			datePtr:         nil,
			wantLayoutUsers: 0,
			wantActions:     map[string]any{"admin": true},
		},
		{
			name:            "MultipleActionsTracking",
			currentClaims:   map[string]any{"admin": true, "superadmin": true},
			initialActions:  map[string]any{"admin": false},
			key:             "superadmin",
			checked:         false,
			datePtr:         nil,
			wantLayoutUsers: 0,
			wantActions:     map[string]any{"admin": false, "superadmin": false},
		},
		{
			name:            "DateEarlyReturnCondition",
			currentClaims:   map[string]any{"expiry": date1},
			initialActions:  map[string]any{},
			key:             "expiry",
			checked:         false,
			datePtr:         &date1,
			wantLayoutUsers: 0,
			wantActions:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFe := mock.NewMockFeIf(ctrl)
			mockFe.EXPECT().LayoutUsers().Times(tt.wantLayoutUsers)
			// Allow any ReplaceTableItem calls by setting up a catch-all expectation
			mockFe.EXPECT().
				ReplaceTableItem(gomock.Any(), gomock.Any(), gomock.Any()).
				AnyTimes()
			common.Fe = mockFe

			uid := "user1"
			global.CrntUsers = []string{uid}
			global.LocalUsers = map[string]*global.User{
				uid: {
					UID:    uid,
					Email:  "user@test.com",
					Name:   "User",
					Claims: claimsMapFromAny(tt.currentClaims),
				},
			}
			global.Actions = map[string]common.ClaimsMap{
				uid: claimsMapFromAny(tt.initialActions),
			}

			// Construct the Claim from the test parameters
			claim := common.Claim{
				Checked: tt.checked,
				Date:    tt.datePtr,
			}

			onActionChange(0, tt.key, claim)

			// Convert expected result for comparison
			var expectedActions common.ClaimsMap
			if tt.wantActions != nil {
				expectedActions = claimsMapFromAny(tt.wantActions)
			}
			assert.Equal(t, expectedActions, global.Actions[uid])
		})
	}
}
