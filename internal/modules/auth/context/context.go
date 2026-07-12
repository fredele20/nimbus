package context

import (
	"errors"

	"github.com/gin-gonic/gin"
)

func UserIDFromContext(c *gin.Context) (string, error) {
	userID := c.GetString("user_id")
	if userID == "" {
		return "", errors.New("authenticated user not found")
	}
	
	return userID, nil
}
