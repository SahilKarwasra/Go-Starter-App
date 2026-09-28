package models

import (
	"time"

	"github.com/google/uuid"
)

type OTP struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Phone     string    `gorm:"index;not null" json:"phone"`
	Code      string    `gorm:"not null" json:"code"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expiresAt"`
	Used      bool      `gorm:"default:false;not null" json:"used"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

func (OTP) TableName() string {
	return "otps"
}
