package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"mindex-api/core/service"
	"mindex-api/pkg/response"

	"github.com/gin-gonic/gin"
)

type LoginHandler struct {
	service service.LoginService
}

func NewLoginHandler(s service.LoginService) *LoginHandler {
	return &LoginHandler{service: s}
}

func (h *LoginHandler) Login(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		response.MethodNotAllowed(c, "Method not allowed")
		return
	}

	var body struct {
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	token, err := h.service.Login(body.Password)
	if errors.Is(err, service.ErrAdminNotConfigured) {
		response.ServiceUnavailable(c, "ADMIN_PASSWORD is not configured on the server")
		return
	}
	if errors.Is(err, service.ErrInvalidPassword) {
		response.Unauthorized(c, "Invalid password")
		return
	}
	if err != nil {
		slog.Error("login failed", "error", err)
		response.InternalServer(c, "Internal server error")
		return
	}

	response.OK(c, "Login successful", gin.H{"token": token})
}

func (h *LoginHandler) Logout(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		response.MethodNotAllowed(c, "Method not allowed")
		return
	}

	if err := h.service.Logout(); err != nil {
		slog.Error("logout failed", "error", err)
		response.InternalServer(c, "Internal server error")
		return
	}

	response.OK(c, "Logout successful", nil)
}
