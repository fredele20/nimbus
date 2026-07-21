package httperror

import (
	"errors"
)

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("conflict")
	ErrValidation   = errors.New("validation")
)

// type ErrorResponse struct {
// 	Success bool   `json:"success"`
// 	Status int
// 	Code    string `json:"code"`
// 	Message string `json:"message"`
// }

// func (e *ErrorResponse) DefaultError(c *gin.Context) {
// 	status := e.Status

// 	if status == 0 {
// 		status = http.StatusBadR
// 	}
// }

// func Error(c *gin.Context, status int, code string, message string) {
// 	c.JSON(status, ErrorResponse{
// 		Success: false,
// 		Code:    code,
// 		Status: status,
// 		Message: message,
// 	})
// }

// func ErrorBadRequest(c *gin.Context, message string) {
// 	response := DefaultError(c)
// 	response.Message = message
// 	c.JSON(http.StatusBadRequest, response)
// }
