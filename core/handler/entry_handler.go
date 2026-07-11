package handler

import (
	"errors"
	"log/slog"
	"strconv"
	"strings"

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
	page, limit, err := parsePagination(c)
	if err != nil {
		response.BadRequest(c, "Invalid pagination")
		return
	}

	archived, err := parseArchivedFilter(c)
	if err != nil {
		response.BadRequest(c, "Invalid archived filter")
		return
	}

	category := strings.TrimSpace(c.Query("category"))
	result, err := h.service.List(c.Request.Context(), domain.ListFilter{
		Page:     page,
		Limit:    limit,
		Category: category,
		Archived: archived,
	})
	if errors.Is(err, service.ErrInvalidCategory) {
		response.BadRequest(c, "Invalid category")
		return
	}
	if err != nil {
		slog.Error("list entries failed", "error", err)
		response.InternalServer(c, "Internal server error")
		return
	}

	response.OK(c, "Entries retrieved successfully", result)
}

func (h *EntryHandler) ListByCategories(c *gin.Context) {
	page, limit, err := parsePagination(c)
	if err != nil {
		response.BadRequest(c, "Invalid pagination")
		return
	}

	archived, err := parseArchivedFilter(c)
	if err != nil {
		response.BadRequest(c, "Invalid archived filter")
		return
	}

	result, err := h.service.ListByCategories(c.Request.Context(), page, limit, archived)
	if err != nil {
		slog.Error("list categories failed", "error", err)
		response.InternalServer(c, "Internal server error")
		return
	}

	response.OK(c, "Categories retrieved successfully", result)
}

func (h *EntryHandler) Create(c *gin.Context) {
	var input domain.EntryInput

	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid entry payload")
		return
	}

	entry, err := h.service.Create(c.Request.Context(), input)
	if errors.Is(err, service.ErrInvalidPayload) {
		response.BadRequest(c, "Invalid entry payload")
		return
	}
	if err != nil {
		slog.Error("create entry failed", "error", err)
		response.InternalServer(c, "Internal server error")
		return
	}

	response.Created(c, "Entry created successfully", entry)
}

func (h *EntryHandler) Update(c *gin.Context) {
	id, err := parseEntryID(c)
	if err != nil {
		response.BadRequest(c, "Invalid entry id")
		return
	}

	var input domain.EntryInput

	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid entry payload")
		return
	}

	entry, err := h.service.Update(c.Request.Context(), id, input)
	if errors.Is(err, service.ErrInvalidEntryID) {
		response.BadRequest(c, "Invalid entry id")
		return
	}
	if errors.Is(err, service.ErrInvalidPayload) {
		response.BadRequest(c, "Invalid entry payload")
		return
	}
	if errors.Is(err, service.ErrEntryNotFound) {
		response.NotFound(c, "Entry not found")
		return
	}
	if err != nil {
		slog.Error("update entry failed", "error", err)
		response.InternalServer(c, "Internal server error")
		return
	}

	response.OK(c, "Entry updated successfully", entry)
}

func (h *EntryHandler) Delete(c *gin.Context) {
	id, err := parseEntryID(c)
	if err != nil {
		response.BadRequest(c, "Invalid entry id")
		return
	}

	err = h.service.Delete(c.Request.Context(), id)
	if errors.Is(err, service.ErrInvalidEntryID) {
		response.BadRequest(c, "Invalid entry id")
		return
	}
	if errors.Is(err, service.ErrEntryNotFound) {
		response.NotFound(c, "Entry not found")
		return
	}
	if err != nil {
		slog.Error("delete entry failed", "error", err)
		response.InternalServer(c, "Internal server error")
		return
	}

	response.OK(c, "Entry deleted successfully", nil)
}

func (h *EntryHandler) Archive(c *gin.Context) {
	id, err := parseEntryID(c)
	if err != nil {
		response.BadRequest(c, "Invalid entry id")
		return
	}

	entry, err := h.service.Archive(c.Request.Context(), id)
	if errors.Is(err, service.ErrInvalidEntryID) {
		response.BadRequest(c, "Invalid entry id")
		return
	}
	if errors.Is(err, service.ErrEntryNotFound) {
		response.NotFound(c, "Entry not found")
		return
	}
	if err != nil {
		slog.Error("archive entry failed", "error", err)
		response.InternalServer(c, "Internal server error")
		return
	}

	response.OK(c, "Entry archived successfully", entry)
}

func (h *EntryHandler) Unarchive(c *gin.Context) {
	id, err := parseEntryID(c)
	if err != nil {
		response.BadRequest(c, "Invalid entry id")
		return
	}

	entry, err := h.service.Unarchive(c.Request.Context(), id)
	if errors.Is(err, service.ErrInvalidEntryID) {
		response.BadRequest(c, "Invalid entry id")
		return
	}
	if errors.Is(err, service.ErrEntryNotFound) {
		response.NotFound(c, "Entry not found")
		return
	}
	if err != nil {
		slog.Error("unarchive entry failed", "error", err)
		response.InternalServer(c, "Internal server error")
		return
	}

	response.OK(c, "Entry unarchived successfully", entry)
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

func parsePagination(c *gin.Context) (int, int, error) {
	page := domain.DefaultPage
	limit := domain.DefaultLimit

	if pageStr := c.Query("page"); pageStr != "" {
		parsed, err := strconv.Atoi(pageStr)
		if err != nil || parsed < 1 {
			return 0, 0, errors.New("invalid page")
		}
		page = parsed
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		parsed, err := strconv.Atoi(limitStr)
		if err != nil || parsed < 1 {
			return 0, 0, errors.New("invalid limit")
		}
		limit = parsed
	}

	page, limit = domain.NormalizePagination(page, limit)
	return page, limit, nil
}

func parseArchivedFilter(c *gin.Context) (domain.ArchiveScope, error) {
	scope, err := domain.ParseArchiveScope(c.Query("archived"))
	if err != nil {
		return "", err
	}
	return scope, nil
}
