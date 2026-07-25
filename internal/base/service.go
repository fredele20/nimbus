package base

import (
	auditService "github.com/fredele20/nimbus/internal/modules/audit/service"
	"github.com/fredele20/nimbus/internal/modules/auth/password"
	authService "github.com/fredele20/nimbus/internal/modules/auth/service"
	"github.com/fredele20/nimbus/internal/modules/auth/token"
	userService "github.com/fredele20/nimbus/internal/modules/user/service"
)

func (b *base) WithUserService() userService.Service {
	return userService.New(b.WithUserRepository(), b.WithAuditService(), b.logger)
}

func (b *base) WithAuditService() auditService.Service {
	return auditService.New(b.WithAuditRepository())
}

func (b *base) WithAuthService() authService.Service {
	return authService.New(
		b.WithUserRepository(),
		b.WithPasswordHasher(),
		b.WithRefreshTokenRepository(),
		b.WithAuditService(),
		b.WithTokenManager(),
		b.logger,
	)
}

func (b *base) WithPasswordHasher() password.PasswordHasher {
	return password.New()
}

func (b *base) WithTokenManager() token.TokenManager {
	return token.New(b.conf.JWT.Secret)
}
