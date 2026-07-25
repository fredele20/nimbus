package base

import (
	auditRepository "github.com/fredele20/nimbus/internal/modules/audit/repository"
	authRepository "github.com/fredele20/nimbus/internal/modules/auth/repository"
	userRepository "github.com/fredele20/nimbus/internal/modules/user/repository"
)

func (b *base) WithUserRepository() userRepository.Repository {
	return userRepository.New(b.db)
}

func (b *base) WithRefreshTokenRepository() authRepository.RefreshTokenRepository {
	return authRepository.NewRefreshTokenRepository(b.db)
}

func (b *base) WithAuditRepository() auditRepository.Repository {
	return auditRepository.New(b.db)
}

