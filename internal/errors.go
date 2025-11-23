package internal

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

var (
	// b is a global buffer for creating error messages. Because the app is single threaded, all
	// functionality is sequential, it's safe this way. After every operation it is checked.
	// If there's something to show, show it in a modal popup, and clear the buffer.
	b strings.Builder
)

// showMsg shows the given message to the user, based on current frontend.
func showMsg(m string) {
	activePopup = popupMsg
	fe.showMsg(m)
}

// showWarningOnce shows a warning if it wasn't shown yet in this session.
func showWarningOnce(w int) {
	ws, ok := warns[w]
	if !ok {
		return
	}
	delete(warns, w)
	showMsg(ws)
}

// showConfirm shows a confirm dialog with a text, and callback functions for OK and Cancel.
func showConfirm(m string, okFunc, cancelFunc func()) {
	activePopup = popupCnfrm
	fe.showConfirm(m, okFunc, cancelFunc)
}

// confirmDoneFunc returns a function that will call the given "OK" or "cancel" function
// based on selected GUI button in a confirm popup. Cancel may be nil.
func confirmDoneFunc(okFunc, cancelFunc func()) func(int, string) {
	return func(buttonIndex int, buttonLabel string) {
		hidePopup()
		if buttonLabel == sYes {
			okFunc()
			return
		}
		if cancelFunc != nil {
			cancelFunc()
		}
	}
}

// hidePopup hides the current popup window.
func hidePopup() {
	fe.hidePopup()
	activePopup = ""
}

// writeErrorStr adds a new error line to the error buffer.
func writeErrorStr(msg string) {
	if b.Len() > 0 {
		b.WriteByte('\n')
	}
	b.WriteString(msg)
}

// writeErrorListGet creates an error through the error buffer with a given list of input strings.
func writeErrorListGet(msg string, inputs []string, msgs ...string) error {
	writeErrorList(msg, inputs, msgs...)
	return getErrors()
}

// writeErrorListShow shows an error through the error buffer with a given list of input strings.
func writeErrorListShow(msg string, inputs []string, msgs ...string) {
	writeErrorList(msg, inputs, msgs...)

	if errStr := getErrorStr(); len(errStr) > 0 {
		showMsg(errStr)
	}
}

// writeErrorList writes an error to the error buffer with a given list of input strings.
func writeErrorList(msg string, inputs []string, msgs ...string) {
	if len(inputs) == 0 {
		return
	}

	sort.Strings(inputs)
	writeErrorStr(fmt.Sprintf(msg, strings.Join(inputs, ", ")))

	for _, msg = range msgs {
		writeErrorStr(msg)
	}
}

// writeErrorMap creates an error through the error buffer with a given map of input strings.
func writeErrorMap(msg string, inputs map[string]struct{}, msgs ...string) error {
	if len(inputs) == 0 {
		return nil
	}

	l := make([]string, 0, len(inputs))
	for in := range inputs {
		l = append(l, in)
	}

	return writeErrorListGet(msg, l, msgs...)
}

// writeErrorIf adds a new error line if the given error is not empty. Returns if it was OK.
// func writeErrorIf(err error) bool {
// 	if err == nil {
// 		return true
// 	}
// 	writeErrorStr(err.Error())
// 	return false
// }

// getErrors creates current error message if the buffer is not empty as an error.
func getErrors() error {
	s := getErrorStr()
	if len(s) == 0 {
		return nil
	}

	return errors.New(s)
}

// getErrorStr returns current error message if the buffer is not empty as a string.
func getErrorStr() string {
	if b.Len() == 0 {
		return ""
	}

	s := b.String()
	lgr.Info(s)

	b.Reset()
	return s
}
