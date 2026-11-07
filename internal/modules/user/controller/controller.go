package controller

import (
	"net/http"

	authctx "github.com/fredele20/nimbus/internal/modules/auth/context"
	"github.com/fredele20/nimbus/internal/modules/user/dto"
	"github.com/fredele20/nimbus/internal/modules/user/model"
	"github.com/fredele20/nimbus/internal/modules/user/service"
	"github.com/fredele20/nimbus/internal/shared/http/response"
	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
)

type controller struct {
	service service.Service
}

type ControllerInterface interface {
	GetProfile(*gin.Context)
	UpdateProfile(*gin.Context)
	ListUsers(*gin.Context)
	UpdateStatus(*gin.Context)
}

func New(service service.Service) ControllerInterface {
	return &controller{
		service: service,
	}
}

func (c *controller) GetProfile(ctx *gin.Context) {
	user, _ := authctx.UserFromContext(ctx)
	if user == nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	res, err := c.service.GetProfile(ctx.Request.Context(), user.ID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, res)
}

func (c *controller) UpdateProfile(ctx *gin.Context) {
	user, _ := authctx.UserFromContext(ctx)
	if user == nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	var req dto.UpdateProfileRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := c.service.UpdateProfile(ctx.Request.Context(), user.ID, req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (c *controller) ListUsers(ctx *gin.Context) {
	users, err := c.service.ListUsers(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, users)
}

func (c *controller) UpdateStatus(ctx *gin.Context) {
	uid := ctx.Param("id")

	userId, idErr := uuid.Parse(uid)
	if idErr != nil {
		response.JsonErrorResponse(ctx, "invalid user_id")
		return
	}

	var req dto.UpdateUserStatusRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := c.service.UpdateStatus(ctx.Request.Context(), userId, model.Status(req.Status)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.Status(http.StatusNoContent)
}
