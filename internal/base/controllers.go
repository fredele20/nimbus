package base

import (
	authController "github.com/fredele20/nimbus/internal/modules/auth/controller"
	userController "github.com/fredele20/nimbus/internal/modules/user/controller"
	walletController "github.com/fredele20/nimbus/internal/modules/wallet/controller"
)

func (b *base) WithAuthController() authController.ControllerInterface {
	return authController.New(b.WithAuthService())
}

func (b *base) WithUserController() userController.ControllerInterface {
	return userController.New(b.WithUserService())
}

func (b *base) WithWalletController() walletController.ControllerInterface {
	return walletController.New(
		b.WithWalletService(),
		b.WithAdminWalletService(),
	)
}
