package model

import (
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`

	ActorID string `gorm:"type:uuid;index" json:"actor_id"`
	UserID  string `gorm:"type:uuid;index" json:"user_id"`

	Action   string `gorm:"size:100;index" json:"action"`
	Entity   string `gorm:"size:100;index" json:"entity"`
	EntityID string `gorm:"size:255" json:"entity_id"`

	Metadata []byte `gorm:"type:jsonb" json:"metadata"`

	CreatedAt time.Time
}
