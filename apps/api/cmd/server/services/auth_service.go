package services

import (
	"api/cmd/server/repository"
	"api/cmd/server/utils"
	"context"
	"database/models"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthService struct {
	userRepo             repository.UserRepository
	jwtSecret            string
	jwtRefreshSecret     string
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
}

func NewAuthService(
	userRepo repository.UserRepository,
	jwtSecret string,
	jwtRefreshSecret string,
	accessTokenDuration time.Duration,
	refreshTokenDuration time.Duration,
) *AuthService {
	return &AuthService{
		userRepo:             userRepo,
		jwtSecret:            jwtSecret,
		jwtRefreshSecret:     jwtRefreshSecret,
		accessTokenDuration:  accessTokenDuration,
		refreshTokenDuration: refreshTokenDuration,
	}
}

func (s *AuthService) SignUp(ctx context.Context, req SignUpRequest) (*AuthResponse, error) {
	existing, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err == nil && existing != nil {
		return nil, utils.ErrEmailAlreadyExists
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to query database: %w", err)
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := models.User{
		Email:    req.Email,
		Password: hashedPassword,
		Name:     req.Name,
	}

	if err := s.userRepo.CreateUser(ctx, &user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	accessToken, err := utils.GenerateAccessToken(user.ID, user.Email, s.jwtSecret, s.accessTokenDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID, user.Email, s.jwtRefreshSecret, s.refreshTokenDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	if err := s.userRepo.UpdateRefreshToken(ctx, user.ID, refreshToken); err != nil {
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}, nil
}

func (s *AuthService) SignIn(ctx context.Context, req SignInRequest) (*AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to query database: %w", err)
	}

	if !utils.CheckPasswordHash(req.Password, user.Password) {
		return nil, utils.ErrInvalidCredentials
	}

	accessToken, err := utils.GenerateAccessToken(user.ID, user.Email, s.jwtSecret, s.accessTokenDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID, user.Email, s.jwtRefreshSecret, s.refreshTokenDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	if err := s.userRepo.UpdateRefreshToken(ctx, user.ID, refreshToken); err != nil {
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	claims, err := utils.ValidateToken(refreshToken, s.jwtRefreshSecret, utils.TokenTypeRefresh)
	if err != nil {
		return nil, utils.ErrInvalidRefreshToken
	}

	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to query database: %w", err)
	}

	if user.RefreshToken != refreshToken {
		return nil, utils.ErrInvalidRefreshToken
	}

	newAccessToken, err := utils.GenerateAccessToken(user.ID, user.Email, s.jwtSecret, s.accessTokenDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := utils.GenerateRefreshToken(user.ID, user.Email, s.jwtRefreshSecret, s.refreshTokenDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	if err := s.userRepo.UpdateRefreshToken(ctx, user.ID, newRefreshToken); err != nil {
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	return &TokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID) error {
	return s.userRepo.UpdateRefreshToken(ctx, userID, "")
}

func (s *AuthService) GetProfile(ctx context.Context, userID uuid.UUID) (*UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to query database: %w", err)
	}

	return &UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}
