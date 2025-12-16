package window

import (
	"testing"

	"github.com/stretchr/testify/assert"
	testutil "github.com/vendelin8/firemage/internal/util/test"
)

func TestGetListError(t *testing.T) {
	cleanup := testutil.InitLog()
	defer cleanup()
	tests := []struct {
		name        string
		msg         string
		inputs      []string
		msgs        []string
		expectError bool
		wantError   string
	}{
		{
			name:        "shows error with inputs",
			msg:         "Error occurred: %s",
			inputs:      []string{"input1", "input2"},
			msgs:        []string{},
			expectError: true,
			wantError:   "input1, input2",
		},
		{
			name:        "no error when inputs empty",
			msg:         "Error occurred: %s",
			inputs:      []string{},
			msgs:        []string{},
			expectError: false,
		},
		{
			name:        "shows error with inputs and messages",
			msg:         "Error occurred: %s",
			inputs:      []string{"input1"},
			msgs:        []string{"detail1"},
			expectError: true,
			wantError:   "input1\ndetail1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AppendListError(tt.msg, tt.inputs, tt.msgs...)
			err := GetError()

			if tt.expectError {
				assert.NotNil(t, err)
				assert.Equal(t, tt.wantError, err.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
