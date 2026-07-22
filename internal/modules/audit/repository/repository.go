package repository

import (
	"context"

	"github.com/fredele20/nimbus/internal/modules/audit/model"
	"gorm.io/gorm"
)

type repository struct {
	db *gorm.DB
}

type Repository interface {
	Create(ctx context.Context, log *model.AuditLog) error
}

func New(db *gorm.DB) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) Create(ctx context.Context, log *model.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}
