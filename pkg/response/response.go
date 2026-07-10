package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorBody struct {
	Error string `json:"error"`
}

func JSON(c *gin.Context, statusCode int, body any) {
	c.JSON(statusCode, body)
}

func ErrorMessage(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, ErrorBody{Error: message})
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}
