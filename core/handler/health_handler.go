package handler

import (
	"mindex-api/pkg/response"

	"github.com/gin-gonic/gin"
)

func Health(c *gin.Context) {
	response.OK(c, "Server is healthy", gin.H{"status": "ok"})
}
