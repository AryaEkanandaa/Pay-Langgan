package revenue

import (
	"fmt"
	"pay-langgan/internal/models"
	"time"
)

func ParseDashboardRange(from, to, granularity string, location *time.Location, now time.Time) (time.Time, time.Time, error) {
	start, err := time.ParseInLocation("2006-01-02", from, location)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("tanggal Dari tidak valid")
	}
	last, err := time.ParseInLocation("2006-01-02", to, location)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("tanggal Sampai tidak valid")
	}
	today, _ := time.ParseInLocation("2006-01-02", now.In(location).Format("2006-01-02"), location)
	if start.After(last) {
		return time.Time{}, time.Time{}, fmt.Errorf("Dari tidak boleh setelah Sampai")
	}
	if last.After(today) {
		return time.Time{}, time.Time{}, fmt.Errorf("tanggal tidak boleh di masa depan")
	}
	minimum := last.AddDate(0, -24, 0)
	// Clamp leap day to February's last day, matching the frontend calendar logic.
	if minimum.Month() != last.Month() {
		minimum = time.Date(last.Year()-2, last.Month()+1, 0, 0, 0, 0, 0, location)
	}
	if start.Before(minimum) {
		return time.Time{}, time.Time{}, fmt.Errorf("rentang maksimum 24 bulan")
	}
	if granularity != "day" && granularity != "week" && granularity != "month" {
		return time.Time{}, time.Time{}, fmt.Errorf("granularitas tidak valid")
	}
	return start, last.AddDate(0, 0, 1), nil
}

func (s *DashboardService) RangeLocation(businessID string) (*time.Location, error) {
	zone, err := s.repo.BusinessTimezone(businessID)
	if err != nil {
		return nil, err
	}
	location, err := time.LoadLocation(zone)
	if err != nil {
		return time.LoadLocation("Asia/Makassar")
	}
	return location, nil
}

func (s *DashboardService) GetRangeSummary(businessID, from, to, granularity string, location *time.Location, role models.Role) (*models.DashboardSummary, error) {
	start, end, err := ParseDashboardRange(from, to, granularity, location, time.Now())
	if err != nil {
		return nil, err
	}
	summary, err := s.GetSummary(businessID, role)
	if err != nil {
		return nil, err
	}
	series, err := s.repo.RangeSeries(businessID, start, end, location.String(), granularity, role == models.RoleSales)
	if err != nil {
		return nil, err
	}
	plans := []models.PlanRevenuePoint{}
	if role != models.RoleSales {
		plans, err = s.repo.RangePlanRevenue(businessID, start, end)
		if err != nil {
			return nil, err
		}
	}
	summary.RangeStats = &models.DashboardRangeStats{From: from, To: to, Timezone: location.String(), Granularity: granularity, Series: series, RevenueByPlan: plans}
	return summary, nil
}
