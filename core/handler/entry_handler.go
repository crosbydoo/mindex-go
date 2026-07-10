package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"mindex-api/core/domain"
	"mindex-api/core/service"
	"mindex-api/pkg/response"

	"github.com/gin-gonic/gin"
)

type EntryHandler struct {
	service service.EntryService
}

func NewEntryHandler(s service.EntryService) *EntryHandler {
	return &EntryHandler{service: s}
}

func (h *EntryHandler) List(c *gin.Context) {
	entries, err := h.service.List(c.Request.Context())
	if err != nil {
		slog.Error("list entries failed", "error", err)
		response.ErrorMessage(c, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.JSON(c, http.StatusOK, entries)
}

func (h *EntryHandler) Create(c *gin.Context) {
	var input domain.EntryInput

	if err := c.ShouldBindJSON(&input); err != nil {
		response.ErrorMessage(c, http.StatusBadRequest, "Invalid entry payload")
		return
	}

	entry, err := h.service.Create(c.Request.Context(), input)
	if errors.Is(err, service.ErrInvalidPayload) {
		response.ErrorMessage(c, http.StatusBadRequest, "Invalid entry payload")
		return
	}
	if err != nil {
		slog.Error("create entry failed", "error", err)
		response.ErrorMessage(c, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.JSON(c, http.StatusCreated, entry)
}

func (h *EntryHandler) Update(c *gin.Context) {
	id, err := parseEntryID(c)
	if err != nil {
		response.ErrorMessage(c, http.StatusBadRequest, "Invalid entry id")
		return
	}

	var input domain.EntryInput

	if err := c.ShouldBindJSON(&input); err != nil {
		response.ErrorMessage(c, http.StatusBadRequest, "Invalid entry payload")
		return
	}

	entry, err := h.service.Update(c.Request.Context(), id, input)
	if errors.Is(err, service.ErrInvalidEntryID) {
		response.ErrorMessage(c, http.StatusBadRequest, "Invalid entry id")
		return
	}
	if errors.Is(err, service.ErrInvalidPayload) {
		response.ErrorMessage(c, http.StatusBadRequest, "Invalid entry payload")
		return
	}
	if errors.Is(err, service.ErrEntryNotFound) {
		response.ErrorMessage(c, http.StatusNotFound, "Entry not found")
		return
	}
	if err != nil {
		slog.Error("update entry failed", "error", err)
		response.ErrorMessage(c, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.JSON(c, http.StatusOK, entry)
}

func (h *EntryHandler) Delete(c *gin.Context) {
	id, err := parseEntryID(c)
	if err != nil {
		response.ErrorMessage(c, http.StatusBadRequest, "Invalid entry id")
		return
	}

	err = h.service.Delete(c.Request.Context(), id)
	if errors.Is(err, service.ErrInvalidEntryID) {
		response.ErrorMessage(c, http.StatusBadRequest, "Invalid entry id")
		return
	}
	if errors.Is(err, service.ErrEntryNotFound) {
		response.ErrorMessage(c, http.StatusNotFound, "Entry not found")
		return
	}
	if err != nil {
		slog.Error("delete entry failed", "error", err)
		response.ErrorMessage(c, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.NoContent(c)
}

func parseEntryID(c *gin.Context) (int64, error) {
	idStr := c.Query("id")
	if idStr == "" {
		return 0, errors.New("missing id")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid id")
	}

	return id, nil
}
