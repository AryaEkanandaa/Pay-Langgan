package catalog

import (
	"fmt"
	"pay-langgan/internal/models"
	"pay-langgan/internal/repositories/catalog"
	"pay-langgan/internal/utils"
	"strings"
)

type PlanService struct {
	planRepo    *catalog.PlanRepository
	productRepo *catalog.ProductRepository
}

func NewPlanService(planRepo *catalog.PlanRepository, productRepo *catalog.ProductRepository) *PlanService {
	return &PlanService{
		planRepo:    planRepo,
		productRepo: productRepo,
	}
}

func (s *PlanService) GetAll(businessID string, page, limit int, search string) ([]models.Plan, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	return s.planRepo.FindAllByBusinessID(businessID, page, limit, search)
}

func (s *PlanService) GetByID(id int, businessID string) (*models.Plan, error) {
	plan, err := s.planRepo.FindByIDAndBusinessID(id, businessID)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, utils.ErrNotFound
	}
	return plan, nil
}

func (s *PlanService) Create(businessID string, req models.CreatePlanRequest) (*models.Plan, error) {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return nil, fmt.Errorf("%w: name is required", utils.ErrBadRequest)
	}
	if req.ProductID < 1 {
		return nil, utils.ErrBadRequest
	}
	if req.Price <= 0 {
		return nil, utils.ErrBadRequest
	}
	if req.BillingCycle != "monthly" && req.BillingCycle != "yearly" {
		return nil, utils.ErrBadRequest
	}
	if req.TrialDays < 0 {
		return nil, utils.ErrBadRequest
	}
	if len(req.Name) > 100 {
		return nil, fmt.Errorf("%w: name must be at most 100 characters", utils.ErrBadRequest)
	}

	product, err := s.productRepo.FindByIDAndBusinessID(req.ProductID, businessID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, utils.ErrNotFound
	}

	plan := &models.Plan{
		ProductID:    req.ProductID,
		Name:         req.Name,
		Price:        req.Price,
		BillingCycle: req.BillingCycle,
		TrialDays:    req.TrialDays,
		Meta:         req.Meta,
	}

	err = s.planRepo.Create(plan)
	if err != nil {
		return nil, err
	}
	return plan, nil
}

func (s *PlanService) Update(id int, businessID string, req models.UpdatePlanRequest) (*models.Plan, error) {
	plan, err := s.planRepo.FindByIDAndBusinessID(id, businessID)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, utils.ErrNotFound
	}

	if req.Name != "" {
		req.Name = strings.TrimSpace(req.Name)
		if req.Name == "" || len(req.Name) > 100 {
			return nil, fmt.Errorf("%w: invalid plan name", utils.ErrBadRequest)
		}
		plan.Name = req.Name
	}
	if req.Price > 0 {
		plan.Price = req.Price
	}
	if req.BillingCycle != "" {
		if req.BillingCycle != "monthly" && req.BillingCycle != "yearly" {
			return nil, utils.ErrBadRequest
		}
		plan.BillingCycle = req.BillingCycle
	}
	if req.TrialDays != nil {
		if *req.TrialDays < 0 {
			return nil, utils.ErrBadRequest
		}
		plan.TrialDays = *req.TrialDays
	}
	if req.Meta != nil {
		plan.Meta = req.Meta
	}

	err = s.planRepo.Update(plan)
	if err != nil {
		return nil, err
	}
	return plan, nil
}

func (s *PlanService) Delete(id int, businessID string) error {
	plan, err := s.planRepo.FindByIDAndBusinessID(id, businessID)
	if err != nil {
		return err
	}
	if plan == nil {
		return utils.ErrNotFound
	}

	return s.planRepo.DeleteByIDAndBusinessID(id, businessID)
}
