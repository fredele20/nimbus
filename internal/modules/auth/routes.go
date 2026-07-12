package auth

import (
	"github.com/fredele20/nimbus/internal/modules/auth/handler"
	"github.com/fredele20/nimbus/internal/modules/auth/middleware"
	"github.com/gin-gonic/gin"
)


func RegisterRoutes(r *gin.RouterGroup, handler *handler.Handler) {
	// apiV1 := c.Group("/api/v1")

	auth := r.Group("/auth")
	auth.POST("/register", handler.Register)
	auth.POST("/login", handler.Register)

	protectedRoute := r.Group("")
	protectedRoute.Use(middleware.AuthMiddleware)

	protectedRoute.POST("/refresh", handler.Refresh)
	protectedRoute.POST("/logout", handler.Logout)
	protectedRoute.GET("/me", handler.Me)

}
