package model

import (
	"time"

	"github.com/fredele20/nimbus/internal/platform/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role string

const (
	Customer Role = "customer"
	Admin    Role = "admin"
)

type Status string

const (
	StatusActive    Status = "active"
	StatusSuspended Status = "suspended"
	StatusDeleted   Status = "deleted"
)

type User struct {
	database.BaseModel
	FirstName       string     `gorm:"size:255;not null" json:"first_name"`
	LastName        string     `gorm:"size:255;not null" json:"last_name"`
	Email           string     `gorm:"size:255;uniqueIndex;not null" json:"email"`
	Password        string     `gorm:"not null" json:"password"`
	Role            Role       `gorm:"size:50;not null;default:customer" json:"role"`
	Status          Status     `gorm:"size:20;default:active" json:"status"`
	EmailVerifiedAt *time.Time `gorm:"type:TIMESTAMP" json:"email_verified_at"`
	LastLoginAt     *time.Time `gorm:"type:TIMESTAMP" json:"last_login_at"`
}

func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}

	if u.Role == "" {
		u.Role = Customer
	}

	if u.Status == "" {
		u.Status = StatusActive
	}

	return nil
}
