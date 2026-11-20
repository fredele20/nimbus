package middleware

import (
	"net/http"

	authCtx "github.com/fredele20/nimbus/internal/modules/auth/context"
	"github.com/fredele20/nimbus/internal/modules/user/model"
	"github.com/fredele20/nimbus/internal/shared/http/response"
	"github.com/gin-gonic/gin"
)

func RequireActiveUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user, err := authCtx.UserFromContext(ctx)
		if err != nil || user == nil {
			response.JsonStatusErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
			ctx.Abort()
			return
		}

		if user.Status != model.StatusActive {
			response.JsonStatusErrorResponse(ctx, http.StatusForbidden, "account is not active")
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
