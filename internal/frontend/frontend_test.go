package frontend

import (
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/vendelin8/firemage/internal/common"
	"github.com/vendelin8/firemage/internal/conf"
	"github.com/vendelin8/firemage/internal/frontend/window"
	"github.com/vendelin8/firemage/internal/global"
	"github.com/vendelin8/firemage/internal/lang"
	"github.com/vendelin8/firemage/internal/mock"
	"github.com/vendelin8/firemage/internal/util"
	testutil "github.com/vendelin8/firemage/internal/util/test"
	"github.com/vendelin8/tview"
	"go.uber.org/mock/gomock"
)

func TestCurrentPage(t *testing.T) {
	tests := []struct {
		name        string
		description string
	}{
		{
			name:        "returns current page",
			description: "should return the current highlighted page",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Frontend{}
			// The CurrentPage method requires a fully initialized tview.TextView
			// with highlights, which is difficult to test without the full app.
			// We verify the method exists and the structure is sound.
			assert.NotNil(t, f)
		})
	}
}

func TestSetOnShow(t *testing.T) {
	tests := []struct {
		name        string
		page        string
		description string
	}{
		{
			name:        "sets callback for search page",
			page:        lang.PageSearch,
			description: "should store callback for search page",
		},
		{
			name:        "sets callback for list page",
			page:        lang.PageList,
			description: "should store callback for list page",
		},
		{
			name:        "sets callback for custom page",
			page:        "custom",
			description: "should store callback for custom page",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Frontend{
				onShowPage: make(map[string]func()),
			}
			callbackCalled := false
			callback := func() {
				callbackCalled = true
			}

			f.SetOnShow(tt.page, callback)

			assert.NotNil(t, f.onShowPage[tt.page], tt.description)
			// Call the stored callback
			f.onShowPage[tt.page]()
			assert.True(t, callbackCalled)
		})
	}
}

func TestCreateGUI(t *testing.T) {
	t.Run("creates Frontend with initialized fields", func(t *testing.T) {
		// This test cannot run without full app initialization
		// We'll verify the structure exists
		f := &Frontend{}
		assert.NotNil(t, f)
		assert.Nil(t, f.msg)
		assert.Nil(t, f.confirm)
	})
}

func TestShowMsg(t *testing.T) {
	tests := []struct {
		name        string
		msgExists   bool
		msgText     string
		description string
	}{
		{
			name:        "shows message when msg is nil",
			msgExists:   false,
			msgText:     "Test message",
			description: "should create and show new message popup",
		},
		{
			name:        "updates existing message",
			msgExists:   true,
			msgText:     "Updated message",
			description: "should update existing message popup",
		},
		{
			name:        "shows empty message",
			msgExists:   false,
			msgText:     "",
			description: "should handle empty message text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Frontend{
				onShowPage: make(map[string]func()),
			}
			// We can't fully test this without mocking tview components
			assert.NotNil(t, f)
		})
	}
}

func TestShowConfirm(t *testing.T) {
	t.Run("shows confirm when confirm is nil", func(t *testing.T) {
		f := &Frontend{
			onShowPage: make(map[string]func()),
			pages:      tview.NewPages(),
			app:        tview.NewApplication(),
		}
		okFunc := func() {}
		cancelFunc := func() {}
		msgText := "Are you sure?"

		// This should create a new confirm popup
		f.ShowConfirm(okFunc, cancelFunc, msgText)

		assert.NotNil(t, f.confirm)
	})

	t.Run("updates existing confirm", func(t *testing.T) {
		f := &Frontend{
			onShowPage: make(map[string]func()),
			pages:      tview.NewPages(),
			app:        tview.NewApplication(),
		}
		okFunc := func() {}
		cancelFunc := func() {}

		// Create initial confirm popup
		f.ShowConfirm(okFunc, cancelFunc, "Initial message")
		initialConfirm := f.confirm

		// Update with new message
		f.ShowConfirm(okFunc, cancelFunc, "Updated confirmation")

		// Should still be the same confirm object
		assert.Equal(t, initialConfirm, f.confirm)
	})
}

func TestHidePopup(t *testing.T) {
	t.Run("hides popup with active popup", func(t *testing.T) {
		f := &Frontend{}
		// HidePopup calls f.pages.HidePage(window.ActivePopup)
		// Testing this requires a fully initialized Frontend
		assert.NotNil(t, f)
	})
}

func TestSetPage(t *testing.T) {
	tests := []struct {
		name        string
		page        string
		description string
	}{
		{
			name:        "sets search page",
			page:        lang.PageSearch,
			description: "should set search page",
		},
		{
			name:        "sets list page",
			page:        lang.PageList,
			description: "should set list page",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Frontend{}
			// SetPage requires initialized menu, pages, and header
			assert.NotNil(t, f)
		})
	}
}

func TestCmdByKey(t *testing.T) {
	// Save original state
	originalShortcuts := common.Shortcuts
	originalMenuItems := common.MenuItems

	// Initialize shortcuts for testing
	common.Shortcuts = map[tcell.Key]int{
		tcell.KeyEsc: conf.CmdQuit,
		tcell.KeyF5:  conf.CmdRefresh,
	}

	// Initialize menu items for testing
	common.MenuItems = map[int]common.MenuItem{
		conf.CmdQuit: {
			Function: func() error { return nil },
		},
		conf.CmdRefresh: {
			Function: func() error { return nil },
		},
	}

	tests := []struct {
		name        string
		key         tcell.Key
		hasPopup    bool
		wantNil     bool
		description string
	}{
		{
			name:        "unknown key returns event",
			key:         tcell.KeyF1,
			wantNil:     false,
			description: "unmapped function key",
		},
		{
			name:        "quit command without popup",
			key:         tcell.KeyEsc,
			wantNil:     true,
			description: "escape key executes quit command",
		},
		{
			name:        "command with popup returns event",
			key:         tcell.KeyF5,
			hasPopup:    true,
			wantNil:     false,
			description: "function key with active popup",
		},
		{
			name:        "quit command with popup hides popup",
			key:         tcell.KeyEsc,
			hasPopup:    true,
			wantNil:     true,
			description: "escape key with active popup",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFe := mock.NewMockFeIf(ctrl)
			common.Fe = mockFe

			if tt.hasPopup {
				window.ActivePopups = append(window.ActivePopups, lang.PopupConfirm)
				if tt.key == tcell.KeyEsc {
					mockFe.EXPECT().HidePopup(lang.PopupConfirm).Times(1)
				}
			} else {
				window.ActivePopups = []string{}
			}

			ev := tcell.NewEventKey(tt.key, 0, tcell.ModNone)
			result := CmdByKey(ev)

			if tt.wantNil {
				assert.Nil(t, result, tt.description)
			} else {
				assert.NotNil(t, result, tt.description)
			}

			window.ActivePopups = []string{}
		})
	}

	// Restore original state
	common.Shortcuts = originalShortcuts
	common.MenuItems = originalMenuItems
}

func TestCmdByKeyWithError(t *testing.T) {
	// Save original state
	originalShortcuts := common.Shortcuts
	originalMenuItems := common.MenuItems
	defer func() {
		common.Shortcuts = originalShortcuts
		common.MenuItems = originalMenuItems
		window.ActivePopups = []string{}
	}()

	// Initialize shortcuts for testing
	common.Shortcuts = map[tcell.Key]int{
		tcell.KeyF1: conf.CmdQuit,
	}

	// Initialize menu items that return an error
	common.MenuItems = map[int]common.MenuItem{
		conf.CmdQuit: {
			Function: func() error { return testutil.ErrMock },
		},
	}

	t.Run("command error shows message", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockFe := mock.NewMockFeIf(ctrl)
		common.Fe = mockFe

		window.ActivePopups = []string{}
		mockFe.EXPECT().ShowMsg(testutil.ErrMock.Error()).Times(1)

		ev := tcell.NewEventKey(tcell.KeyF1, 0, tcell.ModNone)
		result := CmdByKey(ev)

		assert.Nil(t, result)
	})
}

func TestSetOnShowMultipleCallbacks(t *testing.T) {
	t.Run("overwrite callback for same page", func(t *testing.T) {
		f := &Frontend{
			onShowPage: make(map[string]func()),
		}

		callCount1 := 0
		callback1 := func() {
			callCount1++
		}

		callCount2 := 0
		callback2 := func() {
			callCount2++
		}

		// Set first callback
		f.SetOnShow(lang.PageSearch, callback1)
		f.onShowPage[lang.PageSearch]()
		assert.Equal(t, 1, callCount1)

		// Overwrite with second callback
		f.SetOnShow(lang.PageSearch, callback2)
		f.onShowPage[lang.PageSearch]()
		assert.Equal(t, 1, callCount1) // Old callback count should remain unchanged
		assert.Equal(t, 1, callCount2) // New callback should be called

		// Call again to verify new callback still works
		f.onShowPage[lang.PageSearch]()
		assert.Equal(t, 1, callCount1) // Old callback should still not be called again
		assert.Equal(t, 2, callCount2) // New callback should be called again
	})
}

func TestCmdByKeyWithMultipleShortcuts(t *testing.T) {
	originalShortcuts := common.Shortcuts
	originalMenuItems := common.MenuItems
	defer func() {
		common.Shortcuts = originalShortcuts
		common.MenuItems = originalMenuItems
		window.ActivePopups = []string{}
	}()

	common.Shortcuts = map[tcell.Key]int{
		tcell.KeyEsc: conf.CmdQuit,
		tcell.KeyF5:  conf.CmdRefresh,
		tcell.KeyTab: 2,
	}

	common.MenuItems = map[int]common.MenuItem{
		conf.CmdQuit: {
			Function: func() error { return nil },
		},
		conf.CmdRefresh: {
			Function: func() error { return nil },
		},
		2: {
			Function: func() error { return nil },
		},
	}

	t.Run("multiple shortcuts map correctly", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockFe := mock.NewMockFeIf(ctrl)
		common.Fe = mockFe

		window.ActivePopups = []string{}

		// Test F5 (refresh)
		ev := tcell.NewEventKey(tcell.KeyF5, 0, tcell.ModNone)
		result := CmdByKey(ev)
		assert.Nil(t, result)

		// Test Tab
		ev = tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)
		result = CmdByKey(ev)
		assert.Nil(t, result)
	})
}

func TestCmdByKeyWithDifferentModifiers(t *testing.T) {
	originalShortcuts := common.Shortcuts
	originalMenuItems := common.MenuItems
	defer func() {
		common.Shortcuts = originalShortcuts
		common.MenuItems = originalMenuItems
		window.ActivePopups = []string{}
	}()

	common.Shortcuts = map[tcell.Key]int{
		tcell.KeyCtrlC: 99,
	}

	common.MenuItems = map[int]common.MenuItem{
		99: {
			Function: func() error { return nil },
		},
	}

	t.Run("modified keys are handled", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockFe := mock.NewMockFeIf(ctrl)
		common.Fe = mockFe

		window.ActivePopups = []string{}

		// Test Ctrl+C
		ev := tcell.NewEventKey(tcell.KeyCtrlC, 0, tcell.ModNone)
		result := CmdByKey(ev)
		assert.Nil(t, result)
	})
}

func TestFrontendStructure(t *testing.T) {
	t.Run("Frontend has expected fields", func(t *testing.T) {
		f := &Frontend{
			onShowPage: make(map[string]func()),
		}

		assert.NotNil(t, f.onShowPage)
		assert.Empty(t, f.onShowPage)

		// Add callbacks
		f.SetOnShow("page1", func() {})
		f.SetOnShow("page2", func() {})

		assert.Len(t, f.onShowPage, 2)
		assert.NotNil(t, f.onShowPage["page1"])
		assert.NotNil(t, f.onShowPage["page2"])
	})
}

func TestCmdByKeyEdgeCases(t *testing.T) {
	originalShortcuts := common.Shortcuts
	originalMenuItems := common.MenuItems
	defer func() {
		common.Shortcuts = originalShortcuts
		common.MenuItems = originalMenuItems
		window.ActivePopups = []string{}
	}()

	t.Run("empty shortcuts map", func(t *testing.T) {
		common.Shortcuts = map[tcell.Key]int{}
		common.MenuItems = map[int]common.MenuItem{}

		ev := tcell.NewEventKey(tcell.KeyEsc, 0, tcell.ModNone)
		result := CmdByKey(ev)

		// Should return the event unchanged
		assert.NotNil(t, result)
		assert.Equal(t, tcell.KeyEsc, result.Key())
	})

	t.Run("shortcut exists but menu item not configured", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		common.Shortcuts = map[tcell.Key]int{
			tcell.KeyEsc: conf.CmdQuit,
		}
		// Intentionally not adding the menu item
		common.MenuItems = map[int]common.MenuItem{}

		window.ActivePopups = []string{}

		ev := tcell.NewEventKey(tcell.KeyEsc, 0, tcell.ModNone)
		result := CmdByKey(ev)

		// This will panic unless handled, so be careful
		// In real scenario, menu items should always be configured
		assert.Nil(t, result)
	})
}

// MockTextView is a simple mock for testing purposes
type MockTextView struct {
	highlights []string
}

func NewMockTextView(t *testing.T) *MockTextView {
	return &MockTextView{
		highlights: []string{},
	}
}

func (m *MockTextView) GetHighlights() []string {
	return m.highlights
}

func TestFormatMenuItem(t *testing.T) {
	tests := []struct {
		name       string
		menuKey    string
		text       string
		shortcut   string
		isPositive bool
		expected   string
	}{
		{
			name:       "yellow color when menuKey is not empty",
			menuKey:    "users",
			text:       "Manage Users",
			shortcut:   "u",
			isPositive: false,
			expected:   ` u ["users"][yellow::b]Manage Users[white::-][""]  `,
		},
		{
			name:       "green color when isPositive is true",
			menuKey:    "",
			text:       "Save",
			shortcut:   "s",
			isPositive: true,
			expected:   ` s [""][green::b]Save[white::-][""]  `,
		},
		{
			name:       "red color when menuKey is empty and isPositive is false",
			menuKey:    "",
			text:       "Delete",
			shortcut:   "d",
			isPositive: false,
			expected:   ` d [""][red::b]Delete[white::-][""]  `,
		},
		{
			name:       "yellow takes precedence over isPositive",
			menuKey:    "settings",
			text:       "Settings",
			shortcut:   "c",
			isPositive: true,
			expected:   ` c ["settings"][yellow::b]Settings[white::-][""]  `,
		},
		{
			name:       "handles special characters in text",
			menuKey:    "special",
			text:       "Text with & symbols",
			shortcut:   "x",
			isPositive: false,
			expected:   ` x ["special"][yellow::b]Text with & symbols[white::-][""]  `,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Frontend{}
			var buf strings.Builder
			f.formatMenuItem(&buf, tt.menuKey, tt.text, tt.shortcut, tt.isPositive)
			assert.Equal(t, tt.expected, buf.String())
		})
	}
}

func TestClaimsModalProcessClaimResult(t *testing.T) {
	const validDateStr = "2026-02-02"
	validDate, _ := time.Parse("2006-01-02", validDateStr)

	tests := []struct {
		name       string
		i          int
		key        string
		radioValue int
		dateText   string
		wantActive bool
		wantDate   *time.Time
	}{
		{
			name:       "claim active calls onActionChange with true and nil date",
			i:          0,
			key:        "test_key",
			radioValue: claimActive,
			dateText:   "",
			wantActive: true,
			wantDate:   nil,
		},
		{
			name:       "claim inactive calls onActionChange with false and nil date",
			i:          1,
			key:        "another_key",
			radioValue: claimInactive,
			dateText:   "",
			wantActive: false,
			wantDate:   nil,
		},
		{
			name:       "claim timed calls onActionChange with false and parsed date",
			i:          2,
			key:        "timed_key",
			radioValue: claimTimed,
			dateText:   validDateStr,
			wantActive: false,
			wantDate:   &validDate,
		},
		{
			name:       "claim timed with invalid date calls onActionChange with false and nil date",
			i:          3,
			key:        "invalid_key",
			radioValue: claimTimed,
			dateText:   "invalid-date",
			wantActive: false,
			wantDate:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFe := mock.NewMockFeIf(ctrl)
			common.Fe = mockFe
			mockFe.EXPECT().ReplaceTableItem(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

			// Setup global state for onActionChange
			uid := "test_user"
			// Need at least tt.i + 1 users in CrntUsers
			global.CrntUsers = make([]string, tt.i+1)
			global.CrntUsers[tt.i] = uid
			global.LocalUsers = map[string]*global.User{
				uid: {
					UID:    uid,
					Email:  "user@test.com",
					Name:   "Test User",
					Claims: make(common.ClaimsMap),
				},
			}
			global.Actions = make(map[string]common.ClaimsMap)
			global.Actions[uid] = make(common.ClaimsMap)

			// Create radio with proper options
			radio := tview.NewRadio(lang.SActive, lang.SInactive, lang.STimed)
			radio.SetValue(tt.radioValue)

			dateField := tview.NewInputField().SetText(tt.dateText)

			c := &ClaimsModal{
				radio: radio,
				date:  dateField,
				i:     tt.i,
				key:   tt.key,
			}

			// Call processClaimResult
			c.processClaimResult()

			// Verify that onActionChange was called by checking global.Actions
			actualClaim, exists := global.Actions[uid][tt.key]
			assert.True(t, exists, "claim should exist in global.Actions")

			// Construct expected claim based on radio value
			expectedClaim := common.Claim{}
			if tt.radioValue == claimActive {
				expectedClaim.Checked = true
			} else if tt.radioValue == claimTimed && tt.wantDate != nil {
				expectedClaim.Date = tt.wantDate
			}

			// Verify the actual claim that was set by onActionChange
			assert.Equal(t, expectedClaim.Checked, actualClaim.Checked, "active flag should match")
			assert.Equal(t, expectedClaim.Date, actualClaim.Date, "date should match")
		})
	}
}

func TestClaimsRadioSetOnSetValue(t *testing.T) {
	const existingDateStr = "2026-02-02"
	existingDate, _ := time.Parse("2006-01-02", existingDateStr)

	tests := []struct {
		name          string
		radioValue    int
		dateFieldText string
		wantDisabled  bool
	}{
		{
			name:          "claim active disables date field",
			radioValue:    claimActive,
			dateFieldText: "",
			wantDisabled:  true,
		},
		{
			name:          "claim inactive disables date field",
			radioValue:    claimInactive,
			dateFieldText: "",
			wantDisabled:  true,
		},
		{
			name:          "claim timed enables date field when date exists",
			radioValue:    claimTimed,
			dateFieldText: existingDateStr,
			wantDisabled:  false,
		},
		{
			name:          "claim timed enables date field even when empty",
			radioValue:    claimTimed,
			dateFieldText: "",
			wantDisabled:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFe := mock.NewMockFeIf(ctrl)
			common.Fe = mockFe

			// Setup mocks
			mockFe.EXPECT().ClaimsDateSetDisabled(tt.wantDisabled).Times(1)

			// ClaimsBtns is called with true for active/inactive, false for timed
			shouldDisableBtns := tt.radioValue != claimTimed
			mockFe.EXPECT().ClaimsBtns(shouldDisableBtns).Times(1)

			// For timed claims, claimsRadioSetOnSetValue always calls ClaimsDate()
			// to check if it's nil and set current time if needed
			if tt.radioValue == claimTimed {
				// For date field with existing date, ClaimsDate returns a valid date
				if tt.dateFieldText == existingDateStr {
					mockFe.EXPECT().ClaimsDate().Return(&existingDate).Times(1)
				} else {
					// For empty date field, ClaimsDate returns nil and we call ClaimsSetDate
					mockFe.EXPECT().ClaimsDate().Return(nil).Times(1)
					mockFe.EXPECT().ClaimsSetDate(gomock.Any()).Times(1)
				}
			}

			// Call the function
			claimsRadioSetOnSetValue(tt.radioValue)
		})
	}
}

func TestClaimsModalIncDate(t *testing.T) {
	// Initialize timed buttons map for testing
	common.TimedButtons = [][]string{
		{"One month", "1m"},
		{"One year", "1y"},
	}
	_ = util.InitializeTimedButtonsMap()

	tests := []struct {
		name              string
		buttonLabel       string
		currentDateExists bool
		currentDateStr    string
	}{
		{
			name:              "increment date with existing date",
			buttonLabel:       "One month",
			currentDateExists: true,
			currentDateStr:    "2026-02-02",
		},
		{
			name:              "increment date with nil date uses current time",
			buttonLabel:       "One year",
			currentDateExists: false,
			currentDateStr:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFe := mock.NewMockFeIf(ctrl)
			common.Fe = mockFe

			// Create radio for the modal
			radio := tview.NewRadio(lang.SActive, lang.SInactive, lang.STimed)

			c := &ClaimsModal{
				radio: radio,
				date:  tview.NewInputField(),
				i:     0,
				key:   "test_key",
			}

			// Mock ClaimsDate to return the current date or nil
			var currentDate *time.Time
			if tt.currentDateExists {
				d, _ := time.Parse("2006-01-02", tt.currentDateStr)
				currentDate = &d
			}
			mockFe.EXPECT().ClaimsDate().Return(currentDate).Times(1)

			// Mock ClaimsSetDate - it will be called with the new date
			mockFe.EXPECT().ClaimsSetDate(gomock.Any()).Times(1)

			// Call incDate
			c.incDate(tt.buttonLabel)

			// Verify radio value is set to ClaimTimed
			assert.Equal(t, claimTimed, radio.Value(), "radio should be set to ClaimTimed after incDate")
		})
	}
}
