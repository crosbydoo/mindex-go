package handler

import (
	"net/http"

	"mindex-api/pkg/response"

	"github.com/gin-gonic/gin"
)

func Health(c *gin.Context) {
	response.JSON(c, http.StatusOK, gin.H{"status": "ok"})
}
