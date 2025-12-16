package window

import (
	"errors"
	"strings"

	"github.com/vendelin8/firemage/internal/common"
	"github.com/vendelin8/firemage/internal/log"
)

const (
	msgError = iota
	msgConfirm
	msgWarn
)

var (
	// b is a global buffer for creating error messages. Because the app is single threaded, all
	// functionality is sequential, it's safe this way. After every operation it is checked.
	// If there's something to show, show it in a modal popup, and clear the buffer.
	b = &strings.Builder{}

	// bs is a global map of buffers for creating error messages. There may be multiple
	// purposes of messages to collect, this way provides this.
	bs = map[int]*strings.Builder{msgError: b}

	// bStack keeps track of recently used buffers, so it can continue using the last one.
	bStack = []int{msgError}
)

// AppendListError creates an error through the error buffer with a given list of input strings.
func AppendListError(msg string, inputs []string, msgs ...string) bool {
	// Early return in case of no error
	if len(inputs) == 0 {
		return false
	}

	optionalNewline()

	for i, input := range inputs {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(input)
	}

	for _, m := range msgs {
		b.WriteByte('\n')
		b.WriteString(m)
	}

	return true
}

func doGetError() string {
	s := b.String()
	log.Lgr.Info(s)
	b.Reset()

	return s
}

func GetError() error {
	if b.Len() == 0 {
		return nil
	}

	return errors.New(doGetError())
}

// WriteErrorMap creates an error through the error buffer with a given map of input strings.
/*func WriteErrorMap(msg string, inputs map[string]struct{}, msgs ...string) error {
	// Early return in case of no error
	if len(inputs) == 0 {
		return nil
	}

	l := make([]string, 0, len(inputs))
	for in := range inputs {
		l = append(l, in)
	}

	return AppendListError(msg, l, msgs...)
}*/

// GetErrorStr returns current error message if the buffer is not empty as a string.
func GetErrorStr() string {
	// Early return in case of no error
	if b.Len() == 0 {
		return ""
	}

	return doGetError()
}

func optionalNewline() {
	if b.Len() > 0 {
		b.WriteByte('\n')
	}
}

func UseConfirm() {
	pushBuffer(msgConfirm)
}

func UseWarn() {
	pushBuffer(msgWarn)
}

// pushBuffer sets buffer to use.
func pushBuffer(id int) {
	if bStack[len(bStack)-1] == id {
		return
	}

	bStack = append(bStack, id)

	var ok bool
	b, ok = bs[id]
	if ok {
		return
	}

	b = &strings.Builder{}
	bs[id] = b
}

// PopBuffer returns content of the current buffer, and goes back to the previously used one.
func PopBuffer() string {
	result := GetErrorStr()

	bStack = bStack[:len(bStack)-1]
	b = bs[bStack[len(bStack)-1]]

	return result
}

// WriteErrorStr adds a new error line to the error buffer.
func WriteErrorStr(msg string) {
	optionalNewline()
	b.WriteString(msg)
}

// ShowErrorBuffer displays any buffered errors in a popup and clears the buffer.
// Call this at the end of any functionality that uses WriteErrorStr to ensure the buffer is cleared.
func ShowErrorBuffer(err error) {
	if err == nil && b.Len() == 0 {
		return
	}

	if err != nil {
		WriteErrorStr(err.Error())
	}

	common.Fe.ShowMsg(GetErrorStr())
}
