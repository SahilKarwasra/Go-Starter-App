package repository

import (
	"context"
	"database/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OTPRepository interface {
	CreateOTP(ctx context.Context, otp *models.OTP) error
	FindValidOTP(ctx context.Context, phone string, code string) (*models.OTP, error)
	MarkOTPAsUsed(ctx context.Context, id uuid.UUID) error
	InvalidatePreviousOTPs(ctx context.Context, phone string) error
}

type otpRepository struct {
	db *gorm.DB
}

func NewOTPRepository(db *gorm.DB) OTPRepository {
	return &otpRepository{db: db}
}

func (r *otpRepository) CreateOTP(ctx context.Context, otp *models.OTP) error {
	return r.db.WithContext(ctx).Create(otp).Error
}

func (r *otpRepository) FindValidOTP(ctx context.Context, phone string, code string) (*models.OTP, error) {
	var otp models.OTP
	now := time.Now()
	err := r.db.WithContext(ctx).
		Where("phone = ? AND code = ? AND used = ? AND expires_at > ?", phone, code, false, now).
		Order("created_at DESC").
		First(&otp).Error
	if err != nil {
		return nil, err
	}
	return &otp, nil
}

func (r *otpRepository) MarkOTPAsUsed(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&models.OTP{}).Where("id = ?", id).Update("used", true).Error
}

func (r *otpRepository) InvalidatePreviousOTPs(ctx context.Context, phone string) error {
	return r.db.WithContext(ctx).Model(&models.OTP{}).
		Where("phone = ? AND used = ?", phone, false).
		Update("used", true).Error
}
