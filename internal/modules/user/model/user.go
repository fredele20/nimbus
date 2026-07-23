package model

import (
	"time"

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
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FirstName       string    `gorm:"size:255;not null"`
	LastName        string    `gorm:"size:255;not null"`
	Email           string    `gorm:"size:255;uniqueIndex;not null"`
	Password        string    `gorm:"not null"`
	Role            Role      `gorm:"size:50;not null;default:customer"`
	Status          Status    `gorm:"size:20;default:active"`
	EmailVerifiedAt *time.Time
	LastLoginAt     *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt
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
