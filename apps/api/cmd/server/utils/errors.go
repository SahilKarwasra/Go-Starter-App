package utils

import "errors"

var (
	ErrEmailAlreadyExists   = errors.New("email is already registered")
	ErrPhoneAlreadyExists   = errors.New("phone number is already registered")
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrUserNotFound        = errors.New("user not found")
	ErrUnauthorized        = errors.New("unauthorized access")
	ErrMissingToken        = errors.New("authorization header is required")
	ErrInvalidToken        = errors.New("invalid or expired token")
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
	ErrInvalidOTP          = errors.New("invalid or expired OTP")
)
