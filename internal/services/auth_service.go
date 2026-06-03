package services

import (
	"fmt"

	"pay-langgan/internal/config"
	"pay-langgan/internal/models"
	"pay-langgan/internal/repositories"
	"pay-langgan/internal/utils"
)

type AuthService struct {
	authRepo *repositories.AuthRepository
	cfg      *config.Config
}

func NewAuthService(authRepo *repositories.AuthRepository, cfg *config.Config) *AuthService {
	return &AuthService{
		authRepo: authRepo,
		cfg:      cfg,
	}
}

func (s *AuthService) Signup(req models.SignupRequest) (*models.AuthResponse, error) {
	if req.BusinessName == "" || req.Name == "" || req.Email == "" || req.Password == "" {
		return nil, utils.ErrBadRequest
	}

	if len(req.Password) < 6 {
		return nil, fmt.Errorf("password must be at least 6 characters")
	}

	existingUser, err := s.authRepo.FindUserByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("check email: %w", err)
	}
	if existingUser != nil {
		return nil, utils.ErrConflict
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	business := &models.Business{
		ID:   utils.GenerateBusinessID(),
		Name: req.BusinessName,
	}

	user := &models.User{
		BusinessID: business.ID,
		Name:       req.Name,
		Email:      req.Email,
		Password:   hashedPassword,
		Role:       "admin",
	}

	if err := s.authRepo.CreateBusinessAndUserTx(business, user); err != nil {
		return nil, fmt.Errorf("create business and user: %w", err)
	}

	token, err := utils.GenerateToken(
		s.cfg.JWTSecret,
		s.cfg.JWTExpiresIn,
		user.ID,
		business.ID,
		user.Email,
		user.Role,
	)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &models.AuthResponse{
		Token: token,
		User: models.UserDTO{
			ID:         user.ID,
			BusinessID: business.ID,
			Name:       user.Name,
			Email:      user.Email,
			Role:       user.Role,
		},
		Business: models.BusinessDTO{
			ID:     business.ID,
			Name:   business.Name,
			Status: business.Status,
		},
	}, nil
}

func (s *AuthService) Login(req models.LoginRequest) (*models.LoginResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, utils.ErrBadRequest
	}

	user, err := s.authRepo.FindUserByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	if user == nil {
		return nil, utils.ErrUnauthorized
	}

	if !utils.CheckPassword(req.Password, user.Password) {
		return nil, utils.ErrUnauthorized
	}

	business, err := s.authRepo.FindBusinessByID(user.BusinessID)
	if err != nil {
		return nil, fmt.Errorf("find business: %w", err)
	}
	if business == nil || business.Deleted {
		return nil, utils.ErrUnauthorized
	}

	token, err := utils.GenerateToken(
		s.cfg.JWTSecret,
		s.cfg.JWTExpiresIn,
		user.ID,
		business.ID,
		user.Email,
		user.Role,
	)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &models.LoginResponse{
		Token: token,
		User: models.UserDTO{
			ID:         user.ID,
			BusinessID: business.ID,
			Name:       user.Name,
			Email:      user.Email,
			Role:       user.Role,
		},
	}, nil
}

func (s *AuthService) GetMe(userID int, businessID, email, role string) (*models.MeResponse, error) {
	business, err := s.authRepo.FindBusinessByID(businessID)
	if err != nil {
		return nil, fmt.Errorf("find business: %w", err)
	}
	if business == nil {
		return nil, utils.ErrNotFound
	}

	return &models.MeResponse{
		UserID:     userID,
		BusinessID: businessID,
		Email:      email,
		Role:       role,
	}, nil
}
