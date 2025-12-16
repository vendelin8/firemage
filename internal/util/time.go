package util

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/vendelin8/firemage/internal/common"
	"github.com/vendelin8/firemage/internal/lang"
)

// timedButtonsMap stores parsed time values for each TimedButtons entry.
// Key is the label from TimedButtons, value is [years, months, days].
var timedButtonsMap map[string][]int

var (
	ErrEmpty = errors.New(lang.ErrEmptyTime)
	ErrFmt   = errors.New(lang.ErrTimeFmt)
)

type ErrInvalidUnit string

func NewErrInvalidUnit(in string) error {
	e := ErrInvalidUnit(in)
	return &e
}

func (e *ErrInvalidUnit) Error() string {
	str := (*string)(e)
	return fmt.Sprintf(lang.ErrTimeUnit, *str)
}

// InitializeTimedButtonsMap parses all TimedButtons entries and returns a map of [year, month, day] triplets.
func InitializeTimedButtonsMap() error {
	timedButtonsMap = make(map[string][]int, len(common.TimedButtons))
	for _, items := range common.TimedButtons {
		if len(items) != 2 {
			return fmt.Errorf("%s: %v", lang.ErrWrongTimeBtns, items)
		}

		label, timeStr := items[0], items[1]
		years, months, days, err := parseTimedFormat(timeStr)
		if err != nil {
			return fmt.Errorf("%s '%s': '%s' - %w", lang.ErrWrongTimeBtns, label, timeStr, err)
		}
		timedButtonsMap[label] = []int{years, months, days}
	}
	return nil
}

// parseTimedFormat parses a human-readable time format string and returns the years, months, and days.
// Supported formats:
//   - "1d" for 1 day
//   - "1w" for 1 week (7 days)
//   - "1m" for 1 month
//   - "1y" for 1 year
func parseTimedFormat(timeStr string) (years, months, days int, err error) {
	timeStr = strings.TrimSpace(timeStr)
	if timeStr == "" {
		return 0, 0, 0, ErrEmpty
	}

	// Extract the numeric part and the unit
	i := 0
	for i < len(timeStr) && timeStr[i] >= '0' && timeStr[i] <= '9' {
		i++
	}

	if i == 0 || i == len(timeStr) {
		return 0, 0, 0, ErrFmt
	}

	numStr := timeStr[:i]
	unit := strings.ToLower(timeStr[i:])

	num, parseErr := strconv.Atoi(numStr)
	if parseErr != nil {
		// this shouldn't happen, as it was tested to be between 0-9 inclusive above
		return 0, 0, 0, fmt.Errorf("invalid number: %s", numStr)
	}

	switch unit {
	case "d":
		return 0, 0, num, nil
	case "w":
		return 0, 0, num * 7, nil
	case "m":
		return 0, num, 0, nil
	case "y":
		return num, 0, 0, nil
	default:
		return 0, 0, 0, NewErrInvalidUnit(unit)
	}
}

// AddTimedDate adds a human-readable time duration to a given date and returns the modified date.
// Uses precomputed values from timedButtonsMap for efficient lookups.
func AddTimedDate(date time.Time, label string) time.Time {
	triplet := timedButtonsMap[label]
	years, months, days := triplet[0], triplet[1], triplet[2]
	return date.AddDate(years, months, days)
}
