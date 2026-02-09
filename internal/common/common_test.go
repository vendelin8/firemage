package common

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClaimDiffers(t *testing.T) {
	t.Parallel()

	// Helper function to create a time pointer
	timePtr := func(year int, month time.Month, day int, hour ...int) *time.Time {
		if len(hour) == 0 {
			hour = []int{0}
		}

		t := time.Date(year, month, day, hour[0], 0, 0, 0, time.UTC)

		return &t
	}

	tests := []struct {
		name     string
		claim1   *Claim
		claim2   *Claim
		wantDiff bool
	}{
		{
			name:     "both zero claims (no date, checked false)",
			claim1:   &Claim{Checked: false, Date: nil},
			claim2:   &Claim{Checked: false, Date: nil},
			wantDiff: false,
		},
		{
			name:     "both claims with same checked value (true)",
			claim1:   &Claim{Checked: true, Date: nil},
			claim2:   &Claim{Checked: true, Date: nil},
			wantDiff: false,
		},
		{
			name:     "different checked values",
			claim1:   &Claim{Checked: true, Date: nil},
			claim2:   &Claim{Checked: false, Date: nil},
			wantDiff: true,
		},
		{
			name:     "one has date, other doesn't (checked same)",
			claim1:   &Claim{Checked: true, Date: timePtr(2024, 1, 15)},
			claim2:   &Claim{Checked: true, Date: nil},
			wantDiff: true,
		},
		{
			name:     "one has date, other doesn't (checked same, false)",
			claim1:   &Claim{Checked: false, Date: timePtr(2024, 1, 15)},
			claim2:   &Claim{Checked: false, Date: nil},
			wantDiff: true,
		},
		{
			name:     "both have same date",
			claim1:   &Claim{Checked: false, Date: timePtr(2024, 1, 15)},
			claim2:   &Claim{Checked: false, Date: timePtr(2024, 1, 15)},
			wantDiff: false,
		},
		{
			name:     "both have different dates (different days)",
			claim1:   &Claim{Checked: false, Date: timePtr(2024, 1, 15)},
			claim2:   &Claim{Checked: false, Date: timePtr(2024, 1, 16)},
			wantDiff: true,
		},
		{
			name: "same date but different times (should not differ when truncated to days)",
			claim1: &Claim{
				Checked: false,
				Date:    &time.Time{},
			},
			claim2: &Claim{
				Checked: false,
				Date:    &time.Time{},
			},
			wantDiff: false,
		},
		{
			name:     "dates differ by one day",
			claim1:   &Claim{Checked: false, Date: timePtr(2024, 1, 15)},
			claim2:   &Claim{Checked: false, Date: timePtr(2024, 1, 14)},
			wantDiff: true,
		},
		{
			name:     "both have checked true and same date",
			claim1:   &Claim{Checked: true, Date: timePtr(2024, 2, 1)},
			claim2:   &Claim{Checked: true, Date: timePtr(2024, 2, 1)},
			wantDiff: false,
		},
		{
			name:     "different checked with same date",
			claim1:   &Claim{Checked: true, Date: timePtr(2024, 1, 15)},
			claim2:   &Claim{Checked: false, Date: timePtr(2024, 1, 15)},
			wantDiff: true,
		},
		{
			name:     "far apart dates",
			claim1:   &Claim{Checked: false, Date: timePtr(2024, 1, 1)},
			claim2:   &Claim{Checked: false, Date: timePtr(2024, 12, 31)},
			wantDiff: true,
		},
		{
			name:     "start of day vs end of same day",
			claim1:   &Claim{Checked: false, Date: timePtr(2024, 1, 15, 0)},
			claim2:   &Claim{Checked: false, Date: timePtr(2024, 1, 15, 23)},
			wantDiff: false,
		},
		{
			name:     "both zero values",
			claim1:   &Claim{},
			claim2:   &Claim{},
			wantDiff: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.claim1.Differs(tt.claim2)
			assert.Equal(t, tt.wantDiff, result, "Differs returned unexpected result")
		})
	}
}
