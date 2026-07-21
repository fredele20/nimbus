package service

import (
	"context"

	"github.com/fredele20/nimbus/internal/modules/user/dto"
)

type Service interface {
	GetProfile(ctx context.Context, userID string) (*dto.ProfileResponse, error)
	UpdateProfile(ctx context.Context, userID string, req dto.UpdateProfileRequest) error
	ListUsers(ctx context.Context) ([]dto.ProfileResponse, error)
	UpdateStatus(ctx context.Context, userID string, status string) error
}
