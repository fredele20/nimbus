package repository

import (
	"context"

	"github.com/fredele20/nimbus/internal/modules/auth/model"
	"github.com/fredele20/nimbus/internal/platform/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *model.RefreshToken) error
	FindByToken(ctx context.Context, tokenHash string) (model.RefreshToken, error)
	DeleteOneById(ctx context.Context, id uuid.UUID) error
	UpdateOne(ctx context.Context, token *model.RefreshToken) error
}

type refreshTokenRepository struct {
	*database.Repository[model.RefreshToken]
}

func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepository{
		Repository: database.NewRepository[model.RefreshToken](db),
	}
}

func (r *refreshTokenRepository) FindByToken(ctx context.Context, storedToken string) (model.RefreshToken, error) {
	token, err := r.FindOne(ctx, "token = ?", storedToken)
	return token, err
}

func (r *refreshTokenRepository) Update(ctx context.Context, token *model.RefreshToken) error {
	return r.UpdateOneByID(ctx, token.ID, token)
}
