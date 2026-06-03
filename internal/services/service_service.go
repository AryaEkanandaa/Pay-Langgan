package services

import (
	"pay-langgan/internal/models"
	"pay-langgan/internal/repositories"
	"pay-langgan/internal/utils"
)

type ServiceService struct {
	serviceRepo *repositories.ServiceRepository
}

func NewServiceService(serviceRepo *repositories.ServiceRepository) *ServiceService {
	return &ServiceService{serviceRepo: serviceRepo}
}

func (s *ServiceService) GetAll(businessID string, page, limit int, search string) ([]models.Service, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	return s.serviceRepo.FindAllByBusinessID(businessID, page, limit, search)
}

func (s *ServiceService) GetByID(id int, businessID string) (*models.Service, error) {
	service, err := s.serviceRepo.FindByIDAndBusinessID(id, businessID)
	if err != nil {
		return nil, err
	}
	if service == nil {
		return nil, utils.ErrNotFound
	}
	return service, nil
}

func (s *ServiceService) Create(businessID string, req models.CreateServiceRequest) (*models.Service, error) {
	if req.Name == "" {
		return nil, utils.ErrBadRequest
	}

	service := &models.Service{
		BusinessID:  businessID,
		Name:        req.Name,
		Description: req.Description,
		Meta:        req.Meta,
	}

	err := s.serviceRepo.Create(service)
	if err != nil {
		return nil, err
	}
	return service, nil
}

func (s *ServiceService) Update(id int, businessID string, req models.UpdateServiceRequest) (*models.Service, error) {
	service, err := s.serviceRepo.FindByIDAndBusinessID(id, businessID)
	if err != nil {
		return nil, err
	}
	if service == nil {
		return nil, utils.ErrNotFound
	}

	if req.Name != "" {
		service.Name = req.Name
	}
	service.Description = req.Description
	if req.Meta != nil {
		service.Meta = req.Meta
	}

	err = s.serviceRepo.Update(service)
	if err != nil {
		return nil, err
	}
	return service, nil
}

func (s *ServiceService) Delete(id int, businessID string) error {
	service, err := s.serviceRepo.FindByIDAndBusinessID(id, businessID)
	if err != nil {
		return err
	}
	if service == nil {
		return utils.ErrNotFound
	}

	return s.serviceRepo.DeleteByIDAndBusinessID(id, businessID)
}
