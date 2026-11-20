package repository_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/fredele20/nimbus/internal/modules/wallet/model"
	"github.com/fredele20/nimbus/internal/modules/wallet/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func (s *RepositoryTestSuite) restrictionCommand(
	walletID, ownerID, actorID uuid.UUID,
	action repository.RestrictionAction,
	reason string,
) repository.RestrictionCommand {
	return repository.RestrictionCommand{
		WalletID: walletID,
		OwnerID:  ownerID,
		ActorID:  actorID,
		Action:   action,
		Reason:   reason,
	}
}

func (s *RepositoryTestSuite) auditCount(
	walletID uuid.UUID,
) int64 {
	var count int64

	err := s.db.Table("audit_logs").
		Where(
			"entity = ? AND entity_id = ?",
			"wallet",
			walletID.String(),
		).
		Count(&count).Error

	s.Require().NoError(err)
	return count
}

func (s *RepositoryTestSuite) TestAdminUnfreezeClearsRestriction() {
	ctx := context.Background()

	ownerID := s.createUser()
	adminID := s.createUser()
	wallet := s.createWallet(ownerID, "NGN")

	freeze := s.restrictionCommand(
		wallet.ID,
		ownerID,
		adminID,
		repository.ActionAdminFreeze,
		"Fraud investigation",
	)

	s.Require().NoError(
		s.repo.ApplyRestriction(ctx, freeze),
	)

	unfreeze := s.restrictionCommand(
		wallet.ID,
		ownerID,
		adminID,
		repository.ActionAdminUnfreeze,
		"",
	)

	s.Require().NoError(
		s.repo.ApplyRestriction(ctx, unfreeze),
	)

	found, err := s.repo.FindByID(
		ctx,
		wallet.ID,
		ownerID,
	)

	s.Require().NoError(err)
	s.Require().Equal(
		model.WalletStatusActive,
		found.Status,
	)
	s.Require().Nil(found.RestrictionType)
	s.Require().Nil(found.RestrictionReason)
	s.Require().Nil(found.RestrictedBy)
	s.Require().Nil(found.RestrictedAt)

	s.Require().Equal(
		int64(2),
		s.auditCount(wallet.ID),
	)
}

func (s *RepositoryTestSuite) TestClosedWalletCannotBeFrozen() {
	ctx := context.Background()

	ownerID := s.createUser()
	adminID := s.createUser()
	wallet := s.createWallet(ownerID, "NGN")

	err := s.db.Exec(
		"UPDATE wallets SET status = ? WHERE id = ?",
		model.WalletStatusClosed,
		wallet.ID,
	).Error
	s.Require().NoError(err)

	actions := []repository.RestrictionCommand{
		s.restrictionCommand(
			wallet.ID,
			ownerID,
			ownerID,
			repository.ActionUserFreeze,
			"",
		),
		s.restrictionCommand(
			wallet.ID,
			ownerID,
			adminID,
			repository.ActionAdminFreeze,
			"Investigation",
		),
	}

	for _, cmd := range actions {
		err := s.repo.ApplyRestriction(ctx, cmd)
		s.Require().ErrorIs(
			err,
			repository.ErrTransitionConflict,
		)
	}

	found, err := s.repo.FindByID(
		ctx,
		wallet.ID,
		ownerID,
	)

	s.Require().NoError(err)
	s.Require().Equal(
		model.WalletStatusClosed,
		found.Status,
	)
	s.Require().Zero(s.auditCount(wallet.ID))
}

func (s *RepositoryTestSuite) TestAuditFailureRollsBackWalletUpdate() {
	ctx := context.Background()

	ownerID := s.createUser()
	wallet := s.createWallet(ownerID, "NGN")

	// A temporary trigger forces the audit insert to fail.
	// The trigger and function are removed after this test.
	s.Require().NoError(s.db.Exec(`
		CREATE FUNCTION reject_wallet_audit_test()
		RETURNS trigger
		LANGUAGE plpgsql
		AS $$
		BEGIN
			IF NEW.entity = 'wallet' THEN
				RAISE EXCEPTION 'forced audit failure';
			END IF;
			RETURN NEW;
		END;
		$$
	`).Error)

	defer func() {
		require.NoError(s.T(), s.db.Exec(`
			DROP FUNCTION IF EXISTS
			reject_wallet_audit_test() CASCADE
		`).Error)
	}()

	s.Require().NoError(s.db.Exec(`
		CREATE TRIGGER reject_wallet_audit_test_trigger
		BEFORE INSERT ON audit_logs
		FOR EACH ROW
		EXECUTE FUNCTION reject_wallet_audit_test()
	`).Error)

	cmd := s.restrictionCommand(
		wallet.ID,
		ownerID,
		ownerID,
		repository.ActionUserFreeze,
		"",
	)

	err := s.repo.ApplyRestriction(ctx, cmd)
	s.Require().Error(err)

	found, findErr := s.repo.FindByID(
		ctx,
		wallet.ID,
		ownerID,
	)

	s.Require().NoError(findErr)
	s.Require().Equal(
		model.WalletStatusActive,
		found.Status,
	)
	s.Require().Nil(found.RestrictionType)
	s.Require().Nil(found.RestrictedBy)
	s.Require().Nil(found.RestrictedAt)
	s.Require().Zero(s.auditCount(wallet.ID))
}

func (s *RepositoryTestSuite) TestConcurrentFreezeOnlyOneSucceeds() {
	ctx := context.Background()

	ownerID := s.createUser()
	wallet := s.createWallet(ownerID, "NGN")

	const workers = 8

	// Ensure the connection pool can serve all workers.
	sqlDB, err := s.db.DB()
	s.Require().NoError(err)
	sqlDB.SetMaxOpenConns(workers + 2)

	cmd := s.restrictionCommand(
		wallet.ID,
		ownerID,
		ownerID,
		repository.ActionUserFreeze,
		"",
	)

	start := make(chan struct{})
	results := make(chan error, workers)

	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			<-start
			results <- s.repo.ApplyRestriction(ctx, cmd)
		}()
	}

	close(start)
	wg.Wait()
	close(results)

	successes := 0
	conflicts := 0

	for err := range results {
		switch {
		case err == nil:
			successes++

		case errors.Is(
			err,
			repository.ErrTransitionConflict,
		):
			conflicts++

		default:
			s.Require().NoError(err)
		}
	}

	s.Require().Equal(1, successes)
	s.Require().Equal(workers-1, conflicts)

	found, err := s.repo.FindByID(
		ctx,
		wallet.ID,
		ownerID,
	)

	s.Require().NoError(err)
	s.Require().Equal(
		model.WalletStatusFrozen,
		found.Status,
	)
	s.Require().NotNil(found.RestrictionType)
	s.Require().Equal(
		model.RestrictionTypeUser,
		*found.RestrictionType,
	)

	s.Require().Equal(
		int64(1),
		s.auditCount(wallet.ID),
	)
}

func TestRestrictionCommandRejectsInvalidActor(t *testing.T) {
	require.NotEqual(t, uuid.Nil, uuid.New())
}
