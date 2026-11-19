package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/fredele20/nimbus/internal/modules/wallet/dto"
	"github.com/fredele20/nimbus/internal/modules/wallet/model"
	"github.com/fredele20/nimbus/internal/modules/wallet/repository"
	"github.com/fredele20/nimbus/internal/modules/wallet/service"
	wallet_mocks "github.com/fredele20/nimbus/mocks/wallet"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestCreateWallet(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	repo := wallet_mocks.NewMockRepository(t)
	svc := service.New(repo)

	repo.EXPECT().
		Create(ctx, orgWallet(userID, "NGN")).
		Return(nil).
		Once()

	wallet, err := svc.Create(ctx, userID, " ngn ")

	require.NoError(t, err)
	require.NotNil(t, wallet)
	require.Equal(t, userID.String(), wallet.UserID)
	require.Equal(t, "NGN", wallet.Currency)
	require.Equal(t, model.WalletStatusActive, wallet.Status)
}

func TestCreateRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name     string
		userID   uuid.UUID
		currency string
		want     error
	}{
		{"missing user", uuid.Nil, "NGN", service.ErrInvalidUserID},
		{"unsupported currency", uuid.New(), "XYZ", service.ErrInvalidCurrency},
		{"empty currency", uuid.New(), "", service.ErrInvalidCurrency},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := wallet_mocks.NewMockRepository(t)
			svc := service.New(repo)

			wallet, err := svc.Create(
				context.Background(),
				tt.userID,
				tt.currency,
			)

			require.Empty(t, wallet)
			require.ErrorIs(t, err, tt.want)
		})
	}
}

func TestCreateDuplicateWallet(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	repo := wallet_mocks.NewMockRepository(t)
	svc := service.New(repo)

	repo.EXPECT().
		Create(ctx, orgWallet(userID, "NGN")).
		Return(&pgconn.PgError{
			Code:           "23505",
			ConstraintName: "uq_wallets_user_currency",
		}).
		Once()

	wallet, err := svc.Create(ctx, userID, "NGN")

	require.Empty(t, wallet)
	require.ErrorIs(t, err, service.ErrWalletAlreadyExists)
}

func TestGetByID(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	walletID := uuid.New()
	repo := wallet_mocks.NewMockRepository(t)
	svc := service.New(repo)

	wallet := model.Wallet{
		UserID:   userID,
		Currency: "NGN",
		Status:   model.WalletStatusActive,
	}

	repo.EXPECT().
		FindByID(ctx, walletID, userID).
		Return(wallet, nil).
		Once()

	walletResponse, err := svc.GetByID(ctx, walletID, userID)

	require.NoError(t, err)
	require.Equal(t, dto.FromWallet(wallet), walletResponse)
}

func TestGetByIDNotFound(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	walletID := uuid.New()
	repo := wallet_mocks.NewMockRepository(t)
	svc := service.New(repo)

	repo.EXPECT().
		FindByID(ctx, walletID, userID).
		Return(model.Wallet{}, gorm.ErrRecordNotFound).
		Once()

	wallet, err := svc.GetByID(ctx, walletID, userID)

	require.Empty(t, wallet)
	require.ErrorIs(t, err, service.ErrWalletNotFound)
}

func TestGetByCurrency(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	repo := wallet_mocks.NewMockRepository(t)
	svc := service.New(repo)

	wallet := model.Wallet{
		UserID:   userID,
		Currency: "USD",
		Status:   model.WalletStatusActive,
	}

	repo.EXPECT().
		FindByUserAndCurrency(ctx, userID, "USD").
		Return(wallet, nil).
		Once()

	walletResponse, err := svc.GetByCurrency(ctx, userID, " usd ")

	require.NoError(t, err)
	require.Equal(t, dto.FromWallet(wallet), walletResponse)
}

func TestListByUser(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	repo := wallet_mocks.NewMockRepository(t)
	svc := service.New(repo)

	expected := []model.Wallet{
		{UserID: userID, Currency: "NGN"},
		{UserID: userID, Currency: "USD"},
	}

	repo.EXPECT().
		FindByUserID(ctx, userID).
		Return(expected, nil).
		Once()

	wallets, err := svc.ListByUser(ctx, userID)

	require.NoError(t, err)
	require.Equal(t, dto.FromWallets(expected), wallets)
}

func TestFreezeOwnWallet(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	walletID := uuid.New()
	repo := wallet_mocks.NewMockRepository(t)
	svc := service.New(repo)

	repo.EXPECT().
		ApplyRestriction(ctx, repository.RestrictionCommand{
			WalletID: walletID,
			OwnerID:  userID,
			ActorID:  userID,
			Action:   repository.ActionUserFreeze,
		}).
		Return(nil).
		Once()

	err := svc.FreezeOwnWallet(ctx, walletID, userID)

	require.NoError(t, err)
}

func TestUnfreezeOwnWallet(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	walletID := uuid.New()
	repo := wallet_mocks.NewMockRepository(t)
	svc := service.New(repo)

	repo.EXPECT().
		ApplyRestriction(ctx, repository.RestrictionCommand{
			WalletID: walletID,
			OwnerID:  userID,
			ActorID:  userID,
			Action:   repository.ActionUserUnfreeze,
		}).
		Return(nil).
		Once()

	err := svc.UnfreezeOwnWallet(ctx, walletID, userID)

	require.NoError(t, err)
}

func TestRestrictionValidation(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name   string
		freeze bool
		userID uuid.UUID
		wallet uuid.UUID
		want   error
	}{
		{
			name:   "freeze missing user",
			freeze: true,
			userID: uuid.Nil,
			wallet: uuid.New(),
			want:   service.ErrInvalidUserID,
		},
		{
			name:   "freeze missing wallet",
			freeze: true,
			userID: uuid.New(),
			wallet: uuid.Nil,
			want:   service.ErrInvalidWalletID,
		},
		{
			name:   "unfreeze missing user",
			userID: uuid.Nil,
			wallet: uuid.New(),
			want:   service.ErrInvalidUserID,
		},
		{
			name:   "unfreeze missing wallet",
			userID: uuid.New(),
			wallet: uuid.Nil,
			want:   service.ErrInvalidWalletID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := wallet_mocks.NewMockRepository(t)
			svc := service.New(repo)

			var err error
			if tt.freeze {
				err = svc.FreezeOwnWallet(ctx, tt.wallet, tt.userID)
			} else {
				err = svc.UnfreezeOwnWallet(ctx, tt.wallet, tt.userID)
			}

			require.ErrorIs(t, err, tt.want)
		})
	}
}

func TestRestrictionErrorTranslation(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name string
		err  error
		want error
	}{
		{
			name: "wallet not found",
			err:  gorm.ErrRecordNotFound,
			want: service.ErrWalletNotFound,
		},
		{
			name: "transition conflict",
			err:  repository.ErrTransitionConflict,
			want: service.ErrWalletRestrictionConflict,
		},
		{
			name: "invalid actor",
			err:  repository.ErrInvalidActor,
			want: service.ErrWalletAccessDenied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID := uuid.New()
			walletID := uuid.New()
			repo := wallet_mocks.NewMockRepository(t)
			svc := service.New(repo)

			repo.EXPECT().
				ApplyRestriction(ctx, repository.RestrictionCommand{
					WalletID: walletID,
					OwnerID:  userID,
					ActorID:  userID,
					Action:   repository.ActionUserFreeze,
				}).
				Return(tt.err).
				Once()

			err := svc.FreezeOwnWallet(ctx, walletID, userID)

			require.ErrorIs(t, err, tt.want)
		})
	}
}

func TestRestrictionPreservesUnexpectedError(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	walletID := uuid.New()
	dbErr := errors.New("database unavailable")

	repo := wallet_mocks.NewMockRepository(t)
	svc := service.New(repo)

	repo.EXPECT().
		ApplyRestriction(ctx, repository.RestrictionCommand{
			WalletID: walletID,
			OwnerID:  userID,
			ActorID:  userID,
			Action:   repository.ActionUserFreeze,
		}).
		Return(dbErr).
		Once()

	err := svc.FreezeOwnWallet(ctx, walletID, userID)

	require.ErrorIs(t, err, dbErr)
}

func orgWallet(userID uuid.UUID, currency string) *model.Wallet {
	return &model.Wallet{
		UserID:   userID,
		Currency: currency,
		Status:   model.WalletStatusActive,
	}
}
