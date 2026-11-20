package context

import (
	"errors"

	"github.com/fredele20/nimbus/internal/modules/user/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	UserIDKey = "user_id"
	UserKey   = "authenticated_user"
)

var ErrUserNotFound = errors.New("user not found")

func UserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	userID, err := uuid.Parse(c.GetString(UserIDKey))
	if err != nil || userID == uuid.Nil {
		return uuid.Nil, ErrUserNotFound
	}

	return userID, nil
}

func UserFromContext(c *gin.Context) (*model.User, error) {
	value, exists := c.Get(UserKey)
	if !exists {
		return nil, ErrUserNotFound
	}

	user, ok := value.(*model.User)
	if !ok || user == nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}
