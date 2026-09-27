package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Email        string    `gorm:"uniqueIndex;not null" json:"email"`
	Password     string    `gorm:"not null" json:"-"`
	Name         string    `json:"name"`
	RefreshToken string    `gorm:"type:text" json:"-"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName specifies the table name for User
func (User) TableName() string {
	return "users"
}

// Users is an alias to User to support both singular and plural naming
type Users = User
