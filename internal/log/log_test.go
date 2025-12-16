package log

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMust(t *testing.T) {
	tests := []struct {
		name        string
		description string
		err         error
		wantPanic   bool
		wantMsg     string
	}{
		{
			name:        "no error should not panic",
			description: "initialization",
			err:         nil,
			wantPanic:   false,
		},
		{
			name:        "error should panic with description and error message",
			description: "failed to load config",
			err:         errors.New("config file not found"),
			wantPanic:   true,
			wantMsg:     "failed to load config: config file not found",
		},
		{
			name:        "error with empty description",
			description: "",
			err:         errors.New("some error"),
			wantPanic:   true,
			wantMsg:     ": some error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				defer func() {
					r := recover()
					assert.NotNil(t, r)
					if r != nil {
						assert.Contains(t, r.(error).Error(), tt.wantMsg)
					}
				}()
				Must(tt.description, tt.err)
				assert.Fail(t, "expected panic but none occurred")
			} else {
				assert.NotPanics(t, func() {
					Must(tt.description, tt.err)
				})
			}
		})
	}
}
