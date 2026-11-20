package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	auditModel "github.com/fredele20/nimbus/internal/modules/audit/model"
	"github.com/fredele20/nimbus/internal/modules/wallet/model"
	"github.com/fredele20/nimbus/internal/platform/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrInvalidAction      = errors.New("invalid wallet restriction action")
	ErrTransitionConflict = errors.New("wallet restriction transition conflict")
	ErrInvalidActor       = errors.New("invalid wallet restriction actor")
)

type RestrictionAction string

const (
	ActionUserFreeze    RestrictionAction = "user_freeze"
	ActionUserUnfreeze  RestrictionAction = "user_unfreeze"
	ActionAdminFreeze   RestrictionAction = "admin_freeze"
	ActionAdminUnfreeze RestrictionAction = "admin_unfreeze"
)

type RestrictionCommand struct {
	WalletID uuid.UUID
	OwnerID  uuid.UUID
	ActorID  uuid.UUID
	Action   RestrictionAction
	Reason   string
}

type Repository interface {
	Create(ctx context.Context, wallet *model.Wallet) error
	FindByID(ctx context.Context, walletID, userID uuid.UUID) (model.Wallet, error)
	FindByWalletID(ctx context.Context, walletID uuid.UUID) (model.Wallet, error)
	FindByUserAndCurrency(ctx context.Context, userID uuid.UUID, currency string) (model.Wallet, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]model.Wallet, error)
	ApplyRestriction(ctx context.Context, restrictionCommand RestrictionCommand) error
}

type repository struct {
	*database.Repository[model.Wallet]
	db *gorm.DB
}

func New(db *gorm.DB) Repository {
	return &repository{
		Repository: database.NewRepository[model.Wallet](db),
		db:         db,
	}
}

func (r *repository) FindByID(ctx context.Context, walletID, userID uuid.UUID) (model.Wallet, error) {
	return r.FindOne(ctx, "id = ? AND user_id = ?", walletID, userID)
}

func (r *repository) FindByWalletID(ctx context.Context, walletID uuid.UUID) (model.Wallet, error) {
	return r.FindOne(ctx, "id = ?", walletID)
}

func (r *repository) FindByUserAndCurrency(ctx context.Context, userID uuid.UUID, currency string) (model.Wallet, error) {
	return r.FindOne(ctx, "user_id = ? AND currency = ?", userID, currency)
}

func (r *repository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]model.Wallet, error) {
	var wallets []model.Wallet

	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&wallets).Error

	if err != nil {
		return nil, err
	}

	return wallets, nil
}

func (r *repository) ApplyRestriction(ctx context.Context, restrictionCommand RestrictionCommand) error {
	if restrictionCommand.WalletID == uuid.Nil || restrictionCommand.OwnerID == uuid.Nil || restrictionCommand.ActorID == uuid.Nil {
		return ErrInvalidActor
	}

	switch restrictionCommand.Action {
	case ActionUserFreeze, ActionUserUnfreeze:
		if restrictionCommand.ActorID != restrictionCommand.OwnerID {
			return ErrInvalidActor
		}
	case ActionAdminFreeze, ActionAdminUnfreeze:
		// The service must verify the actor's admin permission
	default:
		return ErrInvalidAction
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		query, args, err := restrictionUpdate(restrictionCommand)
		if err != nil {
			return err
		}

		result := tx.Exec(query, args...)
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected != 1 {
			var count int64
			err := tx.Model(&model.Wallet{}).
				Where("id = ? AND user_id = ?", restrictionCommand.WalletID, restrictionCommand.OwnerID).
				Count(&count).Error

			if err != nil {
				return err
			}
			if count == 0 {
				return gorm.ErrRecordNotFound
			}
			return ErrTransitionConflict
		}

		metadata, err := json.Marshal(map[string]any{
			"restriction_action": restrictionCommand.Action,
			"reason":             strings.TrimSpace(restrictionCommand.Reason),
		})
		if err != nil {
			return fmt.Errorf("marshal wallet audit metadata: %w", err)
		}

		audit := auditModel.AuditLog{
			ActorID:  restrictionCommand.ActorID.String(),
			UserID:   restrictionCommand.OwnerID.String(),
			Action:   string(restrictionCommand.Action),
			Entity:   "wallet",
			EntityID: restrictionCommand.WalletID.String(),
			Metadata: metadata,
		}
		if err := tx.Create(&audit).Error; err != nil {
			return fmt.Errorf("create wallet restriction audit: %w", err)
		}
		return nil
	})
}

func restrictionUpdate(restrictionCmd RestrictionCommand) (string, []any, error) {
	base := `
		UPDATE wallets
		SET status = ?,
			restriction_type = ?,
			restriction_reason = ?,
			restricted_by = ?,
			restricted_at = NOW(),
			updated_at = NOW()
		WHERE id = ?
			AND user_id = ?
			AND deleted_at IS NULL
	`

	switch restrictionCmd.Action {
	case ActionUserFreeze:
		return base + `
			AND status = 'active'
			AND restriction_type IS NULL
		`, []any{
				model.WalletStatusFrozen,
				model.RestrictionTypeUser,
				nil,
				restrictionCmd.ActorID,
				restrictionCmd.WalletID,
				restrictionCmd.OwnerID,
			}, nil
	case ActionAdminFreeze:
		reason := strings.TrimSpace(restrictionCmd.Reason)
		if reason == "" {
			return "", nil, errors.New("admin restriction reason is required")
		}

		return base + `
				AND (
					status = 'active'
					OR (
						status = 'frozen'
						AND restriction_type = 'user'
					)
				)
			`, []any{
				model.WalletStatusFrozen,
				model.RestrictionTypeAdmin,
				reason,
				restrictionCmd.ActorID,
				restrictionCmd.WalletID,
				restrictionCmd.OwnerID,
			}, nil

	case ActionUserUnfreeze, ActionAdminUnfreeze:
		restrictionType := model.RestrictionTypeUser
		if restrictionCmd.Action == ActionAdminUnfreeze {
			restrictionType = model.RestrictionTypeAdmin
		}

		return `
				UPDATE wallets
				SET status = 'active',
					restriction_type = NULL,
					restriction_reason = NULL,
					restricted_by = NULL,
					restricted_at = NULL,
					updated_at = NOW()
				WHERE id = ?
					AND user_id = ?
					AND deleted_at IS NULL
					AND status = 'frozen'
					AND restriction_type = ?
			`, []any{
				restrictionCmd.WalletID,
				restrictionCmd.OwnerID,
				restrictionType,
			}, nil

	default:
		return "", nil, ErrInvalidAction
	}
}
