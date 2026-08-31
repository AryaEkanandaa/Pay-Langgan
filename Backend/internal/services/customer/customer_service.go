package customer

import (
	"fmt"
	"pay-langgan/internal/models"
	"pay-langgan/internal/repositories/customer"
	"pay-langgan/internal/utils"
	"strings"
)

type CustomerService struct {
	customerRepo *customer.CustomerRepository
}

func NewCustomerService(customerRepo *customer.CustomerRepository) *CustomerService {
	return &CustomerService{customerRepo: customerRepo}
}

func (s *CustomerService) GetAll(businessID string, page, limit int, search string) ([]models.Customer, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	return s.customerRepo.FindAllByBusinessID(businessID, page, limit, search)
}

func (s *CustomerService) GetByID(id int, businessID string) (*models.Customer, error) {
	customer, err := s.customerRepo.FindByIDAndBusinessID(id, businessID)
	if err != nil {
		return nil, err
	}
	if customer == nil {
		return nil, utils.ErrNotFound
	}
	return customer, nil
}

func (s *CustomerService) Create(businessID string, req models.CreateCustomerRequest) (*models.Customer, error) {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return nil, fmt.Errorf("%w: name is required", utils.ErrBadRequest)
	}
	if len(req.Name) > 100 {
		return nil, fmt.Errorf("%w: name must be at most 100 characters", utils.ErrBadRequest)
	}

	c := &models.Customer{
		BusinessID: businessID,
		Name:       req.Name,
		Email:      req.Email,
		Contact:    req.Contact,
		Meta:       req.Meta,
	}

	err := s.customerRepo.Create(c)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (s *CustomerService) Update(id int, businessID string, req models.UpdateCustomerRequest) (*models.Customer, error) {
	c, err := s.customerRepo.FindByIDAndBusinessID(id, businessID)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, utils.ErrNotFound
	}

	if req.Name != "" {
		req.Name = strings.TrimSpace(req.Name)
		if req.Name == "" || len(req.Name) > 100 {
			return nil, fmt.Errorf("%w: invalid customer name", utils.ErrBadRequest)
		}
		c.Name = req.Name
	}
	if req.Email != nil {
		c.Email = req.Email
	}
	if req.Contact != nil {
		c.Contact = req.Contact
	}
	if req.Meta != nil {
		c.Meta = req.Meta
	}

	err = s.customerRepo.Update(c)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (s *CustomerService) Delete(id int, businessID string) error {
	c, err := s.customerRepo.FindByIDAndBusinessID(id, businessID)
	if err != nil {
		return err
	}
	if c == nil {
		return utils.ErrNotFound
	}

	return s.customerRepo.DeleteByIDAndBusinessID(id, businessID)
}
