package service

import (
	"context"

	"github.com/fredele20/nimbus/internal/modules/user/dto"
	"github.com/fredele20/nimbus/internal/modules/user/model"
	"github.com/fredele20/nimbus/internal/modules/user/repository"
)

type Service interface {
	GetProfile(ctx context.Context, userID string) (*dto.ProfileResponse, error)
	UpdateProfile(ctx context.Context, userID string, req dto.UpdateProfileRequest) error
	ListUsers(ctx context.Context) ([]dto.ProfileResponse, error)
	UpdateStatus(ctx context.Context, userID string, status model.Status) error
}

type service struct {
	repo repository.Repository
}

func New(repo repository.Repository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) GetProfile(ctx context.Context, userID string) (*dto.ProfileResponse, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, ErrUserNotFound
	}

	return mapProfile(user), nil
}

func (s *service) UpdateProfile(ctx context.Context, userID string, req dto.UpdateProfileRequest) error {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	if user == nil {
		return ErrUserNotFound
	}

	user.FirstName = req.FirstName
	user.LastName = req.LastName

	return s.repo.Update(ctx, user)
}

func (s *service) ListUsers(ctx context.Context) ([]dto.ProfileResponse, error) {
	users, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.ProfileResponse, 0, len(users))

	for _, user := range users {
		responses = append(responses, *mapProfile(&user))
	}

	return responses, nil
}

func (s *service) UpdateStatus(ctx context.Context, userID string, status model.Status) error {
	switch status {
	case model.StatusActive,
		model.StatusSuspended,
		model.StatusDeleted:
	default:
		return ErrInvalidStatus
	}

	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	if user == nil {
		return ErrUserNotFound
	}

	user.Status = status

	return s.repo.Update(ctx, user)
}

func mapProfile(user *model.User) *dto.ProfileResponse {
	return &dto.ProfileResponse{
		ID:        user.ID.String(),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Role:      string(user.Role),
		Status:    string(user.Status),
	}
}
