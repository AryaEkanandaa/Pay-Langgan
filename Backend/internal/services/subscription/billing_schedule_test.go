package subscription

import (
	"testing"
	"time"
)

func TestInitialBillingSchedule(t *testing.T) {
	tests := []struct {
		name, start, cycle, next string
		trial                    int
	}{
		{"monthly", "2026-09-09", "monthly", "2026-10-09", 0},
		{"yearly", "2026-09-09", "yearly", "2027-09-09", 0},
		{"month end", "2026-01-31", "monthly", "2026-02-28", 0},
		{"leap month", "2024-01-31", "monthly", "2024-02-29", 0},
		{"leap year anniversary", "2024-02-29", "yearly", "2025-02-28", 0},
		{"year boundary", "2026-12-31", "monthly", "2027-01-31", 0},
		{"trial crosses month", "2026-01-25", "monthly", "2026-02-08", 14},
		{"yearly trial", "2026-12-25", "yearly", "2027-01-08", 14},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, _ := time.Parse(time.RFC3339Nano, tt.start+"T08:15:30.123Z")
			want, _ := time.Parse(time.RFC3339Nano, tt.next+"T08:15:30.123Z")
			got, err := initialBillingSchedule(start, tt.cycle, tt.trial)
			if err != nil {
				t.Fatal(err)
			}
			if !got.nextBillingDate.Equal(want) {
				t.Fatalf("next billing = %v, want %v", got.nextBillingDate, want)
			}
			if tt.trial > 0 {
				if got.status != "trial" || got.trialEndsAt == nil || !got.trialEndsAt.Equal(want) || got.invoiceDueDate != nil {
					t.Fatalf("trial must defer invoice until trial end: %+v", got)
				}
			} else if got.status != "pending" || got.trialEndsAt != nil || got.invoiceDueDate == nil || !got.invoiceDueDate.Equal(start) {
				t.Fatalf("non-trial must invoice now and schedule a later period: %+v", got)
			}
		})
	}
}

func TestInitialBillingScheduleRejectsInvalidPlan(t *testing.T) {
	for _, tt := range []struct {
		cycle string
		trial int
	}{{"weekly", 0}, {"", 14}, {"monthly", -1}} {
		if _, err := initialBillingSchedule(time.Now(), tt.cycle, tt.trial); err == nil {
			t.Fatalf("expected invalid plan error for %+v", tt)
		}
	}
}
