package middleware

import (
	"net/http"

	authCtx "github.com/fredele20/nimbus/internal/modules/auth/context"
	"github.com/fredele20/nimbus/internal/modules/user/repository"
	"github.com/fredele20/nimbus/internal/shared/http/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func LoadUser(db *gorm.DB) gin.HandlerFunc {
	userRepository := repository.New(db)

	return func(ctx *gin.Context) {
		userID, err := authCtx.UserIDFromContext(ctx)
		if err != nil {
			response.JsonStatusErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
			ctx.Abort()
			return
		}

		user, err := userRepository.FindOneByID(ctx.Request.Context(), userID)
		if err != nil {
			response.JsonStatusErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
			ctx.Abort()
			return
		}

		ctx.Set(authCtx.UserKey, &user)
		ctx.Next()
	}
}
