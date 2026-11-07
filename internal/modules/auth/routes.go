package auth

import (
	"github.com/fredele20/nimbus/internal/modules/auth/controller"
	"github.com/fredele20/nimbus/internal/modules/auth/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	r *gin.RouterGroup,
	controller controller.ControllerInterface,
	jwtSecret string,
) {
	auth := r.Group("/auth")
	auth.POST("/register", controller.Register)
	auth.POST("/login", controller.Login)

	protectedRoute := r.Group("")
	protectedRoute.Use(middleware.AuthMiddleware(jwtSecret))

	protectedRoute.POST("/refresh", controller.Refresh)
	protectedRoute.POST("/logout", controller.Logout)
}
