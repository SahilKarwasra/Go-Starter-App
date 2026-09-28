package services

import (
	"time"

	"github.com/google/uuid"
)

type SignUpRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name"`
}

type SignInRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type SendOTPRequest struct {
	Phone string `json:"phone" binding:"required,min=8,max=16"`
}

type SendOTPResponse struct {
	Phone string `json:"phone"`
}

type VerifyOTPRequest struct {
	Phone string `json:"phone" binding:"required,min=8,max=16"`
	OTP   string `json:"otp" binding:"required,len=6,numeric"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email,omitempty"`
	Phone     string    `json:"phone,omitempty"`
	Name      string    `json:"name,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type AuthResponse struct {
	AccessToken  string       `json:"accessToken"`
	RefreshToken string       `json:"refreshToken"`
	User         UserResponse `json:"user"`
}

type TokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}
