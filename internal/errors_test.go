package internal

import (
	"testing"
)

func TestConfirm(t *testing.T) {
	defer setup()()

	var okPressed, cancelPressed bool
	okFunc := func() {
		okPressed = true
	}
	cancelFunc := func() {
		cancelPressed = true
	}
	confirmF := confirmDoneFunc(okFunc, cancelFunc)

	cases := []struct {
		buttonLabel string
		check       func()
	}{
		{
			buttonLabel: sYes,
			check: func() {
				if !okPressed {
					t.Error("okPressed should be true")
				}
				if cancelPressed {
					t.Error("cancelPressed should be false")
				}
			},
		},
		{
			buttonLabel: sNo,
			check: func() {
				if okPressed {
					t.Error("okPressed should be false")
				}
				if !cancelPressed {
					t.Error("cancelPressed should be true")
				}
			},
		},
	}

	for _, c := range cases {
		t.Run(c.buttonLabel, func(t *testing.T) {
			confirmF(0, c.buttonLabel)
			c.check()
			if hasPopup() {
				t.Fatal("should NOT have popup but does")
			}
			okPressed = false
			cancelPressed = false
		})
	}
}
