package database

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uuid.UUID      `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time      `gorm:"column:created_at;type:TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;type:TIMESTAMP" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at" json:"deleted_at"`
}
