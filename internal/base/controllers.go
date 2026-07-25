package base

import (
	authController "github.com/fredele20/nimbus/internal/modules/auth/controller"
	userController "github.com/fredele20/nimbus/internal/modules/user/controller"
)

func (b *base) WithAuthController() authController.ControllerInterface {
	return authController.New(b.WithAuthService())
}

func (b *base) WithUserController() userController.ControllerInterface {
	return userController.New(b.WithUserService())
}
