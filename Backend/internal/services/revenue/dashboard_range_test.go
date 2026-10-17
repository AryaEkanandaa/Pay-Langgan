package revenue

import (
	"testing"
	"time"
)

func TestDashboardRangeBoundaries(t *testing.T) {
	location, err := time.LoadLocation("Asia/Makassar")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 10, 6, 9, 0, 0, 0, location)
	start, end, err := ParseDashboardRange("2026-10-06", "2026-10-06", "day", location, now)
	if err != nil {
		t.Fatal(err)
	}
	if start.UTC().Format(time.RFC3339) != "2026-10-05T16:00:00Z" || end.UTC().Format(time.RFC3339) != "2026-10-06T16:00:00Z" {
		t.Fatalf("local single-day bounds: %v %v", start, end)
	}
	for _, test := range []struct{ from, to, grain string }{
		{"2026-10-07", "2026-10-06", "day"}, {"2026-10-06", "2026-10-07", "day"},
		{"2024-10-05", "2026-10-06", "month"}, {"2026-02-30", "2026-10-06", "day"},
		{"", "2026-10-06", "day"}, {"2026-10-06", "2026-10-06", "invalid"},
	} {
		if _, _, err := ParseDashboardRange(test.from, test.to, test.grain, location, now); err == nil {
			t.Errorf("accepted invalid range: %+v", test)
		}
	}
	leapDay := time.Date(2028, 2, 29, 9, 0, 0, 0, location)
	if _, _, err := ParseDashboardRange("2026-02-28", "2028-02-29", "month", location, leapDay); err != nil {
		t.Fatalf("clamped 24-month boundary: %v", err)
	}
}
