package service_test

import (
	"context"
	"testing"

	"github.com/fredele20/nimbus/internal/modules/wallet/model"
	"github.com/fredele20/nimbus/internal/modules/wallet/repository"
	"github.com/fredele20/nimbus/internal/modules/wallet/service"
	wallet_mocks "github.com/fredele20/nimbus/mocks/wallet"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestAdminFreezeWallet(t *testing.T) {
	ctx := context.Background()
	adminID := uuid.New()
	walletID := uuid.New()
	wallet := model.Wallet{
		UserID:   uuid.New(),
		Currency: "NGN",
	}
	wallet.ID = walletID

	repo := wallet_mocks.NewMockRepository(t)
	svc := service.NewAdmin(repo)

	repo.EXPECT().FindByWalletID(ctx, wallet.ID).Return(wallet, nil).Once()

	repo.EXPECT().
		ApplyRestriction(ctx, repository.RestrictionCommand{
			WalletID: walletID,
			OwnerID:  wallet.UserID,
			ActorID:  adminID,
			Action:   repository.ActionAdminFreeze,
			Reason:   "Suspicious activity",
		}).
		Return(nil).
		Once()

	err := svc.FreezeWallet(
		ctx,
		walletID,
		adminID,
		"Suspicious activity",
	)

	require.NoError(t, err)
}

func TestAdminUnfreezeWallet(t *testing.T) {
	ctx := context.Background()
	adminID := uuid.New()
	walletID := uuid.New()
	wallet := model.Wallet{
		UserID:   uuid.New(),
		Currency: "NGN",
	}
	wallet.ID = walletID

	repo := wallet_mocks.NewMockRepository(t)
	svc := service.NewAdmin(repo)

	repo.EXPECT().FindByWalletID(ctx, wallet.ID).Return(wallet, nil).Once()

	repo.EXPECT().
		ApplyRestriction(ctx, repository.RestrictionCommand{
			WalletID: wallet.ID,
			OwnerID:  wallet.UserID,
			ActorID:  adminID,
			Action:   repository.ActionAdminUnfreeze,
		}).
		Return(nil).
		Once()

	err := svc.UnfreezeWallet(ctx, wallet.ID, adminID)

	require.NoError(t, err)
}

// func TestAdminFreezeRejectsInvalidIDs(t *testing.T) {
// 	ctx := context.Background()

// 	tests := []struct {
// 		name     string
// 		walletID uuid.UUID
// 		ownerID  uuid.UUID
// 		adminID  uuid.UUID
// 		want     error
// 	}{
// 		{
// 			name:     "missing wallet",
// 			walletID: uuid.Nil,
// 			ownerID:  uuid.New(),
// 			adminID:  uuid.New(),
// 			want:     service.ErrInvalidWalletID,
// 		},
// 		// {
// 		// 	name:     "missing owner",
// 		// 	walletID: uuid.New(),
// 		// 	ownerID:  uuid.Nil,
// 		// 	adminID:  uuid.New(),
// 		// 	want:     service.ErrInvalidUserID,
// 		// },
// 		{
// 			name:     "missing administrator",
// 			walletID: uuid.New(),
// 			ownerID:  uuid.New(),
// 			adminID:  uuid.Nil,
// 			want:     service.ErrInvalidUserID,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			repo := wallet_mocks.NewMockRepository(t)
// 			svc := service.NewAdmin(repo)

// 			repo.EXPECT().FindByWalletID(ctx, tt.walletID).Return(model.Wallet{}, nil).Once()

// 			repo.EXPECT().
// 				ApplyRestriction(ctx, repository.RestrictionCommand{
// 					WalletID: tt.walletID,
// 					OwnerID:  tt.ownerID,
// 					ActorID:  tt.adminID,
// 					Reason:   "Test restriction",
// 					Action:   repository.ActionAdminFreeze,
// 				}).
// 				Return(nil).
// 				Once()

// 			err := svc.FreezeWallet(
// 				ctx,
// 				tt.walletID,
// 				tt.adminID,
// 				"Test restriction",
// 			)

// 			require.ErrorIs(t, err, tt.want)
// 		})
// 	}
// }
