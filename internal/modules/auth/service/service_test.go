package service_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/fredele20/nimbus/internal/modules/audit"
	"github.com/fredele20/nimbus/internal/modules/auth/dto"
	authModel "github.com/fredele20/nimbus/internal/modules/auth/model"
	"github.com/fredele20/nimbus/internal/modules/auth/service"
	userModel "github.com/fredele20/nimbus/internal/modules/user/model"
	audit_mocks "github.com/fredele20/nimbus/mocks/audit"
	auth_mocks "github.com/fredele20/nimbus/mocks/auth"
	user_mocks "github.com/fredele20/nimbus/mocks/user"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ServiceTestSuite struct {
	suite.Suite

	service service.Service

	userRepo       *user_mocks.MockRepository
	passwordHasher *auth_mocks.MockPasswordHasher
	refreshRepo    *auth_mocks.MockRefreshTokenRepository
	tokenManager   *auth_mocks.MockTokenManager
	auditService   *audit_mocks.MockService
	logger         *slog.Logger
}

func (s *ServiceTestSuite) SetupTest() {
	s.userRepo = user_mocks.NewMockRepository(s.T())
	s.passwordHasher = auth_mocks.NewMockPasswordHasher(s.T())
	s.refreshRepo = auth_mocks.NewMockRefreshTokenRepository(s.T())
	s.tokenManager = auth_mocks.NewMockTokenManager(s.T())
	s.auditService = audit_mocks.NewMockService(s.T())
	s.logger = &slog.Logger{}

	s.service = service.New(s.userRepo, s.passwordHasher, s.refreshRepo, s.auditService, s.tokenManager, s.logger)
}

func TestServiceSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}

func (s *ServiceTestSuite) Test_Register() {
	ctx := context.Background()
	request := dto.RegisterRequest{
		FirstName: "firstname",
		LastName:  "lastname",
		Email:     "email",
		Password:  "password",
	}

	s.Run("Register_Success", func() {
		s.passwordHasher.EXPECT().Hash(request.Password).Return("hashed_password", nil).Once()

		s.userRepo.EXPECT().Create(ctx, mock.Anything).Return(nil).Once()

		s.auditService.EXPECT().Log(
			mock.Anything,
			mock.Anything,
			mock.Anything,
			audit.ActionUserCreated,
			"user",
			mock.Anything,
			mock.Anything,
		).Return(nil).Once()

		res, err := s.service.Register(ctx, request)

		s.Require().NoError(err)
		s.Require().NotNil(res)
		s.Require().NotEmpty(res.ID)

		s.userRepo.AssertExpectations(s.T())
		s.auditService.AssertExpectations(s.T())
	})

	s.Run("Register_Failure - Create User Error", func() {
		s.passwordHasher.EXPECT().Hash(request.Password).Return("hashed_password", nil).Once()

		s.userRepo.EXPECT().Create(ctx, mock.Anything).Return(errors.New("failed to create")).Once()

		res, err := s.service.Register(ctx, request)

		s.Require().Error(err)
		s.Require().Nil(res)

		s.userRepo.AssertExpectations(s.T())
	})
}

func (s *ServiceTestSuite) Test_Login() {
	ctx := context.Background()
	request := dto.LoginRequest{
		Email:    "email",
		Password: "password",
	}

	testUser := userModel.User{
		FirstName: "test-firstname",
		LastName:  "test-lastname",
		Email:     request.Email,
	}

	s.Run("Login_Success", func() {
		s.userRepo.EXPECT().FindByEmail(ctx, request.Email).Return(testUser, nil).Once()

		s.passwordHasher.EXPECT().Verify(testUser.Password, request.Password).Return(nil).Once()

		s.tokenManager.EXPECT().GenerateAccessToken(testUser.ID.String()).Return(mock.Anything, nil).Once()

		s.tokenManager.EXPECT().GenerateRefreshToken().Return(mock.Anything, nil).Once()

		s.tokenManager.EXPECT().HashRefreshToken(mock.AnythingOfType("string")).Return(mock.Anything).Once()

		s.refreshRepo.EXPECT().Create(ctx, mock.AnythingOfType("*model.RefreshToken")).Return(nil).Once()

		s.auditService.EXPECT().Log(mock.Anything, mock.Anything, mock.Anything, audit.ActionUserLogin, "user", mock.Anything, mock.Anything).Return(nil).Once()

		res, err := s.service.Login(ctx, request)

		s.Require().NoError(err)
		s.Require().NotNil(res)

		s.userRepo.AssertExpectations(s.T())
		s.passwordHasher.AssertExpectations(s.T())
		s.tokenManager.AssertExpectations(s.T())
		s.refreshRepo.AssertExpectations(s.T())
	})

	s.Run("Login_Failure - Database error", func() {
		s.userRepo.EXPECT().FindByEmail(ctx, testUser.Email).Return(userModel.User{}, errors.New("database error")).Once()

		res, err := s.service.Login(ctx, request)

		s.Require().Error(err)
		s.Require().Nil(res)

		s.userRepo.AssertExpectations(s.T())
	})

	s.Run("Login_Failure - user not found", func() {
		s.userRepo.EXPECT().FindByEmail(ctx, request.Email).Return(userModel.User{}, service.ErrInvalidCredentials).Once()

		res, err := s.service.Login(ctx, request)

		s.Require().ErrorIs(err, service.ErrInvalidCredentials)
		s.Require().Nil(res)

		s.userRepo.AssertExpectations(s.T())
	})

	s.Run("Login_Failure - Password verification error", func() {
		s.userRepo.EXPECT().FindByEmail(ctx, testUser.Email).Return(testUser, nil).Once()

		s.passwordHasher.EXPECT().Verify(testUser.Password, request.Password).Return(errors.New("passwords do not match")).Once()

		res, err := s.service.Login(ctx, request)

		s.Require().ErrorIs(err, service.ErrInvalidCredentials)
		s.Require().Nil(res)

		s.userRepo.AssertExpectations(s.T())
		s.passwordHasher.AssertExpectations(s.T())
	})

	s.Run("Login_Failure - Access token generation failed", func() {
		s.userRepo.EXPECT().FindByEmail(ctx, request.Email).Return(testUser, nil).Once()

		s.passwordHasher.EXPECT().Verify(testUser.Password, request.Password).Return(nil).Once()

		s.tokenManager.EXPECT().GenerateAccessToken(testUser.ID.String()).Return(mock.Anything, errors.New("failed to generate access token")).Once()

		res, err := s.service.Login(ctx, request)

		s.Require().Error(err)
		s.Require().Nil(res)

		s.userRepo.AssertExpectations(s.T())
		s.passwordHasher.AssertExpectations(s.T())
		s.tokenManager.AssertExpectations(s.T())
	})

	s.Run("Login_Failure - Refresh token generation failed", func() {
		s.userRepo.EXPECT().FindByEmail(ctx, request.Email).Return(testUser, nil).Once()

		s.passwordHasher.EXPECT().Verify(testUser.Password, request.Password).Return(nil).Once()

		s.tokenManager.EXPECT().GenerateAccessToken(testUser.ID.String()).Return(mock.Anything, nil).Once()

		s.tokenManager.EXPECT().GenerateRefreshToken().Return(mock.Anything, errors.New("failed to generate refresh token")).Once()

		res, err := s.service.Login(ctx, request)

		s.Require().Error(err)
		s.Require().Nil(res)

		s.userRepo.AssertExpectations(s.T())
		s.passwordHasher.AssertExpectations(s.T())
		s.tokenManager.AssertExpectations(s.T())
	})

	s.Run("Login_Failure - Failed to create refresh token", func() {
		s.userRepo.EXPECT().FindByEmail(ctx, request.Email).Return(testUser, nil).Once()

		s.passwordHasher.EXPECT().Verify(testUser.Password, request.Password).Return(nil).Once()

		s.tokenManager.EXPECT().GenerateAccessToken(testUser.ID.String()).Return(mock.Anything, nil).Once()

		s.tokenManager.EXPECT().GenerateRefreshToken().Return(mock.Anything, nil).Once()

		s.tokenManager.EXPECT().HashRefreshToken(mock.AnythingOfType("string")).Return(mock.Anything).Once()

		s.refreshRepo.EXPECT().Create(ctx, mock.AnythingOfType("*model.RefreshToken")).Return(errors.New("failed to create refresh token")).Once()

		res, err := s.service.Login(ctx, request)

		s.Require().Error(err)
		s.Require().Nil(res)

		s.userRepo.AssertExpectations(s.T())
		s.passwordHasher.AssertExpectations(s.T())
		s.tokenManager.AssertExpectations(s.T())
		s.refreshRepo.AssertExpectations(s.T())
	})
}

func (s *ServiceTestSuite) Test_RefreshToken() {
	ctx := context.Background()
	request := dto.RefreshRequest{
		RefreshToken: "refresh-token",
	}

	refreshTokenResponse := authModel.RefreshToken{
		UserID:    uuid.New(),
		Token:     "test-token",
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}

	s.Run("Refresh_Success", func() {
		s.tokenManager.EXPECT().HashRefreshToken(request.RefreshToken).Return(mock.Anything).Once()

		s.refreshRepo.EXPECT().FindByToken(ctx, mock.Anything).Return(refreshTokenResponse, nil).Once()

		s.tokenManager.EXPECT().GenerateAccessToken(refreshTokenResponse.UserID.String()).Return(mock.Anything, nil).Once()

		s.tokenManager.EXPECT().GenerateRefreshToken().Return(mock.Anything, nil).Once()

		s.tokenManager.EXPECT().HashRefreshToken(mock.Anything).Return(mock.Anything).Once()

		s.refreshRepo.EXPECT().UpdateOne(ctx, mock.Anything).Return(nil).Once()

		response, err := s.service.Refresh(ctx, request)

		s.Require().NoError(err)
		s.Require().NotNil(response)

		s.tokenManager.AssertExpectations(s.T())
		s.refreshRepo.AssertExpectations(s.T())
	})

	s.Run("Refresh_Failure - Failed to find by token hash", func() {
		s.tokenManager.EXPECT().HashRefreshToken(request.RefreshToken).Return(mock.Anything).Once()

		s.refreshRepo.EXPECT().FindByToken(ctx, mock.Anything).Return(authModel.RefreshToken{}, errors.New("failed to find token")).Once()

		response, err := s.service.Refresh(ctx, request)

		s.Require().Error(err)
		s.Require().Nil(response)

		s.tokenManager.AssertExpectations(s.T())
		s.refreshRepo.AssertExpectations(s.T())
	})

	s.Run("Refresh_Failure - Failed to generate new access token", func() {
		s.tokenManager.EXPECT().HashRefreshToken(request.RefreshToken).Return(mock.Anything).Once()

		s.refreshRepo.EXPECT().FindByToken(ctx, mock.Anything).Return(refreshTokenResponse, nil).Once()

		s.tokenManager.EXPECT().GenerateAccessToken(refreshTokenResponse.UserID.String()).Return("", errors.New("failed to generate new access token")).Once()

		response, err := s.service.Refresh(ctx, request)

		s.Require().Error(err)
		s.Require().Nil(response)

		s.tokenManager.AssertExpectations(s.T())
		s.refreshRepo.AssertExpectations(s.T())
	})

	s.Run("Refresh_Failure - Failed to generate new refresh token", func() {
		s.tokenManager.EXPECT().HashRefreshToken(request.RefreshToken).Return(mock.Anything).Once()

		s.refreshRepo.EXPECT().FindByToken(ctx, mock.Anything).Return(refreshTokenResponse, nil).Once()

		s.tokenManager.EXPECT().GenerateAccessToken(refreshTokenResponse.UserID.String()).Return(mock.Anything, nil).Once()

		s.tokenManager.EXPECT().GenerateRefreshToken().Return(mock.Anything, errors.New("failed to generate new refresh token")).Once()

		response, err := s.service.Refresh(ctx, request)

		s.Require().Error(err)
		s.Require().Nil(response)

		s.tokenManager.AssertExpectations(s.T())
		s.refreshRepo.AssertExpectations(s.T())
	})

	s.Run("Refresh_Failure - Failed to update refresh token", func() {
		s.tokenManager.EXPECT().HashRefreshToken(request.RefreshToken).Return(mock.Anything).Once()

		s.refreshRepo.EXPECT().FindByToken(ctx, mock.Anything).Return(refreshTokenResponse, nil).Once()

		s.tokenManager.EXPECT().GenerateAccessToken(refreshTokenResponse.UserID.String()).Return(mock.Anything, nil).Once()

		s.tokenManager.EXPECT().GenerateRefreshToken().Return(mock.Anything, nil).Once()

		s.tokenManager.EXPECT().HashRefreshToken(mock.Anything).Return(mock.Anything).Once()

		s.refreshRepo.EXPECT().UpdateOne(ctx, mock.Anything).Return(errors.New("failed to update refresh token")).Once()

		response, err := s.service.Refresh(ctx, request)

		s.Require().Error(err)
		s.Require().Nil(response)

		s.tokenManager.AssertExpectations(s.T())
		s.refreshRepo.AssertExpectations(s.T())
	})
}

func (s *ServiceTestSuite) Test_Logout() {
	ctx := context.Background()

	refreshTokenResponse := authModel.RefreshToken{
		UserID:    uuid.New(),
		Token:     "hashed-logout-refresh-token",
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}

	s.Run("Logout_Success", func() {
		refreshToken := "logout-success-token"
		tokenHash := "hashed-logout-refresh-token"

		s.tokenManager.EXPECT().
			HashRefreshToken(refreshToken).
			Return(tokenHash).
			Once()

		s.refreshRepo.EXPECT().
			FindByToken(ctx, tokenHash).
			Return(refreshTokenResponse, nil).
			Once()

		s.refreshRepo.EXPECT().
			DeleteOneById(ctx, refreshTokenResponse.ID).
			Return(nil).
			Once()

		err := s.service.Logout(ctx, refreshToken)

		s.Require().NoError(err)
	})

	s.Run("Logout_Failed - Failed to find existing hashed token", func() {
		refreshToken := "logout-find-failure-token"
		tokenHash := "hashed-logout-find-failure-token"

		s.tokenManager.EXPECT().
			HashRefreshToken(refreshToken).
			Return(tokenHash).
			Once()

		s.refreshRepo.EXPECT().
			FindByToken(ctx, tokenHash).
			Return(authModel.RefreshToken{}, errors.New("failed to find existing token")).
			Once()

		err := s.service.Logout(ctx, refreshToken)

		s.Require().Error(err)
	})

	s.Run("Logout_Failed - Failed to delete existing hashed token", func() {
		refreshToken := "logout-delete-failure-token"
		tokenHash := "hashed-logout-delete-failure-token"

		storedToken := authModel.RefreshToken{
			UserID:    uuid.New(),
			Token:     tokenHash,
			ExpiresAt: time.Now().Add(15 * time.Minute),
		}

		s.tokenManager.EXPECT().
			HashRefreshToken(refreshToken).
			Return(tokenHash).
			Once()

		s.refreshRepo.EXPECT().
			FindByToken(ctx, tokenHash).
			Return(storedToken, nil).
			Once()

		s.refreshRepo.EXPECT().
			DeleteOneById(ctx, storedToken.ID).
			Return(errors.New("failed to delete existing token")).
			Once()

		err := s.service.Logout(ctx, refreshToken)

		s.Require().Error(err)
	})
}
