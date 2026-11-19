package controller

import (
	"errors"
	"net/http"

	authCtx "github.com/fredele20/nimbus/internal/modules/auth/context"
	"github.com/fredele20/nimbus/internal/modules/wallet/dto"
	"github.com/fredele20/nimbus/internal/modules/wallet/repository"
	"github.com/fredele20/nimbus/internal/modules/wallet/service"
	"github.com/fredele20/nimbus/internal/shared/http/response"
	"github.com/fredele20/nimbus/internal/shared/validation"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ControllerInterface interface {
	Create(*gin.Context)
	List(*gin.Context)
	GetByID(*gin.Context)
	Freeze(*gin.Context)
	Unfreeze(*gin.Context)
	AdminFreeze(*gin.Context)
	AdminUnfreeze(*gin.Context)
}

type controller struct {
	service      service.Service
	adminService service.AdminService
	repository   repository.Repository
}

func New(
	walletService service.Service,
	adminService service.AdminService,
) ControllerInterface {
	return &controller{
		service:      walletService,
		adminService: adminService,
	}
}

func (c *controller) Create(ctx *gin.Context) {
	userID, ok := authenticatedUserID(ctx)
	if !ok {
		return
	}

	var req dto.CreateWalletRequest
	if err := validation.Validate(ctx, &req); err != nil {
		response.JsonErrorResponse(ctx, err.Error())
		return
	}

	wallet, err := c.service.Create(ctx.Request.Context(), userID, req.Currency)
	if err != nil {
		writeError(ctx, err)
		return
	}

	response.JsonCreatedResponse(ctx, wallet, "successfully created wallet")
}

func (c *controller) List(ctx *gin.Context) {
	userID, ok := authenticatedUserID(ctx)
	if !ok {
		return
	}

	wallets, err := c.service.ListByUser(ctx.Request.Context(), userID)
	if err != nil {
		writeError(ctx, err)
		return
	}

	response.JsonSuccessResponse(ctx, wallets, "successfully fetch wallets")
}

func (c *controller) GetByID(ctx *gin.Context) {
	userID, ok := authenticatedUserID(ctx)
	if !ok {
		return
	}

	walletID, ok := walletIDFromParam(ctx)
	if !ok {
		return
	}

	wallet, err := c.service.GetByID(ctx.Request.Context(), walletID, userID)
	if err != nil {
		writeError(ctx, err)
		return
	}

	response.JsonSuccessResponse(ctx, wallet, "successfully fetch wallet by id")
}

func (c *controller) Freeze(ctx *gin.Context) {
	c.ownerRestriction(ctx, true)
}

func (c *controller) Unfreeze(ctx *gin.Context) {
	c.ownerRestriction(ctx, false)
}

func (c *controller) AdminFreeze(ctx *gin.Context) {
	adminID, ok := authenticatedUserID(ctx)
	if !ok {
		return
	}

	walletID, ok := walletIDFromParam(ctx)
	if !ok {
		return
	}

	var req dto.AdminFreezeRequest
	if err := validation.Validate(ctx, &req); err != nil {
		response.JsonErrorResponse(ctx, err.Error())
		return
	}

	err := c.adminService.FreezeWallet(ctx.Request.Context(), walletID, adminID, req.Reason)
	if err != nil {
		writeError(ctx, err)
		return
	}

	response.JsonSuccessResponse(ctx, nil, "admin freeze account successful")
}

func (c *controller) AdminUnfreeze(ctx *gin.Context) {
	adminID, ok := authenticatedUserID(ctx)
	if !ok {
		return
	}

	walletID, ok := walletIDFromParam(ctx)
	if !ok {
		return
	}

	err := c.adminService.UnfreezeWallet(ctx.Request.Context(), walletID, adminID)
	if err != nil {
		writeError(ctx, err)
		return
	}

	response.JsonSuccessResponse(ctx, nil, "admin unfreeze account successful")
}

func (c *controller) ownerRestriction(ctx *gin.Context, freeze bool) {
	userID, ok := authenticatedUserID(ctx)
	if !ok {
		return
	}

	walletID, ok := walletIDFromParam(ctx)
	if !ok {
		return
	}

	var err error
	var message string

	if freeze {
		err = c.service.FreezeOwnWallet(ctx.Request.Context(), walletID, userID)
		message = "wallet frozen successfully"
	} else {	
		err = c.service.UnfreezeOwnWallet(ctx.Request.Context(), walletID, userID)
		message = "wallet unfrozen successfully"
	}

	if err != nil {
		writeError(ctx, err)
		return
	}

	response.JsonSuccessResponse(ctx, nil, message)
}

func authenticatedUserID(c *gin.Context) (uuid.UUID, bool) {
	userID, err := authCtx.UserIDFromContext(c)
	if err != nil || userID == uuid.Nil {
		response.JsonStatusErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return uuid.Nil, false
	}

	return userID, true
}

func walletIDFromParam(c *gin.Context) (uuid.UUID, bool) {
	walletID, err := uuid.Parse(c.Param("id"))
	if err != nil || walletID == uuid.Nil {
		response.JsonErrorResponse(c, "invalid wallet id")
		return uuid.Nil, false
	}

	return walletID, true
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrWalletNotFound),
		errors.Is(err, gorm.ErrRecordNotFound):
		response.JsonNotFoundErrorResponse(c, "wallet not found")

	case errors.Is(err, service.ErrWalletAlreadyExists),
		errors.Is(err, service.ErrWalletRestrictionConflict):
		response.JsonStatusErrorResponse(c, http.StatusConflict, err.Error())

	case errors.Is(err, service.ErrInvalidCurrency),
		errors.Is(err, service.ErrInvalidWalletID),
		errors.Is(err, service.ErrInvalidUserID):
		response.JsonErrorResponse(c, err.Error())

	case errors.Is(err, service.ErrWalletAccessDenied):
		response.JsonStatusErrorResponse(c, http.StatusForbidden, "forbidden")

	default:
		response.JsonStatusErrorResponse(c, http.StatusInternalServerError, "something went wrong")
	}
}
