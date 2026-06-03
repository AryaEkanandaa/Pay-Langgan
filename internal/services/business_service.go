package services

import (
	"pay-langgan/internal/models"
	"pay-langgan/internal/repositories"
	"pay-langgan/internal/utils"
)

type BusinessService struct {
	businessRepo *repositories.BusinessRepository
}

func NewBusinessService(businessRepo *repositories.BusinessRepository) *BusinessService {
	return &BusinessService{businessRepo: businessRepo}
}

func (s *BusinessService) GetMyBusiness(businessID string) (*models.Business, error) {
	business, err := s.businessRepo.FindByID(businessID)
	if err != nil {
		return nil, err
	}
	if business == nil {
		return nil, utils.ErrNotFound
	}
	return business, nil
}

func (s *BusinessService) UpdateMyBusiness(businessID string, req models.UpdateBusinessRequest) error {
	business, err := s.businessRepo.FindByID(businessID)
	if err != nil {
		return err
	}
	if business == nil {
		return utils.ErrNotFound
	}

	if req.Name != "" {
		business.Name = req.Name
	}
	if req.Meta != nil {
		business.Meta = req.Meta
	}

	return s.businessRepo.Update(business)
}
