package base

import (
	"log/slog"

	authController "github.com/fredele20/nimbus/internal/modules/auth/controller"
	userController "github.com/fredele20/nimbus/internal/modules/user/controller"
	"github.com/fredele20/nimbus/internal/platform/config"
	"gorm.io/gorm"
)

type base struct {
	db     *gorm.DB
	logger *slog.Logger
	conf   config.Config
}

type baseControllers struct {
	AuthC authController.ControllerInterface
	UserC userController.ControllerInterface
}
