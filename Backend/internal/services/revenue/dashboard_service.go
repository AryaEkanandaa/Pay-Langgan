package revenue

import (
	"strconv"
	"time"

	"pay-langgan/internal/models"
	revenuerepo "pay-langgan/internal/repositories/revenue"
)

type DashboardService struct {
	repo *revenuerepo.DashboardRepository
}

func NewDashboardService(repo *revenuerepo.DashboardRepository) *DashboardService {
	return &DashboardService{repo: repo}
}

func (s *DashboardService) GetSummary(businessID string, roles ...models.Role) (*models.DashboardSummary, error) {
	activeCustomers, err := s.repo.CountActiveCustomers(businessID)
	if err != nil {
		return nil, err
	}
	activeSubscriptions, err := s.repo.CountActiveSubscriptions(businessID)
	if err != nil {
		return nil, err
	}
	mrr, err := s.repo.CalculateMRR(businessID)
	if err != nil {
		return nil, err
	}
	dueInvoices, err := s.repo.CountDueInvoices(businessID)
	if err != nil {
		return nil, err
	}
	monthlyRows, err := s.repo.MonthlyRevenue(businessID)
	if err != nil {
		return nil, err
	}
	yearlyRows, err := s.repo.YearlyRevenue(businessID)
	if err != nil {
		return nil, err
	}
	financials, err := s.repo.Financials(businessID)
	if err != nil {
		return nil, err
	}
	planRevenue, err := s.repo.RevenueByPlan(businessID)
	if err != nil {
		return nil, err
	}

	summary := &models.DashboardSummary{
		ActiveCustomers:        activeCustomers,
		ActiveSubscriptions:    activeSubscriptions,
		MRR:                    mrr,
		DueInvoices:            dueInvoices,
		RevenueThisMonth:       financials.RevenueThisMonth,
		OutstandingAmount:      financials.OutstandingAmount,
		OverdueAmount:          financials.OverdueAmount,
		PaymentsSucceeded:      financials.PaymentsSucceeded,
		PaymentsFailed:         financials.PaymentsFailed,
		AutopaySucceeded:       financials.AutopaySucceeded,
		AutopayFailed:          financials.AutopayFailed,
		NewSubscriptions:       financials.NewSubscriptions,
		CancelledSubscriptions: financials.CancelledSubscriptions,
		RevenueByPlan:          planRevenue,
		MonthlyPerformance:     fillMonthlyPerformance(monthlyRows),
		YearlyBreakdown:        fillYearlyBreakdown(yearlyRows),
	}

	if len(roles) > 0 && roles[0] == models.RoleSales {
		// Sales receives operational metrics only; revenue data stays server-side.
		summary.MRR = 0
		summary.DueInvoices = 0
		summary.RevenueThisMonth = 0
		summary.OutstandingAmount = 0
		summary.OverdueAmount = 0
		summary.PaymentsSucceeded = 0
		summary.PaymentsFailed = 0
		summary.AutopaySucceeded = 0
		summary.AutopayFailed = 0
		summary.RevenueByPlan = []models.PlanRevenuePoint{}
		summary.MonthlyPerformance = []models.MonthlyPerformancePoint{}
		summary.YearlyBreakdown = []models.YearlyBreakdownPoint{}
	}

	return summary, nil
}

func (s *DashboardService) GetPlatformSummary() (models.PlatformDashboardSummary, error) {
	return s.repo.PlatformSummary()
}

func fillMonthlyPerformance(rows []revenuerepo.MonthlyRevenuePoint) []models.MonthlyPerformancePoint {
	values := make(map[string]float64, len(rows))
	for _, row := range rows {
		values[row.Month] = row.Value
	}

	monthNames := [...]string{"Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agt", "Sep", "Okt", "Nov", "Des"}
	now := time.Now().UTC()
	points := make([]models.MonthlyPerformancePoint, 0, 7)
	for offset := -6; offset <= 0; offset++ {
		month := now.AddDate(0, offset, 0)
		key := month.Format("2006-01")
		points = append(points, models.MonthlyPerformancePoint{
			Month: monthNames[month.Month()-1],
			Value: values[key],
		})
	}
	return points
}

func fillYearlyBreakdown(rows []revenuerepo.YearlyRevenuePoint) []models.YearlyBreakdownPoint {
	values := make(map[string]float64, len(rows))
	for _, row := range rows {
		values[row.Year] = row.Income
	}

	currentYear := time.Now().UTC().Year()
	points := make([]models.YearlyBreakdownPoint, 0, 5)
	for offset := -4; offset <= 0; offset++ {
		year := strconv.Itoa(currentYear + offset)
		points = append(points, models.YearlyBreakdownPoint{
			Year:   year,
			Income: values[year],
		})
	}
	return points
}
