package service_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/fredele20/nimbus/internal/modules/auth/dto"
	"github.com/fredele20/nimbus/internal/modules/auth/service"
	userModel "github.com/fredele20/nimbus/internal/modules/user/model"

	authModel "github.com/fredele20/nimbus/internal/modules/auth/model"
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

	testUser := userModel.User{
		ID:        uuid.New(),
		FirstName: "test-firstname",
		LastName:  "test-lastname",
		Email:     request.Email,
	}

	s.Run("Register_Success", func() {
		s.userRepo.EXPECT().FindByEmail(ctx, request.Email).Return(nil, nil).Once()

		s.passwordHasher.EXPECT().Hash(request.Password).Return(string(mock.AnythingOfType("string")), nil).Once()

		s.userRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*model.User")).Return(nil).Once()

		res, err := s.service.Register(ctx, request)

		s.Require().NoError(err)
		s.Require().NotNil(res)

		s.userRepo.AssertExpectations(s.T())
	})

	s.Run("Register_Failure - EmailAlreadyExists", func() {
		s.userRepo.EXPECT().FindByEmail(ctx, testUser.Email).Return(&testUser, nil).Once()

		res, err := s.service.Register(ctx, request)

		s.Require().Error(err)
		s.Require().Nil(res)

		s.userRepo.AssertNotCalled(s.T(), "Create", mock.AnythingOfType("*model.User"))
		s.userRepo.AssertExpectations(s.T())
	})

	s.Run("Register_Failure - Create User Error", func() {
		s.userRepo.EXPECT().FindByEmail(ctx, request.Email).Return(nil, nil).Once()

		s.passwordHasher.EXPECT().Hash(request.Password).Return(string(mock.AnythingOfType("string")), nil).Once()

		s.userRepo.EXPECT().Create(ctx, mock.AnythingOfType("*model.User")).Return(errors.New("failed to create")).Once()

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
		ID:        uuid.New(),
		FirstName: "test-firstname",
		LastName:  "test-lastname",
		Email:     request.Email,
	}

	s.Run("Login_Success", func() {
		s.userRepo.EXPECT().FindByEmail(ctx, request.Email).Return(&testUser, nil).Once()

		s.passwordHasher.EXPECT().Verify(testUser.Password, request.Password).Return(nil).Once()

		s.tokenManager.EXPECT().GenerateAccessToken(testUser.ID.String()).Return(mock.Anything, nil).Once()

		s.tokenManager.EXPECT().GenerateRefreshToken().Return(mock.Anything, nil).Once()

		s.tokenManager.EXPECT().HashRefreshToken(mock.AnythingOfType("string")).Return(mock.Anything).Once()

		s.refreshRepo.EXPECT().Create(ctx, mock.AnythingOfType("*model.RefreshToken")).Return(nil).Once()

		res, err := s.service.Login(ctx, request)

		s.Require().NoError(err)
		s.Require().NotNil(res)

		s.userRepo.AssertExpectations(s.T())
		s.passwordHasher.AssertExpectations(s.T())
		s.tokenManager.AssertExpectations(s.T())
		s.refreshRepo.AssertExpectations(s.T())
	})

	s.Run("Login_Failure - Database error", func() {
		s.userRepo.EXPECT().FindByEmail(ctx, testUser.Email).Return(nil, errors.New("database error")).Once()

		res, err := s.service.Login(ctx, request)

		s.Require().Error(err)
		s.Require().Nil(res)

		s.userRepo.AssertExpectations(s.T())
	})

	s.Run("Login_Failure - user not found", func() {
		s.userRepo.EXPECT().FindByEmail(ctx, request.Email).Return(nil, nil).Once()

		res, err := s.service.Login(ctx, request)

		s.Require().ErrorIs(err, service.ErrInvalidCredentials)
		s.Require().Nil(res)

		s.userRepo.AssertExpectations(s.T())
	})

	s.Run("Login_Failure - Password verification error", func() {
		s.userRepo.EXPECT().FindByEmail(ctx, testUser.Email).Return(&testUser, nil).Once()

		s.passwordHasher.EXPECT().Verify(testUser.Password, request.Password).Return(errors.New("passwords do not match")).Once()

		res, err := s.service.Login(ctx, request)

		s.Require().ErrorIs(err, service.ErrInvalidCredentials)
		s.Require().Nil(res)

		s.userRepo.AssertExpectations(s.T())
		s.passwordHasher.AssertExpectations(s.T())
	})

	s.Run("Login_Failure - Access token generation failed", func() {
		s.userRepo.EXPECT().FindByEmail(ctx, request.Email).Return(&testUser, nil).Once()

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
		s.userRepo.EXPECT().FindByEmail(ctx, request.Email).Return(&testUser, nil).Once()

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
		s.userRepo.EXPECT().FindByEmail(ctx, request.Email).Return(&testUser, nil).Once()

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
		ID:        uuid.NewString(),
		UserID:    uuid.NewString(),
		Token:     "test-token",
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}

	s.Run("Refresh_Success", func() {
		s.tokenManager.EXPECT().HashRefreshToken(request.RefreshToken).Return(mock.Anything).Once()

		s.refreshRepo.EXPECT().FindByTokenHash(ctx, mock.Anything).Return(&refreshTokenResponse, nil).Once()

		s.tokenManager.EXPECT().GenerateAccessToken(refreshTokenResponse.UserID).Return(mock.Anything, nil).Once()

		s.tokenManager.EXPECT().GenerateRefreshToken().Return(mock.Anything, nil).Once()

		s.tokenManager.EXPECT().HashRefreshToken(mock.Anything).Return(mock.Anything).Once()

		s.refreshRepo.EXPECT().Update(ctx, mock.Anything).Return(nil).Once()

		response, err := s.service.Refresh(ctx, request)

		s.Require().NoError(err)
		s.Require().NotNil(response)

		s.tokenManager.AssertExpectations(s.T())
		s.refreshRepo.AssertExpectations(s.T())
	})

	s.Run("Refresh_Failure - Failed to find by token hash", func() {
		s.tokenManager.EXPECT().HashRefreshToken(request.RefreshToken).Return(mock.Anything).Once()

		s.refreshRepo.EXPECT().FindByTokenHash(ctx, mock.Anything).Return(nil, errors.New("failed to find token")).Once()

		response, err := s.service.Refresh(ctx, request)

		s.Require().Error(err)
		s.Require().Nil(response)

		s.tokenManager.AssertExpectations(s.T())
		s.refreshRepo.AssertExpectations(s.T())
	})

	s.Run("Refresh_Failure - Failed to generate new access token", func() {
		s.tokenManager.EXPECT().HashRefreshToken(request.RefreshToken).Return(mock.Anything).Once()

		s.refreshRepo.EXPECT().FindByTokenHash(ctx, mock.Anything).Return(&refreshTokenResponse, nil).Once()

		s.tokenManager.EXPECT().GenerateAccessToken(refreshTokenResponse.UserID).Return("", errors.New("failed to generate new access token")).Once()

		response, err := s.service.Refresh(ctx, request)

		s.Require().Error(err)
		s.Require().Nil(response)

		s.tokenManager.AssertExpectations(s.T())
		s.refreshRepo.AssertExpectations(s.T())
	})

	s.Run("Refresh_Failure - Failed to generate new refresh token", func() {
		s.tokenManager.EXPECT().HashRefreshToken(request.RefreshToken).Return(mock.Anything).Once()

		s.refreshRepo.EXPECT().FindByTokenHash(ctx, mock.Anything).Return(&refreshTokenResponse, nil).Once()

		s.tokenManager.EXPECT().GenerateAccessToken(refreshTokenResponse.UserID).Return(mock.Anything, nil).Once()

		s.tokenManager.EXPECT().GenerateRefreshToken().Return(mock.Anything, errors.New("failed to generate new refresh token")).Once()

		response, err := s.service.Refresh(ctx, request)

		s.Require().Error(err)
		s.Require().Nil(response)

		s.tokenManager.AssertExpectations(s.T())
		s.refreshRepo.AssertExpectations(s.T())
	})

	s.Run("Refresh_Failure - Failed to update refresh token", func() {
		s.tokenManager.EXPECT().HashRefreshToken(request.RefreshToken).Return(mock.Anything).Once()

		s.refreshRepo.EXPECT().FindByTokenHash(ctx, mock.Anything).Return(&refreshTokenResponse, nil).Once()

		s.tokenManager.EXPECT().GenerateAccessToken(refreshTokenResponse.UserID).Return(mock.Anything, nil).Once()

		s.tokenManager.EXPECT().GenerateRefreshToken().Return(mock.Anything, nil).Once()

		s.tokenManager.EXPECT().HashRefreshToken(mock.Anything).Return(mock.Anything).Once()

		s.refreshRepo.EXPECT().Update(ctx, mock.Anything).Return(errors.New("failed to update refresh token")).Once()

		response, err := s.service.Refresh(ctx, request)

		s.Require().Error(err)
		s.Require().Nil(response)

		s.tokenManager.AssertExpectations(s.T())
		s.refreshRepo.AssertExpectations(s.T())
	})
}

func (s *ServiceTestSuite) Test_Logout() {
	ctx := context.Background()
	refreshToken := mock.Anything

	refreshTokenResponse := authModel.RefreshToken{
		ID:        uuid.NewString(),
		UserID:    uuid.NewString(),
		Token:     "test-token",
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}

	s.Run("Logout_Success", func() {
		s.tokenManager.EXPECT().HashRefreshToken(refreshToken).Return(mock.Anything).Once()

		s.refreshRepo.EXPECT().FindByTokenHash(ctx, mock.Anything).Return(&refreshTokenResponse, nil).Once()

		s.refreshRepo.EXPECT().Delete(ctx, refreshTokenResponse.ID).Return(nil).Once()

		err := s.service.Logout(ctx, refreshToken)

		s.Require().NoError(err)

		s.tokenManager.AssertExpectations(s.T())
		s.refreshRepo.AssertExpectations(s.T())
	})

	s.Run("Logout_Failed - Failed to find existing hashed token", func() {
		s.tokenManager.EXPECT().HashRefreshToken(refreshToken).Return(mock.Anything).Once()

		s.refreshRepo.EXPECT().FindByTokenHash(ctx, mock.Anything).Return(nil, errors.New("failed to find existing token")).Once()

		err := s.service.Logout(ctx, refreshToken)

		s.Require().Error(err)

		s.tokenManager.AssertExpectations(s.T())
		s.refreshRepo.AssertExpectations(s.T())
	})

	s.Run("Logout_Failed - Failed to delete existing hashed token", func() {
		s.tokenManager.EXPECT().HashRefreshToken(refreshToken).Return(mock.Anything).Once()

		s.refreshRepo.EXPECT().FindByTokenHash(ctx, mock.Anything).Return(&refreshTokenResponse, nil).Once()

		s.refreshRepo.EXPECT().Delete(ctx, refreshTokenResponse.ID).Return(errors.New("failed to delete existing token")).Once()

		err := s.service.Logout(ctx, refreshToken)

		s.Require().Error(err)

		s.tokenManager.AssertExpectations(s.T())
		s.refreshRepo.AssertExpectations(s.T())
	})
}
