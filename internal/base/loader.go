package base

import (
	"log/slog"

	"github.com/fredele20/nimbus/internal/platform/config"
	"gorm.io/gorm"
)

func New(db *gorm.DB, logger *slog.Logger, conf config.Config) *base {
	return &base{
		db:     db,
		logger: logger,
		conf:   conf,
	}
}

func (b *base) LoadController() baseControllers {
	bc := baseControllers{}

	bc.AuthC = b.WithAuthController()
	bc.UserC = b.WithUserController()

	return bc
}
