package service

import (
	"context"
	"log/slog"

	"github.com/fredele20/nimbus/internal/modules/audit"
	auditService "github.com/fredele20/nimbus/internal/modules/audit/service"
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
	repo         repository.Repository
	auditService auditService.Service
	logger       *slog.Logger
}

func New(
	repo repository.Repository,
	auditService auditService.Service,
	logger *slog.Logger,
) Service {
	return &service{
		repo:         repo,
		auditService: auditService,
		logger:       logger,
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
	action := getActionFromStatus(status)
	if auditErr := s.auditService.Log(ctx, userID, userID, action, "user", userID, map[string]any{"user": user.Email}); auditErr != nil {
		s.logger.Error(
			"failed to write audit log",
			"action", action,
			"user_id", user.ID,
			"error", err,
		)
	}

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

func getActionFromStatus(status model.Status) string {
	var action string
	switch status {
	case model.StatusActive:
		action = audit.ActionUserActivated
	case model.StatusSuspended:
		action = audit.ActionUserSuspended
	case model.StatusDeleted:
		action = audit.ActionUserDeleted
	default:
		return ""
	}
	return action
}
