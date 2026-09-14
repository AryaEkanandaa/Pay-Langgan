package billing

import (
	"fmt"
	"time"
)

// NextPeriod returns the first anchored boundary strictly after start.
// Anchoring to the original day prevents February from shifting future renewals.
func NextPeriod(anchor, start time.Time, cycle string) (time.Time, error) {
	months := 1
	if cycle == "yearly" {
		months = 12
	} else if cycle != "monthly" {
		return time.Time{}, fmt.Errorf("unsupported billing cycle %q", cycle)
	}
	if anchor.IsZero() || start.Before(anchor) {
		return time.Time{}, fmt.Errorf("invalid billing anchor")
	}
	distance := (start.Year()-anchor.Year())*12 + int(start.Month()-anchor.Month())
	step := distance / months
	for {
		target := time.Date(anchor.Year(), anchor.Month()+time.Month(step*months), 1, anchor.Hour(), anchor.Minute(), anchor.Second(), anchor.Nanosecond(), anchor.Location())
		day := anchor.Day()
		if last := target.AddDate(0, 1, -1).Day(); day > last {
			day = last
		}
		target = target.AddDate(0, 0, day-1)
		if target.After(start) {
			return target, nil
		}
		step++
	}
}
