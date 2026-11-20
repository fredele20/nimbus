package wallet

import (
	authMiddleware "github.com/fredele20/nimbus/internal/modules/auth/middleware"
	userModel "github.com/fredele20/nimbus/internal/modules/user/model"
	"github.com/fredele20/nimbus/internal/modules/wallet/controller"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.RouterGroup, walletController controller.ControllerInterface, jwtSecret string, db *gorm.DB) {
	wallets := router.Group("/wallets")
	wallets.Use(
		authMiddleware.AuthMiddleware(jwtSecret),
		authMiddleware.LoadUser(db),
		authMiddleware.RequireActiveUser(),
	)

	wallets.POST("", walletController.Create)
	wallets.GET("", walletController.List)
	wallets.GET("/:id", walletController.GetByID)
	wallets.POST("/:id/freeze", walletController.Freeze)
	wallets.POST("/:id/unfreeze", walletController.Unfreeze)

	admin := router.Group("/admin/wallets")
	admin.Use(
		authMiddleware.AuthMiddleware(jwtSecret),
		authMiddleware.LoadUser(db),
		authMiddleware.RequireRole(string(userModel.Admin)),
	)

	admin.POST("/:id/freeze", walletController.AdminFreeze)
	admin.POST("/:id/unfreeze", walletController.AdminUnfreeze)
}
