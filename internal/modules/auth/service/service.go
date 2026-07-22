package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/fredele20/nimbus/internal/modules/audit"
	auditService "github.com/fredele20/nimbus/internal/modules/audit/service"
	"github.com/fredele20/nimbus/internal/modules/auth/dto"
	authmodel "github.com/fredele20/nimbus/internal/modules/auth/model"
	"github.com/fredele20/nimbus/internal/modules/auth/password"
	authRepo "github.com/fredele20/nimbus/internal/modules/auth/repository"
	"github.com/fredele20/nimbus/internal/modules/auth/token"
	"github.com/fredele20/nimbus/internal/modules/user/model"
	userRepo "github.com/fredele20/nimbus/internal/modules/user/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailExists        = errors.New("email already exists")
)

type Service interface {
	Register(context.Context, dto.RegisterRequest) (*dto.RegisterResponse, error)
	Login(context.Context, dto.LoginRequest) (*dto.LoginResponse, error)
	Refresh(ctx context.Context, req dto.RefreshRequest) (*dto.RefreshResponse, error)
	Logout(ctx context.Context, refreshToken string) error
}

type service struct {
	userRepo         userRepo.Repository
	passwordHasher   password.PasswordHasher
	refreshTokenRepo authRepo.RefreshTokenRepository
	auditService     auditService.Service
	token            token.TokenManager
	logger           *slog.Logger
}

func New(
	userRepo userRepo.Repository,
	passwordHasher password.PasswordHasher,
	refreshTokenRepo authRepo.RefreshTokenRepository,
	auditService auditService.Service,
	token token.TokenManager,
	logger *slog.Logger,
) Service {
	return &service{
		userRepo:         userRepo,
		passwordHasher:   passwordHasher,
		refreshTokenRepo: refreshTokenRepo,
		auditService:     auditService,
		token:            token,
		logger:           logger,
	}
}

func (s *service) Register(ctx context.Context, req dto.RegisterRequest) (*dto.RegisterResponse, error) {
	existing, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		return nil, ErrEmailExists
	}

	hash, err := s.passwordHasher.Hash(req.Password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  hash,
		Status:    model.StatusActive,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	userID := user.ID.String()
	if auditErr := s.auditService.Log(ctx, userID, userID, audit.ActionUserCreated, "user", userID, map[string]any{"user": user.Email}); auditErr != nil {
		s.logger.Error(
			"failed to write audit log",
			"action", audit.ActionUserCreated,
			"user_id", user.ID,
			"error", err,
		)
	}

	return &dto.RegisterResponse{
		ID: userID,
	}, nil
}

func (s *service) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if err := s.passwordHasher.Verify(user.Password, req.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	accessToken, err := s.token.GenerateAccessToken(user.ID.String())
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.token.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	tokenHash := s.token.HashRefreshToken(refreshToken)

	err = s.refreshTokenRepo.Create(ctx, &authmodel.RefreshToken{
		UserID:    user.ID.String(),
		Token:     tokenHash,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	})

	if err != nil {
		return nil, err
	}

	userID := user.ID.String()
	if auditErr := s.auditService.Log(ctx, userID, userID, audit.ActionUserLogin, "user", userID, map[string]any{"user": user.Email}); auditErr != nil {
		s.logger.Error(
			"failed to write audit log",
			"action", audit.ActionUserLogin,
			"user_id", user.ID,
			"error", err,
		)
	}

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *service) Refresh(ctx context.Context, req dto.RefreshRequest) (*dto.RefreshResponse, error) {
	tokenHash := s.token.HashRefreshToken(req.RefreshToken)

	stored, err := s.refreshTokenRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}

	if stored == nil {
		return nil, ErrInvalidCredentials
	}

	if time.Now().After(stored.ExpiresAt) {
		return nil, ErrInvalidCredentials
	}

	newAccessToken, err := s.token.GenerateAccessToken(stored.UserID)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := s.token.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	newHash := s.token.HashRefreshToken(newRefreshToken)

	stored.Token = newHash
	stored.ExpiresAt = time.Now().Add(30 * 24 * time.Hour)

	if err := s.refreshTokenRepo.Update(ctx, stored); err != nil {
		return nil, err
	}

	return &dto.RefreshResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *service) Logout(ctx context.Context, refreshToken string) error {

	hash := s.token.HashRefreshToken(refreshToken)

	stored, err := s.refreshTokenRepo.FindByTokenHash(ctx, hash)
	if err != nil {
		return err
	}

	if stored == nil {
		return nil
	}

	return s.refreshTokenRepo.Delete(ctx, stored.ID)
}
