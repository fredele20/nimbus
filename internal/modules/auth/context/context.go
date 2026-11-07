package context

import (
	"github.com/fredele20/nimbus/internal/modules/user/model"
	"github.com/fredele20/nimbus/internal/modules/user/repository"
	"github.com/fredele20/nimbus/internal/platform/database"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func UserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	uid := c.GetString("user_id")

	userID := uuid.MustParse(uid)
	// if userID == uuid.Nil {
	// 	return uuid.Nil, errors.New("authenticated user not found")
	// }

	return userID, nil
}

func UserFromContext(c *gin.Context) (*model.User, error) {
	userID, err := UserIDFromContext(c)
	if err != nil {
		return nil, err
	}
	user, userErr := repository.New(database.DB).FindOneByID(c, userID)
	if userErr != nil {
		return nil, userErr
	}

	return &user, nil
}
