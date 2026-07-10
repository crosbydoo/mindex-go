package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Body is the standard API response envelope.
type Body struct {
	Code    int    `json:"code"`
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func Success(c *gin.Context, statusCode int, message string, data any) {
	c.JSON(statusCode, Body{
		Code:    statusCode,
		Status:  true,
		Message: message,
		Data:    data,
	})
}

func Error(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, Body{
		Code:    statusCode,
		Status:  false,
		Message: message,
		Data:    nil,
	})
}

func OK(c *gin.Context, message string, data any) {
	Success(c, http.StatusOK, message, data)
}

func Created(c *gin.Context, message string, data any) {
	Success(c, http.StatusCreated, message, data)
}

func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message)
}

func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, message)
}

func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message)
}

func MethodNotAllowed(c *gin.Context, message string) {
	Error(c, http.StatusMethodNotAllowed, message)
}

func ServiceUnavailable(c *gin.Context, message string) {
	Error(c, http.StatusServiceUnavailable, message)
}

func InternalServer(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, message)
}
