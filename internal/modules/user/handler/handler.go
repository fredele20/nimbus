package handler

import (
	"net/http"

	authctx "github.com/fredele20/nimbus/internal/modules/auth/context"
	"github.com/fredele20/nimbus/internal/modules/user/dto"
	"github.com/fredele20/nimbus/internal/modules/user/model"
	"github.com/fredele20/nimbus/internal/modules/user/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service service.Service
}

func New(service service.Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) GetProfile(c *gin.Context) {
	user, _ := authctx.UserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	res, err := h.service.GetProfile(c.Request.Context(), user.ID.String())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	user, _ := authctx.UserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	var req dto.UpdateProfileRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := h.service.UpdateProfile(c.Request.Context(), user.ID.String(), req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) ListUsers(c *gin.Context) {
	users, err := h.service.ListUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, users)
}

func (h *Handler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateUserStatusRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := h.service.UpdateStatus(c.Request.Context(), id, model.Status(req.Status)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}
