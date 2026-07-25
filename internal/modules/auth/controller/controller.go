package controller

import (
	"errors"

	"github.com/gin-gonic/gin"

	authdto "github.com/fredele20/nimbus/internal/modules/auth/dto"
	authservice "github.com/fredele20/nimbus/internal/modules/auth/service"
	"github.com/fredele20/nimbus/internal/shared/http/response"
	"github.com/fredele20/nimbus/internal/shared/validation"
)

type controller struct {
	service authservice.Service
}

type ControllerInterface interface {
	Register(*gin.Context)
	Login(*gin.Context)
	Refresh(*gin.Context)
	Logout(*gin.Context)
}

func New(service authservice.Service) ControllerInterface {
	return &controller{service: service}
}

func (c *controller) Register(ctx *gin.Context) {
	var req authdto.RegisterRequest

	if err := validation.Validate(ctx, &req); err != nil {
		response.JsonErrorResponse(ctx, err.Error())
		// response.JsonErrorResponse(c, err.Error())
		return
	}

	res, err := c.service.Register(ctx.Request.Context(), req)
	if err != nil {
		response.JsonErrorResponse(ctx, err.Error())
		return
	}

	response.JsonCreatedResponse(ctx, res, "")
}

func (c *controller) Login(ctx *gin.Context) {
	var req authdto.LoginRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.JsonErrorResponse(ctx, err.Error())
		return
	}

	res, err := c.service.Login(ctx.Request.Context(), req)
	if err != nil {
		response.JsonErrorResponse(ctx, err.Error())
		return
	}

	response.JsonSuccessResponse(ctx, res, "")
}

func (c *controller) Refresh(ctx *gin.Context) {
	var req authdto.RefreshRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.JsonErrorResponse(ctx, err.Error())
		return
	}

	res, err := c.service.Refresh(ctx.Request.Context(), req)
	if err != nil {
		if errors.Is(err, authservice.ErrInvalidCredentials) {
			response.JsonErrorResponse(ctx, err.Error())
			return
		}

		response.JsonErrorResponse(ctx, err.Error())
		return
	}

	response.JsonSuccessResponse(ctx, res, "")
}

func (c *controller) Logout(ctx *gin.Context) {
	var req authdto.RefreshRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.JsonErrorResponse(ctx, err.Error())
		return
	}

	err := c.service.Logout(ctx.Request.Context(), req.RefreshToken)
	if err != nil {
		response.JsonErrorResponse(ctx, err.Error())
		return
	}

	response.JsonSuccessResponse(ctx, nil, "")
}

// func (h *controller) Me(c *gin.Context) {
// 	userID, err := context.UserIDFromContext(c)
// 	if err != nil {
// 		// httperror.Error(c, http.StatusUnauthorized, "", err.Error())
// 		response.JsonErrorResponse(c, err.Error())
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"user_id": userID,
// 	})
// }
