package services

import (
	"pay-langgan/internal/models"
	"pay-langgan/internal/repositories"
	"pay-langgan/internal/utils"
)

type CustomerService struct {
	customerRepo *repositories.CustomerRepository
}

func NewCustomerService(customerRepo *repositories.CustomerRepository) *CustomerService {
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
	if req.Name == "" {
		return nil, utils.ErrBadRequest
	}

	customer := &models.Customer{
		BusinessID: businessID,
		Name:       req.Name,
		Email:      req.Email,
		Contact:    req.Contact,
		Meta:       req.Meta,
	}

	err := s.customerRepo.Create(customer)
	if err != nil {
		return nil, err
	}
	return customer, nil
}

func (s *CustomerService) Update(id int, businessID string, req models.UpdateCustomerRequest) (*models.Customer, error) {
	customer, err := s.customerRepo.FindByIDAndBusinessID(id, businessID)
	if err != nil {
		return nil, err
	}
	if customer == nil {
		return nil, utils.ErrNotFound
	}

	if req.Name != "" {
		customer.Name = req.Name
	}
	if req.Email != nil {
		customer.Email = req.Email
	}
	if req.Contact != nil {
		customer.Contact = req.Contact
	}
	if req.Meta != nil {
		customer.Meta = req.Meta
	}

	err = s.customerRepo.Update(customer)
	if err != nil {
		return nil, err
	}
	return customer, nil
}

func (s *CustomerService) Delete(id int, businessID string) error {
	customer, err := s.customerRepo.FindByIDAndBusinessID(id, businessID)
	if err != nil {
		return err
	}
	if customer == nil {
		return utils.ErrNotFound
	}

	return s.customerRepo.DeleteByIDAndBusinessID(id, businessID)
}
