package dto

import (
	"time"

	"github.com/fredele20/nimbus/internal/modules/wallet/model"
)

type CreateWalletRequest struct {
	Currency string `json:"currency" binding:"required"`
}

type AdminFreezeRequest struct {
	Reason string `json:"reason" binding:"required"`
}

type WalletResponse struct {
	ID                string
	UserID            string
	Currency          string
	Status            model.WalletStatus
	RestrictionType   *model.RestrictionType
	RestrictionReason *string
	RestrictedAt      *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func FromWallet(w model.Wallet) WalletResponse {
	return WalletResponse{
		ID:                w.ID.String(),
		UserID:            w.UserID.String(),
		Currency:          w.Currency,
		Status:            w.Status,
		RestrictionType:   w.RestrictionType,
		RestrictionReason: w.RestrictionReason,
		RestrictedAt:      w.RestrictedAt,
		CreatedAt:         w.CreatedAt,
		UpdatedAt:         w.CreatedAt,
	}
}

func FromWallets(wallets []model.Wallet) []WalletResponse {
	result := make([]WalletResponse, 0, len(wallets))

	for _, wallet := range wallets {
		result = append(result, FromWallet(wallet))
	}

	return result
}
