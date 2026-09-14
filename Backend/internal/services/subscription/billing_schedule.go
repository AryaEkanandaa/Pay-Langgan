package subscription

import (
	"fmt"
	"time"

	"pay-langgan/internal/services/billing"
)

type billingSchedule struct {
	status          string
	trialEndsAt     *time.Time
	nextBillingDate time.Time
	invoiceDueDate  *time.Time
}

// initialBillingSchedule separates the first invoice's due date from the next
// unbilled period. A trial does not issue an invoice before its end.
func initialBillingSchedule(start time.Time, cycle string, trialDays int) (billingSchedule, error) {
	switch cycle {
	case "monthly", "yearly":
	default:
		return billingSchedule{}, fmt.Errorf("unsupported billing cycle %q", cycle)
	}
	if trialDays < 0 {
		return billingSchedule{}, fmt.Errorf("trial days must not be negative")
	}
	if trialDays > 0 {
		end := start.AddDate(0, 0, trialDays)
		return billingSchedule{status: "trial", trialEndsAt: &end, nextBillingDate: end}, nil
	}

	next, err := billing.NextPeriod(start, start, cycle)
	if err != nil {
		return billingSchedule{}, err
	}
	return billingSchedule{
		status:          "active",
		nextBillingDate: next,
		invoiceDueDate:  &start,
	}, nil
}
