package billing

import (
	"testing"
	"time"
)

func TestNextPeriodKeepsAnchor(t *testing.T) {
	for _, tt := range []struct{ anchor, start, cycle, want string }{
		{"2026-01-31", "2026-02-28", "monthly", "2026-03-31"},
		{"2026-01-30", "2026-02-28", "monthly", "2026-03-30"},
		{"2024-02-29", "2027-02-28", "yearly", "2028-02-29"},
		{"2026-12-31", "2026-12-31", "monthly", "2027-01-31"},
	} {
		anchor, _ := time.Parse("2006-01-02", tt.anchor)
		start, _ := time.Parse("2006-01-02", tt.start)
		got, err := NextPeriod(anchor, start, tt.cycle)
		if err != nil || got.Format("2006-01-02") != tt.want {
			t.Fatalf("%+v: %v, %v", tt, got, err)
		}
	}
}

func TestPriceCharges(t *testing.T) {
	for _, tt := range []struct {
		name    string
		charges []charge
		total   int64
		invalid bool
	}{
		{"fractional percentage floors once", []charge{{Kind: "plan", Name: "Plan", Quantity: 1, Cents: 10001, Cycle: "monthly"}, {Kind: "percentage", Name: "10%", Quantity: 1, Cents: 1000}}, 9001, false},
		{"add-on quantity", []charge{{Kind: "plan", Quantity: 1, Cents: 10000, Cycle: "monthly"}, {Kind: "add_on", Quantity: 3, Cents: 2500, Cycle: "monthly"}}, 17500, false},
		{"stacked coupons", []charge{{Kind: "plan", Quantity: 1, Cents: 10000, Cycle: "monthly"}, {Kind: "percentage", Quantity: 1, Cents: 1000}, {Kind: "fixed", Quantity: 1, Cents: 1000}}, 8000, false},
		{"discount capped", []charge{{Kind: "plan", Quantity: 1, Cents: 10000, Cycle: "monthly"}, {Kind: "fixed", Quantity: 1, Cents: 15000}}, 0, false},
		{"mixed cadence rejected", []charge{{Kind: "add_on", Quantity: 1, Cents: 100, Cycle: "yearly"}}, 0, true},
		{"overflow rejected", []charge{{Kind: "plan", Quantity: 100, Cents: maxInvoiceCents, Cycle: "monthly"}}, 0, true},
		{"invalid percentage", []charge{{Kind: "percentage", Quantity: 1, Cents: 10001}}, 0, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			lines, total, err := priceCharges(tt.charges, "monthly")
			if tt.invalid {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil || total != tt.total {
				t.Fatalf("total=%d error=%v", total, err)
			}
			var sum int64
			for _, line := range lines {
				sum += line
			}
			if sum != total {
				t.Fatal("snapshot does not reconcile")
			}
		})
	}
}
