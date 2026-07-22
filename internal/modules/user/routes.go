package user

import (
	authmiddleware "github.com/fredele20/nimbus/internal/modules/auth/middleware"
	"github.com/fredele20/nimbus/internal/modules/user/handler"
	"github.com/fredele20/nimbus/internal/modules/user/model"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, handler *handler.Handler) {
	authUser := router.Group("")
	authUser.Use(authmiddleware.AuthMiddleware)

	authUser.GET("/me", handler.GetProfile)
	authUser.PATCH("/me", handler.UpdateProfile)

	admin := router.Group("/admin")

	admin.Use(authmiddleware.AuthMiddleware, authmiddleware.RequireRole(string(model.Admin)))

	admin.GET("/users", handler.ListUsers)

	admin.PATCH("/users/:id/status", handler.UpdateStatus)
}
