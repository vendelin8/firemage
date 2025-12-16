package util

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vendelin8/firemage/internal/common"
	"github.com/vendelin8/firemage/internal/lang"
)

func TestAddTimedDate(t *testing.T) {
	baseDate := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	common.TimedButtons = [][]string{
		{"One month", "1m"},
		{"Three months", "3m"},
		{"One year", "1y"},
	}
	InitializeTimedButtonsMap()

	// Use actual TimedButtons entries for testing
	tests := []struct {
		name  string
		date  time.Time
		label string
		want  time.Time
	}{
		{
			name:  "one month",
			date:  baseDate,
			label: "One month",
			want:  time.Date(2025, 2, 1, 12, 0, 0, 0, time.UTC),
		},
		{
			name:  "three months",
			date:  baseDate,
			label: "Three months",
			want:  time.Date(2025, 4, 1, 12, 0, 0, 0, time.UTC),
		},
		{
			name:  "one year",
			date:  baseDate,
			label: "One year",
			want:  time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
		},
		{
			name:  "add month from Dec 31 to next year",
			date:  time.Date(2025, 12, 31, 12, 0, 0, 0, time.UTC),
			label: "One month",
			want:  time.Date(2026, 1, 31, 12, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AddTimedDate(tt.date, tt.label)
			require.Equal(t, got, tt.want)
		})
	}
}

func TestInitializeTimedButtonsMap(t *testing.T) {
	tests := []struct {
		name        string
		timedBtns   [][]string
		wantErr     error
		wantEntries map[string][]int
	}{
		{
			name: "valid entries",
			timedBtns: [][]string{
				{"One day", "1d"},
				{"One week", "1w"},
				{"One month", "1m"},
				{"One year", "1y"},
				{"Seven weeks", "7w"},
			},
			wantErr: nil,
			wantEntries: map[string][]int{
				"One day":     {0, 0, 1},
				"One week":    {0, 0, 7},
				"One month":   {0, 1, 0},
				"One year":    {1, 0, 0},
				"Seven weeks": {0, 0, 49},
			},
		},
		{
			name:        "empty map",
			timedBtns:   [][]string{},
			wantErr:     nil,
			wantEntries: map[string][]int{},
		},
		{
			name: "invalid format - no number",
			timedBtns: [][]string{
				{"Invalid", "d"},
			},
			wantErr: fmt.Errorf("%s 'Invalid': 'd' - %w", lang.ErrWrongTimeBtns, ErrFmt),
		},
		{
			name: "invalid format - no unit",
			timedBtns: [][]string{
				{"Invalid", "5"},
			},
			wantErr: fmt.Errorf("%s 'Invalid': '5' - %w", lang.ErrWrongTimeBtns, ErrFmt),
		},
		{
			name: "invalid unit",
			timedBtns: [][]string{
				{"Invalid", "5x"},
			},
			wantErr: fmt.Errorf("%s 'Invalid': '5x' - %w", lang.ErrWrongTimeBtns, NewErrInvalidUnit("x")),
		},
		{
			name: "empty string value",
			timedBtns: [][]string{
				{"Invalid", ""},
			},
			wantErr: fmt.Errorf("%s 'Invalid': '' - %w", lang.ErrWrongTimeBtns, ErrEmpty),
		},
		{
			name: "multiple entries with one invalid",
			timedBtns: [][]string{
				{"Valid", "1m"},
				{"Invalid", "xyz"},
			},
			wantErr: fmt.Errorf("%s 'Invalid': 'xyz' - %w", lang.ErrWrongTimeBtns, ErrFmt),
		},
		{
			name: "large numbers",
			timedBtns: [][]string{
				{"Large day", "365d"},
				{"Large month", "12m"},
				{"Large year", "10y"},
			},
			wantErr: nil,
			wantEntries: map[string][]int{
				"Large day":   {0, 0, 365},
				"Large month": {0, 12, 0},
				"Large year":  {10, 0, 0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			common.TimedButtons = tt.timedBtns
			err := InitializeTimedButtonsMap()

			if tt.wantErr == nil {
				require.NoError(t, err)
				require.Equal(t, timedButtonsMap, tt.wantEntries)
			} else {
				require.Equal(t, err.Error(), tt.wantErr.Error())
				require.Less(t, len(timedButtonsMap), len(tt.timedBtns))
			}
		})
	}
}

func TestParseTimedFormat(t *testing.T) {
	tests := []struct {
		name       string
		timeStr    string
		wantYears  int
		wantMonths int
		wantDays   int
		wantErr    error
	}{
		// Valid cases
		{"1 day", "1d", 0, 0, 1, nil},
		{"10 days", "10d", 0, 0, 10, nil},
		{"1 week", "1w", 0, 0, 7, nil},
		{"4 weeks", "4w", 0, 0, 28, nil},
		{"1 month", "1m", 0, 1, 0, nil},
		{"3 months", "3m", 0, 3, 0, nil},
		{"1 year", "1y", 1, 0, 0, nil},
		{"2 years", "2y", 2, 0, 0, nil},
		{"whitespace", "  1d  ", 0, 0, 1, nil},
		{"uppercase", "1D", 0, 0, 1, nil},

		// Error cases
		{"empty string", "", 0, 0, 0, ErrEmpty},
		{"no unit", "5", 0, 0, 0, ErrFmt},
		{"no number", "d", 0, 0, 0, ErrFmt},
		{"invalid unit", "5x", 0, 0, 0, NewErrInvalidUnit("x")},
		{"negative", "-5d", 0, 0, 0, ErrFmt},
		{"zero", "0d", 0, 0, 0, nil},
		{"invalid number", "abcd", 0, 0, 0, ErrFmt},
		{"whitespace only", "   ", 0, 0, 0, ErrEmpty},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			years, months, days, err := parseTimedFormat(tt.timeStr)

			require.Equal(t, err, tt.wantErr)
			require.Equal(t, years, tt.wantYears)
			require.Equal(t, months, tt.wantMonths)
			require.Equal(t, days, tt.wantDays)
		})
	}
}
