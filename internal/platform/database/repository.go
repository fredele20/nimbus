package database

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository[M any] struct {
	DB *gorm.DB
}

func NewRepository[M any](db *gorm.DB) *Repository[M] {
	return &Repository[M]{
		DB: db,
	}
}

func (r *Repository[M]) Create(ctx context.Context, data *M) error {
	err := r.DB.WithContext(ctx).Create(data).Error
	return CheckUniqueConstraintErr(err)
}

func (r *Repository[M]) CreateMany(ctx context.Context, data ...*M) error {
	err := r.DB.WithContext(ctx).Create(data).Error
	return CheckUniqueConstraintErr(err)
}

func (r *Repository[M]) UpdateOne(ctx context.Context, data *M) error {
	err := r.DB.WithContext(ctx).Model(data).Updates(data).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository[M]) UpdateOneByID(ctx context.Context, id uuid.UUID, data *M) error {
	err := r.DB.WithContext(ctx).Model(new(M)).Where("id = ?", id).Updates(data).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository[M]) UpdateMany(ctx context.Context, data *M, query any, args ...any) error {
	err := r.DB.WithContext(ctx).Model(new(M)).Where(query, args...).Updates(data).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository[M]) FindOne(ctx context.Context, query any, args ...any) (M, error) {
	data := new(M)
	q := r.DB.WithContext(ctx).Where(query, args...)
	err := q.First(data).Error
	if err != nil {
		return *data, err
	}

	return *data, nil
}

func (r *Repository[M]) FindOneLatest(ctx context.Context, query any, args ...any) (M, error) {
	data := new(M)
	q := r.DB.WithContext(ctx).Where(query, args...).Order("created_at DESC")
	err := q.First(data).Error
	if err != nil {
		return *data, err
	}

	return *data, nil
}

func (r *Repository[M]) FindOneWithStatus(ctx context.Context, query any, args ...any) (M, bool, error) {
	m, err := r.FindOne(ctx, query, args...)
	found := true

	if err != nil {
		found = false
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = nil
		}
	}

	return m, found, err
}

func (r *Repository[M]) FindOneByID(ctx context.Context, id uuid.UUID) (M, error) {
	data := new(M)
	q := r.DB.WithContext(ctx).Where("id = ?", id)
	err := q.First(data).Error
	if err != nil {
		return *data, err
	}

	return *data, nil
}

func (r *Repository[M]) FindMany(ctx context.Context, query any, args ...any) ([]M, error) {
	return r.FindManyWithLimit(ctx, -1, -1, query, args...)
}

func (r *Repository[M]) FindAll(ctx context.Context) ([]M, error) {
	return r.FindManyWithLimit(ctx, -1, -1, nil)
}

func (r *Repository[M]) FindManyWithLimit(ctx context.Context, limit int, offset int, query any, args ...any) ([]M, error) {
	data := new([]M)
	q := r.DB.WithContext(ctx).Where(query, args...).Limit(limit).Offset(offset)
	err := q.Find(data).Error
	if err != nil {
		return nil, err
	}

	return *data, nil
}

func (r *Repository[M]) DeleteOne(ctx context.Context, query any, args ...any) error {
	data := new(M)
	err := r.DB.WithContext(ctx).Where(query, args...).Delete(data).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository[M]) DeleteOneById(ctx context.Context, id uuid.UUID) error {
	data := new(M)
	err := r.DB.WithContext(ctx).Where("id = ?", id).Delete(data).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository[M]) DeleteMany(ctx context.Context, query any, args ...any) error {
	data := new(M)
	err := r.DB.WithContext(ctx).Where(query, args...).Delete(data).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository[M]) WithTrx(trx *gorm.DB) *Repository[M] {
	return &Repository[M]{DB: trx}
}

func (r *Repository[M]) Count(ctx context.Context, query any, args ...any) (i int64, err error) {
	data := new(M)
	err = r.DB.WithContext(ctx).Where(query, args...).Model(data).Count(&i).Error
	return
}
