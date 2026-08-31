package catalog

import (
	"fmt"
	"pay-langgan/internal/models"
	"pay-langgan/internal/repositories/catalog"
	"pay-langgan/internal/utils"
	"strings"
)

type AddOnService struct {
	addOnRepo   *catalog.AddOnRepository
	productRepo *catalog.ProductRepository
}

func NewAddOnService(addOnRepo *catalog.AddOnRepository, productRepo *catalog.ProductRepository) *AddOnService {
	return &AddOnService{
		addOnRepo:   addOnRepo,
		productRepo: productRepo,
	}
}

func (s *AddOnService) GetAll(businessID string, page, limit int, search string) ([]models.AddOn, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	return s.addOnRepo.FindAllByBusinessID(businessID, page, limit, search)
}

func (s *AddOnService) GetByID(id int, businessID string) (*models.AddOn, error) {
	addOn, err := s.addOnRepo.FindByIDAndBusinessID(id, businessID)
	if err != nil {
		return nil, err
	}
	if addOn == nil {
		return nil, utils.ErrNotFound
	}
	return addOn, nil
}

func (s *AddOnService) Create(businessID string, req models.CreateAddOnRequest) (*models.AddOn, error) {
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

	addOn := &models.AddOn{
		ProductID:    req.ProductID,
		Name:         req.Name,
		Price:        req.Price,
		BillingCycle: req.BillingCycle,
		Meta:         req.Meta,
	}

	err = s.addOnRepo.Create(addOn)
	if err != nil {
		return nil, err
	}
	return addOn, nil
}

func (s *AddOnService) Update(id int, businessID string, req models.UpdateAddOnRequest) (*models.AddOn, error) {
	addOn, err := s.addOnRepo.FindByIDAndBusinessID(id, businessID)
	if err != nil {
		return nil, err
	}
	if addOn == nil {
		return nil, utils.ErrNotFound
	}

	if req.Name != "" {
		req.Name = strings.TrimSpace(req.Name)
		if req.Name == "" || len(req.Name) > 100 {
			return nil, fmt.Errorf("%w: invalid add-on name", utils.ErrBadRequest)
		}
		addOn.Name = req.Name
	}
	if req.Price > 0 {
		addOn.Price = req.Price
	}
	if req.BillingCycle != "" {
		if req.BillingCycle != "monthly" && req.BillingCycle != "yearly" {
			return nil, utils.ErrBadRequest
		}
		addOn.BillingCycle = req.BillingCycle
	}
	if req.Meta != nil {
		addOn.Meta = req.Meta
	}

	err = s.addOnRepo.Update(addOn)
	if err != nil {
		return nil, err
	}
	return addOn, nil
}

func (s *AddOnService) Delete(id int, businessID string) error {
	addOn, err := s.addOnRepo.FindByIDAndBusinessID(id, businessID)
	if err != nil {
		return err
	}
	if addOn == nil {
		return utils.ErrNotFound
	}

	return s.addOnRepo.DeleteByIDAndBusinessID(id, businessID)
}
