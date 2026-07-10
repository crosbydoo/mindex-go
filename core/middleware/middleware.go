package middleware

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"mindex-api/core/auth"
	"mindex-api/pkg/response"

	"github.com/gin-gonic/gin"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		slog.Info("request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration", time.Since(start).String(),
		)
	}
}

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				slog.Error("panic recovered", "error", recovered)
				response.InternalServer(c, "Internal server error")
				c.Abort()
			}
		}()
		c.Next()
	}
}

func CORS(origin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func Auth(adminPassword string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			response.Unauthorized(c, "Unauthorized")
			c.Abort()
			return
		}

		token := strings.TrimPrefix(header, "Bearer ")
		if !auth.VerifyToken(token, adminPassword) {
			response.Unauthorized(c, "Unauthorized")
			c.Abort()
			return
		}

		c.Next()
	}
}
