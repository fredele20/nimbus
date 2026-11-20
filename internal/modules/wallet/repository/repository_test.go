package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/fredele20/nimbus/internal/modules/wallet/model"
	"github.com/fredele20/nimbus/internal/modules/wallet/repository"
	"github.com/fredele20/nimbus/internal/platform/testdb"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type RepositoryTestSuite struct {
	suite.Suite

	db   *gorm.DB
	repo repository.Repository
}

func (s *RepositoryTestSuite) SetupSuite() {
	s.db = testdb.NewPostgres(s.T())
	s.repo = repository.New(s.db)
}

func (s *RepositoryTestSuite) SetupTest() {
	err := s.db.Exec("TRUNCATE TABLE wallets, users CASCADE").Error
	s.Require().NoError(err)
}

func TestRepositorySuite(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}

func (s *RepositoryTestSuite) createUser() uuid.UUID {
	userID := uuid.New()

	err := s.db.Exec(`
		INSERT INTO users (
			id, first_name, last_name, email, password
		)
		VALUES (?, ?, ?, ?, ?)
	`,
		userID,
		"Test",
		"User",
		userID.String()+"@example.com",
		"hashed-password",
	).Error

	s.Require().NoError(err)

	return userID
}

func (s *RepositoryTestSuite) createWallet(userID uuid.UUID, currency string) model.Wallet {
	wallet := model.Wallet{
		UserID:   userID,
		Currency: currency,
	}

	err := s.repo.Create(context.Background(), &wallet)

	s.Require().NoError(err)
	s.Require().NotEqual(uuid.Nil, wallet.ID)

	return wallet
}

func (s *RepositoryTestSuite) TestCreateAndFindByID() {
	userID := s.createUser()
	wallet := s.createWallet(userID, "NGN")

	found, err := s.repo.FindByID(context.Background(), wallet.ID, userID)

	s.Require().NoError(err)
	s.Require().Equal(wallet.ID, found.ID)
	s.Require().Equal(userID, found.UserID)
	s.Require().Equal("NGN", found.Currency)
	s.Require().Equal(model.WalletStatusActive, found.Status)
}

func (s *RepositoryTestSuite) TestFindByUserAndCurrency() {
	userID := s.createUser()
	wallet := s.createWallet(userID, "NGN")

	found, err := s.repo.FindByUserAndCurrency(
		context.Background(),
		userID,
		"NGN",
	)

	s.Require().NoError(err)
	s.Require().Equal(wallet.ID, found.ID)
}

func (s *RepositoryTestSuite) TestFindByUserID() {
	userID := s.createUser()

	s.createWallet(userID, "NGN")
	s.createWallet(userID, "USD")

	wallets, err := s.repo.FindByUserID(context.Background(), userID)

	s.Require().NoError(err)
	s.Require().Len(wallets, 2)
}

func (s *RepositoryTestSuite) TestFindByIDEnforcesOwnership() {
	userID := s.createUser()
	anotherUserID := s.createUser()

	wallet := s.createWallet(userID, "NGN")

	_, err := s.repo.FindByID(
		context.Background(),
		wallet.ID,
		anotherUserID,
	)

	s.Require().ErrorIs(err, gorm.ErrRecordNotFound)
}

func (s *RepositoryTestSuite) TestUserFreezeAndUnfreeze() {
	ctx := context.Background()
	userID := s.createUser()
	wallet := s.createWallet(userID, "NGN")

	freeze := repository.RestrictionCommand{
		WalletID: wallet.ID,
		OwnerID:  userID,
		ActorID:  userID,
		Action:   repository.ActionUserFreeze,
	}

	s.Require().NoError(s.repo.ApplyRestriction(ctx, freeze))

	found, err := s.repo.FindByID(ctx, wallet.ID, userID)
	s.Require().NoError(err)
	s.Require().Equal(model.WalletStatusFrozen, found.Status)
	s.Require().NotNil(found.RestrictionType)
	s.Require().Equal(model.RestrictionTypeUser, *found.RestrictionType)
	s.Require().NotNil(found.RestrictedBy)
	s.Require().Equal(userID, *found.RestrictedBy)
	s.Require().NotNil(found.RestrictedAt)

	unfreeze := freeze
	unfreeze.Action = repository.ActionUserUnfreeze

	s.Require().NoError(s.repo.ApplyRestriction(ctx, unfreeze))

	found, err = s.repo.FindByID(ctx, wallet.ID, userID)
	s.Require().NoError(err)
	s.Require().Equal(model.WalletStatusActive, found.Status)
	s.Require().Nil(found.RestrictionType)
	s.Require().Nil(found.RestrictedBy)
	s.Require().Nil(found.RestrictedAt)
}

func (s *RepositoryTestSuite) TestAdminRestrictionOverridesUserFreeze() {
	ctx := context.Background()
	ownerID := s.createUser()
	adminID := s.createUser()
	wallet := s.createWallet(ownerID, "NGN")

	s.Require().NoError(s.repo.ApplyRestriction(ctx,
		repository.RestrictionCommand{
			WalletID: wallet.ID,
			OwnerID:  ownerID,
			ActorID:  ownerID,
			Action:   repository.ActionUserFreeze,
		},
	))

	s.Require().NoError(s.repo.ApplyRestriction(ctx,
		repository.RestrictionCommand{
			WalletID: wallet.ID,
			OwnerID:  ownerID,
			ActorID:  adminID,
			Action:   repository.ActionAdminFreeze,
			Reason:   "Suspicious activity",
		},
	))

	err := s.repo.ApplyRestriction(ctx,
		repository.RestrictionCommand{
			WalletID: wallet.ID,
			OwnerID:  ownerID,
			ActorID:  ownerID,
			Action:   repository.ActionUserUnfreeze,
		},
	)

	s.Require().ErrorIs(err, repository.ErrTransitionConflict)

	found, err := s.repo.FindByID(ctx, wallet.ID, ownerID)
	s.Require().NoError(err)
	s.Require().Equal(model.WalletStatusFrozen, found.Status)
	s.Require().NotNil(found.RestrictionType)
	s.Require().Equal(model.RestrictionTypeAdmin, *found.RestrictionType)
}

func (s *RepositoryTestSuite) TestUserCannotFreezeAnotherUsersWallet() {
	ctx := context.Background()
	ownerID := s.createUser()
	otherID := s.createUser()
	wallet := s.createWallet(ownerID, "NGN")

	err := s.repo.ApplyRestriction(ctx,
		repository.RestrictionCommand{
			WalletID: wallet.ID,
			OwnerID:  ownerID,
			ActorID:  otherID,
			Action:   repository.ActionUserFreeze,
		},
	)

	s.Require().ErrorIs(err, repository.ErrInvalidActor)
}

func (s *RepositoryTestSuite) TestCannotFreezeUnknownWallet() {
	ctx := context.Background()
	userID := s.createUser()

	err := s.repo.ApplyRestriction(ctx,
		repository.RestrictionCommand{
			WalletID: uuid.New(),
			OwnerID:  userID,
			ActorID:  userID,
			Action:   repository.ActionUserFreeze,
		},
	)

	s.Require().ErrorIs(err, gorm.ErrRecordNotFound)
}

func (s *RepositoryTestSuite) TestCannotFreezeAlreadyFrozenWallet() {
	ctx := context.Background()
	userID := s.createUser()
	wallet := s.createWallet(userID, "NGN")

	cmd := repository.RestrictionCommand{
		WalletID: wallet.ID,
		OwnerID:  userID,
		ActorID:  userID,
		Action:   repository.ActionUserFreeze,
	}

	s.Require().NoError(s.repo.ApplyRestriction(ctx, cmd))
	s.Require().ErrorIs(
		s.repo.ApplyRestriction(ctx, cmd),
		repository.ErrTransitionConflict,
	)
}

func (s *RepositoryTestSuite) TestDuplicateUserCurrencyIsRejected() {
	userID := s.createUser()

	s.createWallet(userID, "NGN")

	duplicate := model.Wallet{
		UserID:   userID,
		Currency: "NGN",
	}

	err := s.repo.Create(context.Background(), &duplicate)

	s.Require().Error(err)
}

func (s *RepositoryTestSuite) TestDatabaseRejectsInvalidCurrency() {
	userID := s.createUser()

	wallet := model.Wallet{
		UserID:   userID,
		Currency: "INVALID",
	}

	err := s.repo.Create(context.Background(), &wallet)

	s.Require().Error(err)
}

func (s *RepositoryTestSuite) TestDatabaseRejectsInvalidStatus() {
	userID := s.createUser()

	wallet := model.Wallet{
		UserID:   userID,
		Currency: "NGN",
		Status:   model.WalletStatus("invalid"),
	}

	err := s.repo.Create(context.Background(), &wallet)

	s.Require().Error(err)
}

func (s *RepositoryTestSuite) TestFindByIDReturnsNotFoundForUnknownWallet() {
	_, err := s.repo.FindByID(
		context.Background(),
		uuid.New(),
		uuid.New(),
	)

	s.Require().True(errors.Is(err, gorm.ErrRecordNotFound))
}
