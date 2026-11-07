package user

import (
	authmiddleware "github.com/fredele20/nimbus/internal/modules/auth/middleware"
	"github.com/fredele20/nimbus/internal/modules/user/controller"
	"github.com/fredele20/nimbus/internal/modules/user/model"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, controller controller.ControllerInterface, jwtSecret string) {
	authUser := router.Group("")
	authUser.Use(authmiddleware.AuthMiddleware(jwtSecret))

	authUser.GET("/me", controller.GetProfile)
	authUser.PATCH("/me", controller.UpdateProfile)

	admin := router.Group("/admin")

	admin.Use(authmiddleware.AuthMiddleware(jwtSecret), authmiddleware.RequireRole(string(model.Admin)))

	admin.GET("/users", controller.ListUsers)

	admin.PATCH("/users/:id/status", controller.UpdateStatus)
}
