package service

import (
	"context"

	"github.com/fredele20/nimbus/internal/modules/wallet/repository"
	"github.com/google/uuid"
)

type AdminService interface {
	FreezeWallet(ctx context.Context, walletID, adminID uuid.UUID, reason string) error

	UnfreezeWallet(ctx context.Context, walletID, adminID uuid.UUID) error
}

type adminService struct {
	repo repository.Repository
}

func NewAdmin(repo repository.Repository) AdminService {
	return &adminService{
		repo: repo,
	}
}

func (a *adminService) FreezeWallet(ctx context.Context, walletID, adminID uuid.UUID, reason string) error {
	if err := validateWalletID(walletID); err != nil {
		return err
	}

	wallet, err := a.repo.FindByWalletID(ctx, walletID)
	if err != nil {
		return err
	}

	if err := validateUserID(adminID); err != nil {
		return err
	}

	return translateRepositoryError(
		a.repo.ApplyRestriction(ctx, repository.RestrictionCommand{
			WalletID: walletID,
			OwnerID:  wallet.UserID,
			ActorID:  adminID,
			Action:   repository.ActionAdminFreeze,
			Reason:   reason,
		}),
	)
}

func (a *adminService) UnfreezeWallet(ctx context.Context, walletID, adminID uuid.UUID) error {
	if err := validateWalletID(walletID); err != nil {
		return err
	}

	wallet, err := a.repo.FindByWalletID(ctx, walletID)
	if err != nil {
		return err
	}

	if err := validateUserID(adminID); err != nil {
		return err
	}

	return translateRepositoryError(
		a.repo.ApplyRestriction(ctx, repository.RestrictionCommand{
			WalletID: walletID,
			OwnerID:  wallet.UserID,
			ActorID:  adminID,
			Action:   repository.ActionAdminUnfreeze,
		}),
	)
}
