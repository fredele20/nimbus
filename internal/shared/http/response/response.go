package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Status  int    `json:"status"`
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func (r *Response) DefaultResponse() {
  status := r.Status

  if status == 0 {
    r.Status = http.StatusBadRequest
  }
  if r.Message == "" && status == 0 {
    r.Message = "Bad request"
  }
  if r.Data == nil {
    r.Data = map[string]interface{}{}
  }
}

func JsonResponse(ctx *gin.Context, r *Response) {
  r.DefaultResponse()

  status := r.Status
  response := gin.H{"succes": r.Success, "message": r.Message, "data": r.Data}

  ctx.JSON(status, response)
}

func JsonSuccessResponse(ctx *gin.Context, params any, message string) {
  response := Response{Status: http.StatusOK, Success: true, Message: message, Data: params}
  JsonResponse(ctx, &response)
}

func JsonCreatedResponse(ctx *gin.Context, params any, message string) {
  response := Response{Status: http.StatusCreated, Success: true, Message: message, Data: params}
  JsonResponse(ctx, &response)
}

func JsonErrorResponse(ctx *gin.Context, message string) {
  response := Response{Message: message}
  JsonResponse(ctx, &response)
}

func JsonStatusErrorResponse(ctx *gin.Context, status int, message string) {
  response := Response{Message: message, Status: status}
  JsonResponse(ctx, &response)
}

