package services

import (
	"api/cmd/server/repository"
	"api/cmd/server/utils"
	"context"
	"database/models"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthService struct {
	userRepo             repository.UserRepository
	otpRepo              repository.OTPRepository
	jwtSecret            string
	jwtRefreshSecret     string
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
}

func NewAuthService(
	userRepo repository.UserRepository,
	otpRepo repository.OTPRepository,
	jwtSecret string,
	jwtRefreshSecret string,
	accessTokenDuration time.Duration,
	refreshTokenDuration time.Duration,
) *AuthService {
	return &AuthService{
		userRepo:             userRepo,
		otpRepo:              otpRepo,
		jwtSecret:            jwtSecret,
		jwtRefreshSecret:     jwtRefreshSecret,
		accessTokenDuration:  accessTokenDuration,
		refreshTokenDuration: refreshTokenDuration,
	}
}

func (s *AuthService) toUserResponse(user *models.User) UserResponse {
	resp := UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
	if user.Email != nil {
		resp.Email = *user.Email
	}
	if user.Phone != nil {
		resp.Phone = *user.Phone
	}
	return resp
}

func (s *AuthService) SignUp(ctx context.Context, req SignUpRequest) (*AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	existing, err := s.userRepo.FindByEmail(ctx, email)
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
		Email:    &email,
		Password: hashedPassword,
		Name:     req.Name,
	}

	if err := s.userRepo.CreateUser(ctx, &user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	emailStr := ""
	if user.Email != nil {
		emailStr = *user.Email
	}

	accessToken, err := utils.GenerateAccessToken(user.ID, emailStr, "", s.jwtSecret, s.accessTokenDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID, emailStr, "", s.jwtRefreshSecret, s.refreshTokenDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	if err := s.userRepo.UpdateRefreshToken(ctx, user.ID, refreshToken); err != nil {
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         s.toUserResponse(&user),
	}, nil
}

func (s *AuthService) SignIn(ctx context.Context, req SignInRequest) (*AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to query database: %w", err)
	}

	if !utils.CheckPasswordHash(req.Password, user.Password) {
		return nil, utils.ErrInvalidCredentials
	}

	emailStr := ""
	if user.Email != nil {
		emailStr = *user.Email
	}
	phoneStr := ""
	if user.Phone != nil {
		phoneStr = *user.Phone
	}

	accessToken, err := utils.GenerateAccessToken(user.ID, emailStr, phoneStr, s.jwtSecret, s.accessTokenDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID, emailStr, phoneStr, s.jwtRefreshSecret, s.refreshTokenDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	if err := s.userRepo.UpdateRefreshToken(ctx, user.ID, refreshToken); err != nil {
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         s.toUserResponse(user),
	}, nil
}

func (s *AuthService) SendOTP(ctx context.Context, req SendOTPRequest) (*SendOTPResponse, error) {
	phone := strings.TrimSpace(req.Phone)

	// Invalidate previous active OTPs for this phone
	_ = s.otpRepo.InvalidatePreviousOTPs(ctx, phone)

	code, err := utils.GenerateNumericOTP(6)
	if err != nil {
		return nil, fmt.Errorf("failed to generate OTP: %w", err)
	}

	otp := models.OTP{
		Phone:     phone,
		Code:      code,
		ExpiresAt: time.Now().Add(5 * time.Minute),
		Used:      false,
	}

	if err := s.otpRepo.CreateOTP(ctx, &otp); err != nil {
		return nil, fmt.Errorf("failed to save OTP: %w", err)
	}

	log.Printf("[AUTH] Generated OTP for %s: %s (expires in 5 minutes)", phone, code)

	return &SendOTPResponse{
		Phone: phone,
	}, nil
}

func (s *AuthService) VerifyOTP(ctx context.Context, req VerifyOTPRequest) (*AuthResponse, error) {
	phone := strings.TrimSpace(req.Phone)
	otp, err := s.otpRepo.FindValidOTP(ctx, phone, req.OTP)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrInvalidOTP
		}
		return nil, fmt.Errorf("failed to verify OTP: %w", err)
	}

	// Mark OTP as used
	if err := s.otpRepo.MarkOTPAsUsed(ctx, otp.ID); err != nil {
		return nil, fmt.Errorf("failed to consume OTP: %w", err)
	}

	// Find existing user or automatically create a new user (login / signup)
	user, err := s.userRepo.FindByPhone(ctx, phone)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to query database: %w", err)
	}

	if errors.Is(err, gorm.ErrRecordNotFound) || user == nil {
		newUser := models.User{
			Phone: &phone,
		}
		if err := s.userRepo.CreateUser(ctx, &newUser); err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
		user = &newUser
	}

	emailStr := ""
	if user.Email != nil {
		emailStr = *user.Email
	}
	phoneStr := ""
	if user.Phone != nil {
		phoneStr = *user.Phone
	}

	accessToken, err := utils.GenerateAccessToken(user.ID, emailStr, phoneStr, s.jwtSecret, s.accessTokenDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID, emailStr, phoneStr, s.jwtRefreshSecret, s.refreshTokenDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	if err := s.userRepo.UpdateRefreshToken(ctx, user.ID, refreshToken); err != nil {
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         s.toUserResponse(user),
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

	emailStr := ""
	if user.Email != nil {
		emailStr = *user.Email
	}
	phoneStr := ""
	if user.Phone != nil {
		phoneStr = *user.Phone
	}

	newAccessToken, err := utils.GenerateAccessToken(user.ID, emailStr, phoneStr, s.jwtSecret, s.accessTokenDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := utils.GenerateRefreshToken(user.ID, emailStr, phoneStr, s.jwtRefreshSecret, s.refreshTokenDuration)
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

	resp := s.toUserResponse(user)
	return &resp, nil
}
