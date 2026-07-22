package context

import (
	"errors"

	"github.com/fredele20/nimbus/internal/modules/user/model"
	"github.com/fredele20/nimbus/internal/modules/user/repository"
	"github.com/fredele20/nimbus/internal/platform/database"
	"github.com/gin-gonic/gin"
)

func UserIDFromContext(c *gin.Context) (string, error) {
	userID := c.GetString("user_id")
	if userID == "" {
		return "", errors.New("authenticated user not found")
	}
	
	return userID, nil
}

func UserFromContext(c *gin.Context) (*model.User, error) {
	userID, err := UserIDFromContext(c)
	if err != nil {
		return nil, err
	}
	user, userErr := repository.New(database.DB).FindByID(c, userID)
	if userErr != nil {
		return nil, userErr
	}

	return user, nil
}
