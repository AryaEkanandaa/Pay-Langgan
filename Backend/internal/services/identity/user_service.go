package identity

import (
	"fmt"
	"net/mail"
	"strings"

	"pay-langgan/internal/models"
	identityrepo "pay-langgan/internal/repositories/identity"
	"pay-langgan/internal/utils"
)

type UserService struct {
	userRepo *identityrepo.UserRepository
}

func NewUserService(userRepo *identityrepo.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) ListBusinessUsers(businessID string) ([]models.UserDTO, error) {
	users, err := s.userRepo.FindAllByBusinessID(businessID)
	if err != nil {
		return nil, err
	}

	result := make([]models.UserDTO, 0, len(users))
	for _, user := range users {
		result = append(result, toUserDTO(&user))
	}
	return result, nil
}

func (s *UserService) CreateStaff(businessID string, req models.CreateStaffRequest) (*models.UserDTO, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Role = strings.ToLower(strings.TrimSpace(req.Role))

	if businessID == "" || req.Name == "" || req.Email == "" || req.Password == "" || req.Role == "" {
		return nil, fmt.Errorf("%w: required fields are missing", utils.ErrBadRequest)
	}
	if len(req.Name) > 100 || len(req.Email) > 100 {
		return nil, fmt.Errorf("%w: input exceeds maximum length", utils.ErrBadRequest)
	}
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return nil, fmt.Errorf("%w: invalid email", utils.ErrBadRequest)
	}
	if len(req.Password) < 6 {
		return nil, fmt.Errorf("%w: password must be at least 6 characters", utils.ErrBadRequest)
	}
	if !models.IsStaffRole(req.Role) {
		return nil, fmt.Errorf("%w: only sales and finance staff can be created here", utils.ErrBadRequest)
	}

	existing, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, utils.ErrConflict
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &models.User{
		BusinessID: &businessID,
		Name:       req.Name,
		Email:      req.Email,
		Password:   hashedPassword,
		Role:       req.Role,
	}
	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	dto := toUserDTO(user)
	return &dto, nil
}

func toUserDTO(user *models.User) models.UserDTO {
	businessID := ""
	if user.BusinessID != nil {
		businessID = *user.BusinessID
	}
	return models.UserDTO{
		ID:         user.ID,
		BusinessID: businessID,
		Name:       user.Name,
		Email:      user.Email,
		Role:       user.Role,
	}
}
