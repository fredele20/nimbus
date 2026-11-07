package repository

import (
	"context"

	"github.com/fredele20/nimbus/internal/modules/user/model"
	"github.com/fredele20/nimbus/internal/platform/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, user *model.User) error
	FindByEmail(ctx context.Context, email string) (model.User, error)
	FindOneByID(ctx context.Context, id uuid.UUID) (model.User, error)
	FindAll(ctx context.Context) ([]model.User, error)
	Update(ctx context.Context, user *model.User) error
}

type repository struct {
	*database.Repository[model.User]
}

func New(db *gorm.DB) Repository {
	return &repository{
		Repository: database.NewRepository[model.User](db),
	}
}


func (r *repository) FindByEmail(ctx context.Context, email string) (model.User, error) {
	user, err := r.FindOne(ctx, "email = ?", email) // Inherited directly from embedded repo!
	return user, err
}


func (r *repository) Update(ctx context.Context, user *model.User) error {
	return r.UpdateOneByID(ctx, user.ID, user) // Inherited directly!
}
