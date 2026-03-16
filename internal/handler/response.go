package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response is the unified JSON envelope.
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// Success sends a 200 response with code 0.
func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// Error sends an HTTP response with the given status and business error code.
func Error(c *gin.Context, httpStatus int, code int, message string) {
	c.JSON(httpStatus, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// ErrorWithData sends an HTTP response with the given status, error code, and data payload.
func ErrorWithData(c *gin.Context, httpStatus int, code int, message string, data any) {
	c.JSON(httpStatus, Response{
		Code:    code,
		Message: message,
		Data:    data,
	})
}
