package model

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	ID        uuid.UUID  `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID  `gorm:"type:uuid;index;not null" json:"user_id"`
	Token     string     `gorm:"uniqueIndex;not null" json:"token"`
	ExpiresAt time.Time  `gorm:"not null" json:"expires_at"`
	RevokedAt *time.Time `gorm:"type:TIMESTAMPTZ" json:"revoked_at"`
	CreatedAt time.Time  `gorm:"column:created_at;type:TIMESTAMPTZ" json:"created_at"`
}
