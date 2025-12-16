package window

import (
	"context"
	"fmt"
	"slices"

	"github.com/vendelin8/firemage/internal/common"
	"github.com/vendelin8/firemage/internal/global"
	"github.com/vendelin8/firemage/internal/lang"
	"github.com/vendelin8/firemage/internal/log"
	"go.uber.org/zap"
)

var ActivePopups []string

// ShowWarningOnce shows a warning if it wasn't shown yet in this session.
func ShowWarningOnce(w int) {
	ws, ok := lang.Warns[w]
	if !ok {
		return
	}
	delete(lang.Warns, w)
	common.Fe.ShowMsg(ws)
}

func PushPopup(popup string) {
	ActivePopups = append(ActivePopups, popup)
	log.Lgr.Debug("PushPopup end", zap.String("popup", popup), zap.Strings("ActivePopups", ActivePopups))
}

// ShowConfirm shows a confirm dialog with a text, and callback functions for OK and Cancel.
func ShowConfirm(onYes, onNo func(), ms ...string) {
	PushPopup(lang.PopupConfirm)
	common.Fe.ShowConfirm(onYes, onNo, ms...)
}

// ShowProgress shows a progress dialog with a text, and callback functions for OK and Cancel.
func ShowProgress(ctx context.Context, onCancel func(), ms ...string) {
	PushPopup(lang.PopupProgress)
	common.Fe.ShowProgress(ctx, onCancel, ms...)
}

// ShowWarn shows a warning dialog with a text.
func ShowWarn(ms ...string) {
	PushPopup(lang.PopupWarn)
	common.Fe.ShowMsg(ms...)
}

// ConfirmDoneFunc returns a function that will call the given "OK" or "cancel" function
// based on selected GUI button in a confirm popup. Cancel may be nil.
func ConfirmDoneFunc(okFunc, cancelFunc func()) func(int, string) {
	return func(buttonIndex int, buttonLabel string) {
		HidePopup(lang.PopupConfirm)
		if buttonLabel == lang.SYes {
			okFunc()
			return
		}
		if cancelFunc != nil {
			cancelFunc()
		}
	}
}

// HidePopup hides the current popup window.
func HidePopup(popup string) {
	common.Fe.HidePopup(popup)
	for i, p := range ActivePopups {
		if popup == p {
			ActivePopups = slices.Delete(ActivePopups, i, i+1)
			break
		}
	}
	log.Lgr.Debug("HidePopup end", zap.String("popup", popup), zap.Strings("ActivePopups", ActivePopups))
}

// HasPopup returns if there's an active popup window.
func HasPopup() bool {
	return len(ActivePopups) > 0
}

// Quit exists the application.
func Quit() error {
	if len(global.Actions) == 0 {
		common.Fe.Quit()
		return nil
	}

	common.Fe.ShowConfirm(func() { common.Fe.Quit() }, nil, fmt.Sprintf(lang.WarnUnsaved, len(global.Actions)))
	return nil
}
