package service

import (
	"context"
	"errors"

	"github.com/fredele20/nimbus/internal/modules/auth/password"
	"github.com/fredele20/nimbus/internal/modules/user/dto"
	"github.com/fredele20/nimbus/internal/modules/user/model"
	"github.com/fredele20/nimbus/internal/modules/user/repository"
)

var (
	ErrEmailAlreadyExist = errors.New("email already exists")
)

type Service interface {
	Register(ctx context.Context, req dto.RegisterRequest) (*model.User, error)
}

type service struct {
	repo repository.Repository
}

func New(repo repository.Repository) Service {
	return &service{repo: repo}
}

func (s *service) Register(ctx context.Context, req dto.RegisterRequest) (*model.User, error) {
	existingUser, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	if existingUser != nil {
		return nil, ErrEmailAlreadyExist
	}

	hash, err := password.Hash(req.Password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        req.Email,
		PasswordHash: hash,
		Status:       model.StatusActive,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}
