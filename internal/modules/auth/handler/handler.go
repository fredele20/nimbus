package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/fredele20/nimbus/internal/modules/auth/context"
	authdto "github.com/fredele20/nimbus/internal/modules/auth/dto"
	authservice "github.com/fredele20/nimbus/internal/modules/auth/service"
	"github.com/fredele20/nimbus/internal/shared/http/response"
	"github.com/fredele20/nimbus/internal/shared/validation"
)

type Handler struct {
	service authservice.Service
}

func New(service authservice.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(c *gin.Context) {
	var req authdto.RegisterRequest

	if err := validation.Validate(c, &req); err != nil {
		response.JsonErrorResponse(c, err.Error())
		// response.JsonErrorResponse(c, err.Error())
		return
	}

	res, err := h.service.Register(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, authservice.ErrEmailExists) {
			response.JsonErrorResponse(c, err.Error())
			return
		}

		response.JsonErrorResponse(c, err.Error())
		return
	}

	response.JsonCreatedResponse(c, res, "")
}

func (h *Handler) Login(c *gin.Context) {
	var req authdto.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.JsonErrorResponse(c, err.Error())
		return
	}

	res, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, authservice.ErrInvalidCredentials) {
			// httperror.Handle(c, fmt.Errorf("%w: invalid email or password", httperror.ErrUnauthorized))
			return
		}

		response.JsonErrorResponse(c, err.Error())
		return
	}

	response.JsonSuccessResponse(c, res, "")
}

func (h *Handler) Refresh(c *gin.Context) {
	var req authdto.RefreshRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.JsonErrorResponse(c, err.Error())
		return
	}

	res, err := h.service.Refresh(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, authservice.ErrInvalidCredentials) {
			response.JsonErrorResponse(c, err.Error())
			return
		}

		response.JsonErrorResponse(c, err.Error())
		return
	}

	response.JsonSuccessResponse(c, res, "")
}

func (h *Handler) Logout(c *gin.Context) {
	var req authdto.RefreshRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.JsonErrorResponse(c, err.Error())
		return
	}

	err := h.service.Logout(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.JsonErrorResponse(c, err.Error())
		return
	}

	response.JsonSuccessResponse(c, nil, "")
}

func (h *Handler) Me(c *gin.Context) {
	userID, err := context.UserIDFromContext(c)
	if err != nil {
		// httperror.Error(c, http.StatusUnauthorized, "", err.Error())
		response.JsonErrorResponse(c, err.Error())
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
	})
}
