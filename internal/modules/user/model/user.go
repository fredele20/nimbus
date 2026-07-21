package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Status string

const (
	StatusActive    Status = "active"
	StatusSuspended Status = "suspended"
)

type User struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FirstName       string    `gorm:"size:255;not null"`
	LastName        string    `gorm:"size:255;not null"`
	Email           string    `gorm:"size:255;uniqueIndex;not null"`
	Password        string    `gorm:"not null"`
	Status          Status    `gorm:"size:20;default:active"`
	EmailVerifiedAt *time.Time
	LastLoginAt     *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt
}
