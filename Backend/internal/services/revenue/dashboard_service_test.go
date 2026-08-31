package revenue

import (
	"strconv"
	"testing"
	"time"

	revenuerepo "pay-langgan/internal/repositories/revenue"
)

func TestFillMonthlyPerformanceReturnsLastSevenMonths(t *testing.T) {
	now := time.Now()
	currentKey := now.Format("2006-01")
	previousKey := now.AddDate(0, -2, 0).Format("2006-01")

	points := fillMonthlyPerformance([]revenuerepo.MonthlyRevenuePoint{
		{Month: currentKey, Value: 1500000},
		{Month: previousKey, Value: 750000},
	})

	if len(points) != 7 {
		t.Fatalf("expected 7 monthly points, got %d", len(points))
	}
	if points[6].Value != 1500000 {
		t.Errorf("expected current month value 1500000, got %v", points[6].Value)
	}
	if points[4].Value != 750000 {
		t.Errorf("expected value from two months ago 750000, got %v", points[4].Value)
	}
}

func TestFillYearlyBreakdownReturnsCurrentAndPreviousFourYears(t *testing.T) {
	currentYear := time.Now().Year()
	rows := []revenuerepo.YearlyRevenuePoint{
		{Year: strconv.Itoa(currentYear), Income: 24000000},
		{Year: strconv.Itoa(currentYear - 2), Income: 12000000},
	}

	points := fillYearlyBreakdown(rows)

	if len(points) != 5 {
		t.Fatalf("expected 5 yearly points, got %d", len(points))
	}
	if points[4].Income != 24000000 {
		t.Errorf("expected current year income 24000000, got %v", points[4].Income)
	}
	if points[2].Income != 12000000 {
		t.Errorf("expected income from two years ago 12000000, got %v", points[2].Income)
	}
	if points[4].Spending != 0 {
		t.Errorf("expected spending to be 0 until expense tracking exists, got %v", points[4].Spending)
	}
}
