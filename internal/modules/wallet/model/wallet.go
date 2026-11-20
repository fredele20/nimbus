package model

import (
	"time"

	"github.com/fredele20/nimbus/internal/platform/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WalletStatus string

const (
	WalletStatusActive WalletStatus = "active"
	WalletStatusFrozen WalletStatus = "frozen"
	WalletStatusClosed WalletStatus = "closed"
)

type RestrictionType string

const (
	RestrictionTypeUser  RestrictionType = "user"
	RestrictionTypeAdmin RestrictionType = "admin"
)

type Wallet struct {
	database.BaseModel

	UserID   uuid.UUID    `gorm:"type:uuid;not null;uniqueIndex:uq_wallets_user_currency,priority:1" json:"user_id"`
	Currency string       `gorm:"size:3;not null;uniqueIndex:uq_wallets_user_currency,priority:2" json:"currency"`
	Status   WalletStatus `gorm:"size:20;not null;default:active;index" json:"status"`

	RestrictionType   *RestrictionType `gorm:"column:restriction_type;size:20" json:"restriction_type,omitempty"`
	RestrictionReason *string          `gorm:"column:restriction_reason;type:text" json:"restriction_reason,omitempty"`
	RestrictedBy      *uuid.UUID       `gorm:"column:restricted_by;type:uuid" json:"restricted_by,omitempty"`
	RestrictedAt      *time.Time       `gorm:"column:restricted_at;type:timestamptz" json:"restricted_at,omitempty"`
}

func (w *Wallet) BeforeCreate(_ *gorm.DB) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}

	if w.Status == "" {
		w.Status = WalletStatusActive
	}

	return nil
}
