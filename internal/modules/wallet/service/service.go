package service

import (
	"context"
	"errors"
	"strings"

	"github.com/fredele20/nimbus/internal/modules/wallet/dto"
	"github.com/fredele20/nimbus/internal/modules/wallet/model"
	"github.com/fredele20/nimbus/internal/modules/wallet/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type Service interface {
	Create(ctx context.Context, userID uuid.UUID, currency string) (dto.WalletResponse, error)

	GetByID(ctx context.Context, walletID, userID uuid.UUID) (dto.WalletResponse, error)

	GetByCurrency(ctx context.Context, userID uuid.UUID, currency string) (dto.WalletResponse, error)

	ListByUser(ctx context.Context, userID uuid.UUID) ([]dto.WalletResponse, error)

	FreezeOwnWallet(ctx context.Context, walletID, userID uuid.UUID) error

	UnfreezeOwnWallet(ctx context.Context, walletID, userID uuid.UUID) error
}

type service struct {
	repo repository.Repository
}

func New(repo repository.Repository) Service {
	return &service{
		repo: repo,
	}
}

var supportedCurrencies = map[string]struct{}{
	"NGN": {},
	"USD": {},
	"GBP": {},
	"EUR": {},
}

func (s *service) Create(ctx context.Context, userID uuid.UUID, currency string) (dto.WalletResponse, error) {
	var walletResponse dto.WalletResponse

	if err := validateUserID(userID); err != nil {
		return walletResponse, err
	}

	currency, err := validateCurrency(currency)
	if err != nil {
		return walletResponse, err
	}

	wallet := &model.Wallet{
		UserID:   userID,
		Currency: currency,
		Status:   model.WalletStatusActive,
	}

	if err := s.repo.Create(ctx, wallet); err != nil {
		return walletResponse, translateRepositoryError(err)
	}

	return dto.FromWallet(*wallet), nil
}

func (s *service) GetByID(ctx context.Context, walletID, userID uuid.UUID) (dto.WalletResponse, error) {
	var walletResponse dto.WalletResponse

	if err := validateUserID(userID); err != nil {
		return walletResponse, err
	}

	if err := validateWalletID(walletID); err != nil {
		return walletResponse, err
	}

	wallet, err := s.repo.FindByID(ctx, walletID, userID)
	if err != nil {
		return walletResponse, translateRepositoryError(err)
	}

	return dto.FromWallet(wallet), nil
}

func (s *service) GetByCurrency(ctx context.Context, userID uuid.UUID, currency string) (dto.WalletResponse, error) {
	var walletResponse dto.WalletResponse

	if err := validateUserID(userID); err != nil {
		return walletResponse, err
	}

	currency, err := validateCurrency(currency)
	if err != nil {
		return walletResponse, err
	}

	wallet, err := s.repo.FindByUserAndCurrency(ctx, userID, currency)
	if err != nil {
		return walletResponse, translateRepositoryError(err)
	}

	return dto.FromWallet(wallet), nil
}

func (s *service) ListByUser(ctx context.Context, userID uuid.UUID,
) ([]dto.WalletResponse, error) {
	if err := validateUserID(userID); err != nil {
		return nil, err
	}

	wallets, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return dto.FromWallets(wallets), nil
}

func (s *service) FreezeOwnWallet(ctx context.Context, walletID, userID uuid.UUID) error {
	if err := validateUserID(userID); err != nil {
		return err
	}

	if err := validateWalletID(walletID); err != nil {
		return err
	}

	return translateRepositoryError(
		s.repo.ApplyRestriction(ctx, repository.RestrictionCommand{
			WalletID: walletID,
			OwnerID:  userID,
			ActorID:  userID,
			Action:   repository.ActionUserFreeze,
		}),
	)
}

func (s *service) UnfreezeOwnWallet(ctx context.Context, walletID, userID uuid.UUID) error {
	if err := validateUserID(userID); err != nil {
		return err
	}

	if err := validateWalletID(walletID); err != nil {
		return err
	}

	return translateRepositoryError(
		s.repo.ApplyRestriction(ctx, repository.RestrictionCommand{
			WalletID: walletID,
			OwnerID:  userID,
			ActorID:  userID,
			Action:   repository.ActionUserUnfreeze,
		}),
	)
}

func validateCurrency(currency string) (string, error) {
	currency = strings.ToUpper(strings.TrimSpace(currency))

	if _, ok := supportedCurrencies[currency]; !ok {
		return "", ErrInvalidCurrency
	}

	return currency, nil
}

func validateUserID(userID uuid.UUID) error {
	if userID == uuid.Nil {
		return ErrInvalidUserID
	}

	return nil
}

func validateWalletID(walletID uuid.UUID) error {
	if walletID == uuid.Nil {
		return ErrInvalidWalletID
	}

	return nil
}

func translateRepositoryError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrWalletNotFound
	}

	if errors.Is(err, repository.ErrTransitionConflict) {
		return ErrWalletRestrictionConflict
	}

	if errors.Is(err, repository.ErrInvalidActor) {
		return ErrWalletAccessDenied
	}

	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) &&
		pgErr.Code == "23505" &&
		pgErr.ConstraintName == "uq_wallets_user_currency" {
		return ErrWalletAlreadyExists
	}

	return err
}
