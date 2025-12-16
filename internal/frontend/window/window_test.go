package window

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vendelin8/firemage/internal/common"
	"github.com/vendelin8/firemage/internal/global"
	"github.com/vendelin8/firemage/internal/lang"
	"github.com/vendelin8/firemage/internal/mock"
	testutil "github.com/vendelin8/firemage/internal/util/test"
	"go.uber.org/mock/gomock"
)

func TestHasPopup(t *testing.T) {
	tests := []struct {
		name        string
		activePopup string
		want        bool
	}{
		{
			name:        "no active popup",
			activePopup: "",
			want:        false,
		},
		{
			name:        "confirm popup active",
			activePopup: lang.PopupConfirm,
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.activePopup != "" {
				ActivePopups = append(ActivePopups, tt.activePopup)
			}
			result := HasPopup()
			assert.Equal(t, tt.want, result)
			ActivePopups = []string{}
		})
	}
}

func TestShowMsg(t *testing.T) {
	t.Run("shows message and pushes popup", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockFe := mock.NewMockFeIf(ctrl)
		common.Fe = mockFe

		msgText := "Test message"
		mockFe.EXPECT().ShowMsg(msgText).Times(1)

		ShowWarn(msgText)

		assert.Contains(t, ActivePopups, lang.PopupWarn)
		ActivePopups = []string{}
	})

	t.Run("shows empty message", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockFe := mock.NewMockFeIf(ctrl)
		common.Fe = mockFe

		mockFe.EXPECT().ShowMsg().Times(1)

		ShowWarn()

		assert.Contains(t, ActivePopups, lang.PopupWarn)
		ActivePopups = []string{}
	})
}

func TestShowWarningOnce(t *testing.T) {
	tests := []struct {
		name        string
		warningID   int
		warningText string
		shouldShow  bool
		description string
	}{
		{
			name:        "shows warning when configured",
			warningID:   1,
			warningText: "Test warning",
			shouldShow:  true,
			description: "configured warning should be shown",
		},
		{
			name:        "ignores unknown warning",
			warningID:   999,
			warningText: "",
			shouldShow:  false,
			description: "unconfigured warning should not be shown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFe := mock.NewMockFeIf(ctrl)
			common.Fe = mockFe

			// Setup warnings
			lang.Warns = map[int]string{
				1: "Test warning",
				2: "Another warning",
			}

			if tt.shouldShow {
				mockFe.EXPECT().ShowMsg(tt.warningText).Times(1)
			}

			ShowWarningOnce(tt.warningID)

			// If it was shown, it should be deleted from Warns
			if tt.shouldShow {
				_, exists := lang.Warns[tt.warningID]
				assert.False(t, exists, "warning should be deleted after showing")
			}

			// Reset
			lang.Warns = make(map[int]string)
			ActivePopups = []string{}
		})
	}
}

func TestShowConfirm(t *testing.T) {
	t.Run("shows confirm dialog and pushes popup", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockFe := mock.NewMockFeIf(ctrl)
		common.Fe = mockFe

		msgText := "Are you sure?"
		okFunc := func() {}
		cancelFunc := func() {}

		mockFe.EXPECT().ShowConfirm(gomock.Any(), gomock.Any(), msgText).Times(1)

		ShowConfirm(okFunc, cancelFunc, msgText)

		assert.Contains(t, ActivePopups, lang.PopupConfirm)
		ActivePopups = []string{}
	})

	t.Run("shows confirm with nil cancel function", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockFe := mock.NewMockFeIf(ctrl)
		common.Fe = mockFe

		msgText := "Confirm action?"
		okFunc := func() {}

		mockFe.EXPECT().ShowConfirm(gomock.Any(), nil, msgText).Times(1)

		ShowConfirm(okFunc, nil, msgText)

		assert.Contains(t, ActivePopups, lang.PopupConfirm)
		ActivePopups = []string{}
	})
}

func TestConfirmDoneFunc(t *testing.T) {
	tests := []struct {
		name         string
		buttonLabel  string
		okCalled     bool
		cancelCalled bool
		description  string
	}{
		{
			name:         "yes button calls ok function",
			buttonLabel:  lang.SYes,
			okCalled:     true,
			cancelCalled: false,
			description:  "clicking yes should call ok function",
		},
		{
			name:         "other button calls cancel function",
			buttonLabel:  lang.SNo,
			okCalled:     false,
			cancelCalled: true,
			description:  "clicking no should call cancel function",
		},
		{
			name:         "empty label calls cancel function",
			buttonLabel:  "",
			okCalled:     false,
			cancelCalled: true,
			description:  "empty label should call cancel function",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFe := mock.NewMockFeIf(ctrl)
			common.Fe = mockFe

			okCalled := false
			okFunc := func() {
				okCalled = true
			}

			cancelCalled := false
			cancelFunc := func() {
				cancelCalled = true
			}

			mockFe.EXPECT().HidePopup(lang.PopupConfirm).Times(1)

			ActivePopups = []string{lang.PopupConfirm}
			doneFunc := ConfirmDoneFunc(okFunc, cancelFunc)
			doneFunc(0, tt.buttonLabel)

			assert.Equal(t, tt.okCalled, okCalled, tt.description)
			assert.Equal(t, tt.cancelCalled, cancelCalled, tt.description)
			assert.NotContains(t, ActivePopups, lang.PopupConfirm, "popup should be hidden")
		})
	}
}

func TestConfirmDoneFuncWithoutCancelFunc(t *testing.T) {
	t.Run("cancel function can be nil", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockFe := mock.NewMockFeIf(ctrl)
		common.Fe = mockFe

		okCalled := false
		okFunc := func() {
			okCalled = true
		}

		mockFe.EXPECT().HidePopup(lang.PopupConfirm).Times(1)

		ActivePopups = []string{lang.PopupConfirm}
		doneFunc := ConfirmDoneFunc(okFunc, nil)
		doneFunc(0, lang.SNo)

		assert.False(t, okCalled, "ok function should not be called")
		assert.NotContains(t, ActivePopups, lang.PopupConfirm, "popup should be hidden")
	})
}

func TestHidePopup(t *testing.T) {
	t.Run("hides popup and removes from active popups", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockFe := mock.NewMockFeIf(ctrl)
		common.Fe = mockFe

		ActivePopups = []string{lang.PopupMsg}
		mockFe.EXPECT().HidePopup(lang.PopupMsg).Times(1)

		HidePopup(lang.PopupMsg)

		assert.NotContains(t, ActivePopups, lang.PopupMsg, "popup should be removed from active popups")
	})

	t.Run("hides popup when already empty", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockFe := mock.NewMockFeIf(ctrl)
		common.Fe = mockFe

		ActivePopups = []string{}
		mockFe.EXPECT().HidePopup(lang.PopupMsg).Times(1)

		HidePopup(lang.PopupMsg)

		assert.Empty(t, ActivePopups)
	})
}

func TestQuit(t *testing.T) {
	defer testutil.InitLog()

	tests := []struct {
		name         string
		setupActions map[string]map[string]any
	}{
		{
			name:         "quit with no pending actions",
			setupActions: map[string]map[string]any{},
		},
		{
			name:         "quit with pending actions shows confirmation",
			setupActions: map[string]map[string]any{"uid1": {"admin": true}},
		},
		{
			name:         "quit with multiple pending actions",
			setupActions: map[string]map[string]any{"uid1": {"admin": true}, "uid2": {"moderator": false}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFe := mock.NewMockFeIf(ctrl)
			common.Fe = mockFe

			global.Actions = testutil.BuildActionsMap(tt.setupActions)

			if len(tt.setupActions) == 0 {
				mockFe.EXPECT().Quit().Times(1)
			} else {
				mockFe.EXPECT().ShowConfirm(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
			}

			err := Quit()
			assert.NoError(t, err)

			global.Actions = map[string]common.ClaimsMap{}
			ActivePopups = []string{}
		})
	}
}
