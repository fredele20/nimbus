package middleware

import (
	"net/http"
	"strings"

	"github.com/fredele20/nimbus/internal/modules/auth/token"
	"github.com/fredele20/nimbus/internal/platform/config"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	const prefix = "Bearer"

	if authHeader == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "authorization header is required",
		})
		return
	}

	parts := strings.SplitN(authHeader, " ", 2)

	if len(parts) != 2 || parts[0] != "Bearer" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "invalid authorization header",
		})
		return
	}

	tokenString := parts[0]
	claims, err := token.New(config.EnvJWTSecret).Parse(tokenString)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "invalid or expired token",
		})
		return
	}

	c.Set("user_id", claims.UserID)

	c.Next()
}
