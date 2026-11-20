package middleware

import (
	"net/http"

	"github.com/fredele20/nimbus/internal/modules/auth/context"
	"github.com/fredele20/nimbus/internal/modules/user/model"
	"github.com/fredele20/nimbus/internal/shared/http/response"
	"github.com/gin-gonic/gin"
)

func RequireRole(roles ...string) gin.HandlerFunc {
	allowedRoles := make(map[string]struct{}, len(roles))

	for _, role := range roles {
		allowedRoles[role] = struct{}{}
	}

	return func(ctx *gin.Context) {
		user, userErr := context.UserFromContext(ctx)
		if userErr != nil || user == nil {
			response.JsonStatusErrorResponse(ctx, http.StatusUnauthorized, "unauthorized")
			ctx.Abort()
			return
		}

		if user.Status != model.StatusActive {
			response.JsonStatusErrorResponse(ctx, http.StatusForbidden, "account is not active	")
			ctx.Abort()
			return
		}

		if _, allowed := allowedRoles[string(user.Role)]; !allowed {
			response.JsonStatusErrorResponse(ctx, http.StatusForbidden, "forbidden")
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
