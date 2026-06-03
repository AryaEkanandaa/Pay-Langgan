package services

import (
	"pay-langgan/internal/models"
	"pay-langgan/internal/repositories"
	"pay-langgan/internal/utils"
)

type ProductService struct {
	productRepo  *repositories.ProductRepository
	serviceRepo  *repositories.ServiceRepository
}

func NewProductService(productRepo *repositories.ProductRepository, serviceRepo *repositories.ServiceRepository) *ProductService {
	return &ProductService{
		productRepo: productRepo,
		serviceRepo: serviceRepo,
	}
}

func (s *ProductService) GetAll(businessID string, page, limit int, search string) ([]models.Product, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	return s.productRepo.FindAllByBusinessID(businessID, page, limit, search)
}

func (s *ProductService) GetByID(id int, businessID string) (*models.Product, error) {
	product, err := s.productRepo.FindByIDAndBusinessID(id, businessID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, utils.ErrNotFound
	}
	return product, nil
}

func (s *ProductService) Create(businessID string, req models.CreateProductRequest) (*models.Product, error) {
	if req.Name == "" {
		return nil, utils.ErrBadRequest
	}
	if req.ServiceID == 0 {
		return nil, utils.ErrBadRequest
	}

	service, err := s.serviceRepo.FindByIDAndBusinessID(req.ServiceID, businessID)
	if err != nil {
		return nil, err
	}
	if service == nil {
		return nil, utils.ErrNotFound
	}

	status := req.Status
	if status == "" {
		status = "active"
	}

	product := &models.Product{
		ServiceID:   req.ServiceID,
		Name:        req.Name,
		Description: req.Description,
		Status:      status,
		Meta:        req.Meta,
	}

	err = s.productRepo.Create(product)
	if err != nil {
		return nil, err
	}
	return product, nil
}

func (s *ProductService) Update(id int, businessID string, req models.UpdateProductRequest) (*models.Product, error) {
	product, err := s.productRepo.FindByIDAndBusinessID(id, businessID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, utils.ErrNotFound
	}

	if req.Name != "" {
		product.Name = req.Name
	}
	product.Description = req.Description
	if req.Status != "" {
		product.Status = req.Status
	}
	if req.Meta != nil {
		product.Meta = req.Meta
	}

	err = s.productRepo.Update(product)
	if err != nil {
		return nil, err
	}
	return product, nil
}

func (s *ProductService) Delete(id int, businessID string) error {
	product, err := s.productRepo.FindByIDAndBusinessID(id, businessID)
	if err != nil {
		return err
	}
	if product == nil {
		return utils.ErrNotFound
	}

	return s.productRepo.DeleteByIDAndBusinessID(id, businessID)
}
