package conf

import (
	"io"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/vendelin8/firemage/internal/common"
)

func TestSaveShortcuts(t *testing.T) {
	tests := []struct {
		name           string
		setupMenuItems map[int]common.MenuItem
		wantCallCount  int
	}{
		{
			name: "default shortcuts",
			setupMenuItems: map[int]common.MenuItem{
				CmdSearch: {
					Shortcut: "F2",
					Keys:     []tcell.Key{tcell.KeyF2},
					MenuKey:  "search",
					Text:     "Search",
					Positive: false,
					IsDef:    true,
				},
				CmdList: {
					Shortcut: "F3",
					Keys:     []tcell.Key{tcell.KeyF3},
					MenuKey:  "list",
					Text:     "List",
					Positive: false,
					IsDef:    true,
				},
				CmdRefresh: {
					Shortcut: "F5",
					Keys:     []tcell.Key{tcell.KeyF5},
					MenuKey:  "",
					Text:     "Refresh",
					Positive: true,
					IsDef:    true,
				},
				CmdSave: {
					Shortcut: "F6",
					Keys:     []tcell.Key{tcell.KeyF6},
					MenuKey:  "",
					Text:     "Save",
					Positive: true,
					IsDef:    true,
				},
				CmdCancel: {
					Shortcut: "F8",
					Keys:     []tcell.Key{tcell.KeyF8},
					MenuKey:  "",
					Text:     "Cancel",
					Positive: false,
					IsDef:    true,
				},
				CmdQuit: {
					Shortcut: "Esc",
					Keys:     []tcell.Key{tcell.KeyEsc},
					MenuKey:  "",
					Text:     "Quit",
					Positive: false,
					IsDef:    true,
				},
			},
			wantCallCount: 6,
		},
		{
			name: "custom shortcut with multiple keys",
			setupMenuItems: map[int]common.MenuItem{
				CmdSearch: {
					Shortcut: "F2",
					Keys:     []tcell.Key{tcell.KeyF2, tcell.KeyCtrlS},
					MenuKey:  "search",
					Text:     "Search",
					Positive: false,
					IsDef:    false,
				},
				CmdList: {
					Shortcut: "F3",
					Keys:     []tcell.Key{tcell.KeyF3},
					MenuKey:  "list",
					Text:     "List",
					Positive: false,
					IsDef:    true,
				},
				CmdRefresh: {
					Shortcut: "F5",
					Keys:     []tcell.Key{tcell.KeyF5},
					MenuKey:  "",
					Text:     "Refresh",
					Positive: true,
					IsDef:    true,
				},
				CmdSave: {
					Shortcut: "F6",
					Keys:     []tcell.Key{tcell.KeyF6},
					MenuKey:  "",
					Text:     "Save",
					Positive: true,
					IsDef:    true,
				},
				CmdCancel: {
					Shortcut: "F8",
					Keys:     []tcell.Key{tcell.KeyF8},
					MenuKey:  "",
					Text:     "Cancel",
					Positive: false,
					IsDef:    true,
				},
				CmdQuit: {
					Shortcut: "Esc",
					Keys:     []tcell.Key{tcell.KeyEsc},
					MenuKey:  "",
					Text:     "Quit",
					Positive: false,
					IsDef:    true,
				},
			},
			wantCallCount: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original common.MenuItems
			originalMenuItems := common.MenuItems
			originalShortcuts := common.Shortcuts

			// Setup test common.MenuItems
			common.MenuItems = tt.setupMenuItems
			common.Shortcuts = make(map[tcell.Key]int)

			callCount := 0
			saveShortcuts(func(menuKey, text, shortcut string, isPositive bool) {
				callCount++
				assert.NotEmpty(t, text)
				assert.NotEmpty(t, shortcut)
			})

			assert.Equal(t, tt.wantCallCount, callCount)

			// Verify shortcuts map was populated
			assert.Greater(t, len(common.Shortcuts), 0)

			// Restore original common.MenuItems
			common.MenuItems = originalMenuItems
			common.Shortcuts = originalShortcuts
		})
	}
}

func TestLoadYamlConf(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantKc    map[string]any
		wantError bool
		wantMsg   string
	}{
		{
			name:   "nil reader returns nil",
			input:  "",
			wantKc: nil,
		},
		{
			name: "valid shortcuts config",
			input: `keyboardShortcuts:
  F2: Search
  F3: List
  F5: Refresh`,
			wantKc: map[string]any{
				"F2": "Search",
				"F3": "List",
				"F5": "Refresh",
			},
		},
		{
			name:   "empty shortcuts section",
			input:  `keyboardShortcuts: {}`,
			wantKc: map[string]any{},
		},
		{
			name:      "invalid yaml",
			input:     "invalid: [unclosed",
			wantError: true,
			wantMsg:   "error while parsing config file",
		},
		{
			name:      "missing shortcuts section",
			input:     "other: value",
			wantError: true,
			wantMsg:   "config file is invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reader strings.Reader
			if tt.input != "" {
				reader = *strings.NewReader(tt.input)
			}

			var kc map[string]any
			var err error
			if tt.input == "" {
				kc, err = loadYamlConf(nil)
			} else {
				kc, err = loadYamlConf(&reader)
			}

			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantKc, kc)
			}
		})
	}
}

func TestLoadConf(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		setupMenuItems   map[int]common.MenuItem
		wantError        bool
		wantCallbackCall bool
		wantMenuModified bool
	}{
		{
			name:  "nil reader with default menu items",
			input: "",
			setupMenuItems: map[int]common.MenuItem{
				CmdSearch: {
					Shortcut: "F2",
					Keys:     []tcell.Key{tcell.KeyF2},
					MenuKey:  "search",
					Text:     "Search",
					Positive: false,
					IsDef:    true,
				},
				CmdList: {
					Shortcut: "F3",
					Keys:     []tcell.Key{tcell.KeyF3},
					MenuKey:  "list",
					Text:     "List",
					Positive: false,
					IsDef:    true,
				},
				CmdRefresh: {
					Shortcut: "F5",
					Keys:     []tcell.Key{tcell.KeyF5},
					MenuKey:  "",
					Text:     "Refresh",
					Positive: true,
					IsDef:    true,
				},
				CmdSave: {
					Shortcut: "F6",
					Keys:     []tcell.Key{tcell.KeyF6},
					MenuKey:  "",
					Text:     "Save",
					Positive: true,
					IsDef:    true,
				},
				CmdCancel: {
					Shortcut: "F8",
					Keys:     []tcell.Key{tcell.KeyF8},
					MenuKey:  "",
					Text:     "Cancel",
					Positive: false,
					IsDef:    true,
				},
				CmdQuit: {
					Shortcut: "Esc",
					Keys:     []tcell.Key{tcell.KeyEsc},
					MenuKey:  "",
					Text:     "Quit",
					Positive: false,
					IsDef:    true,
				},
			},
			wantCallbackCall: true,
		},
		{
			name:  "empty keyboard shortcuts",
			input: `keyboardShortcuts: {}`,
			setupMenuItems: map[int]common.MenuItem{
				CmdSearch: {
					Shortcut: "F2",
					Keys:     []tcell.Key{tcell.KeyF2},
					MenuKey:  "search",
					Text:     "Search",
					Positive: false,
					IsDef:    true,
				},
				CmdList: {
					Shortcut: "F3",
					Keys:     []tcell.Key{tcell.KeyF3},
					MenuKey:  "list",
					Text:     "List",
					Positive: false,
					IsDef:    true,
				},
				CmdRefresh: {
					Shortcut: "F5",
					Keys:     []tcell.Key{tcell.KeyF5},
					MenuKey:  "",
					Text:     "Refresh",
					Positive: true,
					IsDef:    true,
				},
				CmdSave: {
					Shortcut: "F6",
					Keys:     []tcell.Key{tcell.KeyF6},
					MenuKey:  "",
					Text:     "Save",
					Positive: true,
					IsDef:    true,
				},
				CmdCancel: {
					Shortcut: "F8",
					Keys:     []tcell.Key{tcell.KeyF8},
					MenuKey:  "",
					Text:     "Cancel",
					Positive: false,
					IsDef:    true,
				},
				CmdQuit: {
					Shortcut: "Esc",
					Keys:     []tcell.Key{tcell.KeyEsc},
					MenuKey:  "",
					Text:     "Quit",
					Positive: false,
					IsDef:    true,
				},
			},
			wantCallbackCall: true,
		},
		{
			name:  "invalid keyboard shortcut key",
			input: `keyboardShortcuts: { invalidKey: Search }`,
			setupMenuItems: map[int]common.MenuItem{
				CmdSearch: {
					Shortcut: "F2",
					Keys:     []tcell.Key{tcell.KeyF2},
					MenuKey:  "search",
					Text:     "Search",
					Positive: false,
					IsDef:    true,
				},
				CmdList: {
					Shortcut: "F3",
					Keys:     []tcell.Key{tcell.KeyF3},
					MenuKey:  "list",
					Text:     "List",
					Positive: false,
					IsDef:    true,
				},
				CmdRefresh: {
					Shortcut: "F5",
					Keys:     []tcell.Key{tcell.KeyF5},
					MenuKey:  "",
					Text:     "Refresh",
					Positive: true,
					IsDef:    true,
				},
				CmdSave: {
					Shortcut: "F6",
					Keys:     []tcell.Key{tcell.KeyF6},
					MenuKey:  "",
					Text:     "Save",
					Positive: true,
					IsDef:    true,
				},
				CmdCancel: {
					Shortcut: "F8",
					Keys:     []tcell.Key{tcell.KeyF8},
					MenuKey:  "",
					Text:     "Cancel",
					Positive: false,
					IsDef:    true,
				},
				CmdQuit: {
					Shortcut: "Esc",
					Keys:     []tcell.Key{tcell.KeyEsc},
					MenuKey:  "",
					Text:     "Quit",
					Positive: false,
					IsDef:    true,
				},
			},
			wantError:        true,
			wantCallbackCall: true,
		},
		{
			name:  "invalid keyboard command text",
			input: `keyboardShortcuts: { F10: InvalidCommand }`,
			setupMenuItems: map[int]common.MenuItem{
				CmdSearch: {
					Shortcut: "F2",
					Keys:     []tcell.Key{tcell.KeyF2},
					MenuKey:  "search",
					Text:     "Search",
					Positive: false,
					IsDef:    true,
				},
				CmdList: {
					Shortcut: "F3",
					Keys:     []tcell.Key{tcell.KeyF3},
					MenuKey:  "list",
					Text:     "List",
					Positive: false,
					IsDef:    true,
				},
				CmdRefresh: {
					Shortcut: "F5",
					Keys:     []tcell.Key{tcell.KeyF5},
					MenuKey:  "",
					Text:     "Refresh",
					Positive: true,
					IsDef:    true,
				},
				CmdSave: {
					Shortcut: "F6",
					Keys:     []tcell.Key{tcell.KeyF6},
					MenuKey:  "",
					Text:     "Save",
					Positive: true,
					IsDef:    true,
				},
				CmdCancel: {
					Shortcut: "F8",
					Keys:     []tcell.Key{tcell.KeyF8},
					MenuKey:  "",
					Text:     "Cancel",
					Positive: false,
					IsDef:    true,
				},
				CmdQuit: {
					Shortcut: "Esc",
					Keys:     []tcell.Key{tcell.KeyEsc},
					MenuKey:  "",
					Text:     "Quit",
					Positive: false,
					IsDef:    true,
				},
			},
			wantError:        true,
			wantCallbackCall: true,
		},
		{
			name:  "multiple invalid keyboard commands",
			input: `keyboardShortcuts: { F10: Command1, F11: Command2 }`,
			setupMenuItems: map[int]common.MenuItem{
				CmdSearch: {
					Shortcut: "F2",
					Keys:     []tcell.Key{tcell.KeyF2},
					MenuKey:  "search",
					Text:     "Search",
					Positive: false,
					IsDef:    true,
				},
				CmdList: {
					Shortcut: "F3",
					Keys:     []tcell.Key{tcell.KeyF3},
					MenuKey:  "list",
					Text:     "List",
					Positive: false,
					IsDef:    true,
				},
				CmdRefresh: {
					Shortcut: "F5",
					Keys:     []tcell.Key{tcell.KeyF5},
					MenuKey:  "",
					Text:     "Refresh",
					Positive: true,
					IsDef:    true,
				},
				CmdSave: {
					Shortcut: "F6",
					Keys:     []tcell.Key{tcell.KeyF6},
					MenuKey:  "",
					Text:     "Save",
					Positive: true,
					IsDef:    true,
				},
				CmdCancel: {
					Shortcut: "F8",
					Keys:     []tcell.Key{tcell.KeyF8},
					MenuKey:  "",
					Text:     "Cancel",
					Positive: false,
					IsDef:    true,
				},
				CmdQuit: {
					Shortcut: "Esc",
					Keys:     []tcell.Key{tcell.KeyEsc},
					MenuKey:  "",
					Text:     "Quit",
					Positive: false,
					IsDef:    true,
				},
			},
			wantError:        true,
			wantCallbackCall: true,
		},
		{
			name:  "valid custom shortcut mapping",
			input: `keyboardShortcuts: { F10: Search }`,
			setupMenuItems: map[int]common.MenuItem{
				CmdSearch: {
					Shortcut: "F2",
					Keys:     []tcell.Key{tcell.KeyF2},
					MenuKey:  "search",
					Text:     "Search",
					Positive: false,
					IsDef:    true,
				},
				CmdList: {
					Shortcut: "F3",
					Keys:     []tcell.Key{tcell.KeyF3},
					MenuKey:  "list",
					Text:     "List",
					Positive: false,
					IsDef:    true,
				},
				CmdRefresh: {
					Shortcut: "F5",
					Keys:     []tcell.Key{tcell.KeyF5},
					MenuKey:  "",
					Text:     "Refresh",
					Positive: true,
					IsDef:    true,
				},
				CmdSave: {
					Shortcut: "F6",
					Keys:     []tcell.Key{tcell.KeyF6},
					MenuKey:  "",
					Text:     "Save",
					Positive: true,
					IsDef:    true,
				},
				CmdCancel: {
					Shortcut: "F8",
					Keys:     []tcell.Key{tcell.KeyF8},
					MenuKey:  "",
					Text:     "Cancel",
					Positive: false,
					IsDef:    true,
				},
				CmdQuit: {
					Shortcut: "Esc",
					Keys:     []tcell.Key{tcell.KeyEsc},
					MenuKey:  "",
					Text:     "Quit",
					Positive: false,
					IsDef:    true,
				},
			},
			wantCallbackCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original state
			originalMenuItems := common.MenuItems
			originalShortcuts := common.Shortcuts

			// Setup test common.MenuItems
			common.MenuItems = tt.setupMenuItems
			common.Shortcuts = make(map[tcell.Key]int)

			var ioReader interface{} = nil
			if tt.input != "" {
				ioReader = strings.NewReader(tt.input)
			}

			callbackCalled := false
			var actualErr error
			if ioReader != nil {
				actualErr = loadConf(func(menuKey, text, shortcut string, isPositive bool) {
					callbackCalled = true
				}, ioReader.(io.Reader))
			} else {
				actualErr = loadConf(func(menuKey, text, shortcut string, isPositive bool) {
					callbackCalled = true
				}, nil)
			}

			if tt.wantError {
				assert.Error(t, actualErr, "expected error in test case: %s", tt.name)
			} else {
				assert.NoError(t, actualErr, "unexpected error in test case: %s", tt.name)
			}

			if tt.wantCallbackCall {
				assert.True(t, callbackCalled, "callback should have been called in test case: %s", tt.name)
			}

			// Verify shortcuts map was populated
			assert.Greater(t, len(common.Shortcuts), 0, "shortcuts map should be populated in test case: %s", tt.name)

			// Restore original state
			common.MenuItems = originalMenuItems
			common.Shortcuts = originalShortcuts
		})
	}
}

func TestLoadConfErrorMessages(t *testing.T) {
	// Helper function to create a test menu items map
	createTestMenuItems := func() map[int]common.MenuItem {
		return map[int]common.MenuItem{
			CmdSearch: {
				Shortcut: "F2",
				Keys:     []tcell.Key{tcell.KeyF2},
				MenuKey:  "search",
				Text:     "Search",
				Positive: false,
				IsDef:    true,
			},
			CmdList: {
				Shortcut: "F3",
				Keys:     []tcell.Key{tcell.KeyF3},
				MenuKey:  "list",
				Text:     "List",
				Positive: false,
				IsDef:    true,
			},
			CmdRefresh: {
				Shortcut: "F5",
				Keys:     []tcell.Key{tcell.KeyF5},
				MenuKey:  "",
				Text:     "Refresh",
				Positive: true,
				IsDef:    true,
			},
			CmdSave: {
				Shortcut: "F6",
				Keys:     []tcell.Key{tcell.KeyF6},
				MenuKey:  "",
				Text:     "Save",
				Positive: true,
				IsDef:    true,
			},
			CmdCancel: {
				Shortcut: "F8",
				Keys:     []tcell.Key{tcell.KeyF8},
				MenuKey:  "",
				Text:     "Cancel",
				Positive: false,
				IsDef:    true,
			},
			CmdQuit: {
				Shortcut: "Esc",
				Keys:     []tcell.Key{tcell.KeyEsc},
				MenuKey:  "",
				Text:     "Quit",
				Positive: false,
				IsDef:    true,
			},
		}
	}

	tests := []struct {
		name          string
		input         string
		wantErrSubstr string
	}{
		{
			name:          "invalid keyboard shortcut key",
			input:         `keyboardShortcuts: { invalidKey: Search }`,
			wantErrSubstr: "invalidKey",
		},
		{
			name:          "invalid keyboard command",
			input:         `keyboardShortcuts: { F10: InvalidCommand }`,
			wantErrSubstr: "InvalidCommand",
		},
		{
			name:          "multiple invalid keyboard commands",
			input:         `keyboardShortcuts: { F10: Command1, F11: Command2 }`,
			wantErrSubstr: "Command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original state
			originalMenuItems := common.MenuItems
			originalShortcuts := common.Shortcuts

			// Setup test common.MenuItems
			common.MenuItems = createTestMenuItems()
			common.Shortcuts = make(map[tcell.Key]int)

			ioReader := strings.NewReader(tt.input)

			var actualErr error
			actualErr = loadConf(func(menuKey, text, shortcut string, isPositive bool) {}, ioReader)

			// Verify error occurred
			assert.Error(t, actualErr, "expected error for test case: %s", tt.name)

			// Verify error message contains expected substring
			assert.Contains(t, actualErr.Error(), tt.wantErrSubstr,
				"error message should contain '%s' for test case: %s", tt.wantErrSubstr, tt.name)

			// Restore original state
			common.MenuItems = originalMenuItems
			common.Shortcuts = originalShortcuts
		})
	}
}

func TestLoadConfShortcutsPopulated(t *testing.T) {
	// Test that shortcuts are correctly populated in the common.Shortcuts map
	originalMenuItems := common.MenuItems
	originalShortcuts := common.Shortcuts

	common.MenuItems = map[int]common.MenuItem{
		CmdSearch: {
			Shortcut: "F2",
			Keys:     []tcell.Key{tcell.KeyF2},
			MenuKey:  "search",
			Text:     "Search",
			Positive: false,
			IsDef:    true,
		},
		CmdList: {
			Shortcut: "F3",
			Keys:     []tcell.Key{tcell.KeyF3},
			MenuKey:  "list",
			Text:     "List",
			Positive: false,
			IsDef:    true,
		},
		CmdRefresh: {
			Shortcut: "F5",
			Keys:     []tcell.Key{tcell.KeyF5},
			MenuKey:  "",
			Text:     "Refresh",
			Positive: true,
			IsDef:    true,
		},
		CmdSave: {
			Shortcut: "F6",
			Keys:     []tcell.Key{tcell.KeyF6},
			MenuKey:  "",
			Text:     "Save",
			Positive: true,
			IsDef:    true,
		},
		CmdCancel: {
			Shortcut: "F8",
			Keys:     []tcell.Key{tcell.KeyF8},
			MenuKey:  "",
			Text:     "Cancel",
			Positive: false,
			IsDef:    true,
		},
		CmdQuit: {
			Shortcut: "Esc",
			Keys:     []tcell.Key{tcell.KeyEsc},
			MenuKey:  "",
			Text:     "Quit",
			Positive: false,
			IsDef:    true,
		},
	}
	common.Shortcuts = make(map[tcell.Key]int)

	err := loadConf(func(menuKey, text, shortcut string, isPositive bool) {}, nil)
	assert.NoError(t, err)

	// Verify all default shortcuts are in the map
	assert.Equal(t, CmdSearch, common.Shortcuts[tcell.KeyF2], "F2 should map to CmdSearch")
	assert.Equal(t, CmdList, common.Shortcuts[tcell.KeyF3], "F3 should map to CmdList")
	assert.Equal(t, CmdRefresh, common.Shortcuts[tcell.KeyF5], "F5 should map to CmdRefresh")
	assert.Equal(t, CmdSave, common.Shortcuts[tcell.KeyF6], "F6 should map to CmdSave")
	assert.Equal(t, CmdCancel, common.Shortcuts[tcell.KeyF8], "F8 should map to CmdCancel")
	assert.Equal(t, CmdQuit, common.Shortcuts[tcell.KeyEsc], "Esc should map to CmdQuit")

	// Restore original state
	common.MenuItems = originalMenuItems
	common.Shortcuts = originalShortcuts
}
