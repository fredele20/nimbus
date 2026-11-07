package middleware

import (
	"net/http"
	"strings"

	"github.com/fredele20/nimbus/internal/modules/auth/token"
	"github.com/fredele20/nimbus/internal/shared/http/response"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		const prefix = "Bearer"

		if authHeader == "" {
			response.JsonStatusErrorResponse(c, http.StatusUnauthorized, "authorization header is required")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)

		if len(parts) != 2 || parts[0] != prefix || parts[1] == "" {
			response.JsonStatusErrorResponse(c, http.StatusUnauthorized, "invalid authorization header")
			return
		}

		claims, err := token.New(jwtSecret).Parse(parts[1])
		if err != nil {
			response.JsonStatusErrorResponse(c, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		c.Set("user_id", claims.UserID)
		c.Next()
	}
}
