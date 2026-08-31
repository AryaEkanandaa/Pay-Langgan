package identity

import (
	"fmt"
	"net/mail"
	"strings"

	"pay-langgan/internal/config"
	"pay-langgan/internal/models"
	"pay-langgan/internal/repositories/identity"
	"pay-langgan/internal/utils"
)

type AuthService struct {
	authRepo *identity.AuthRepository
	cfg      *config.Config
}

func NewAuthService(authRepo *identity.AuthRepository, cfg *config.Config) *AuthService {
	return &AuthService{
		authRepo: authRepo,
		cfg:      cfg,
	}
}

func (s *AuthService) Signup(req models.SignupRequest) (*models.AuthResponse, error) {
	req.BusinessName = strings.TrimSpace(req.BusinessName)
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if req.BusinessName == "" || req.Name == "" || req.Email == "" || req.Password == "" {
		return nil, fmt.Errorf("%w: required fields are missing", utils.ErrBadRequest)
	}
	if len(req.BusinessName) > 100 || len(req.Name) > 100 || len(req.Email) > 100 {
		return nil, fmt.Errorf("%w: input exceeds maximum length", utils.ErrBadRequest)
	}
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return nil, fmt.Errorf("%w: invalid email", utils.ErrBadRequest)
	}

	if len(req.Password) < 6 {
		return nil, fmt.Errorf("%w: password must be at least 6 characters", utils.ErrBadRequest)
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
		BusinessID: &business.ID,
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
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if req.Email == "" || req.Password == "" {
		return nil, fmt.Errorf("%w: email and password are required", utils.ErrBadRequest)
	}
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return nil, fmt.Errorf("%w: invalid email", utils.ErrBadRequest)
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

	if !models.IsValidRole(user.Role) {
		return nil, utils.ErrUnauthorized
	}

	var business *models.Business
	if user.BusinessID != nil {
		business, err = s.authRepo.FindBusinessByID(*user.BusinessID)
		if err != nil {
			return nil, fmt.Errorf("find business: %w", err)
		}
		if business == nil || business.Deleted || !models.IsTenantRole(user.Role) {
			return nil, utils.ErrUnauthorized
		}
	} else if user.Role != string(models.RoleSuperAdmin) {
		return nil, utils.ErrUnauthorized
	}

	businessID := ""
	if business != nil {
		businessID = business.ID
	}

	token, err := utils.GenerateToken(
		s.cfg.JWTSecret,
		s.cfg.JWTExpiresIn,
		user.ID,
		businessID,
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
			BusinessID: businessID,
			Name:       user.Name,
			Email:      user.Email,
			Role:       user.Role,
		},
	}, nil
}

func (s *AuthService) GetMe(userID int, businessID, email, role string) (*models.MeResponse, error) {
	if role == string(models.RoleSuperAdmin) && businessID == "" {
		return &models.MeResponse{
			UserID:     userID,
			BusinessID: businessID,
			Email:      email,
			Role:       role,
		}, nil
	}

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
