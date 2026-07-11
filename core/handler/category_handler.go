package handler

import (
	"errors"
	"log/slog"
	"strconv"

	"mindex-api/core/domain"
	"mindex-api/core/service"
	"mindex-api/pkg/response"

	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	service service.CategoryService
}

func NewCategoryHandler(s service.CategoryService) *CategoryHandler {
	return &CategoryHandler{service: s}
}

func (h *CategoryHandler) List(c *gin.Context) {
	result, err := h.service.List(c.Request.Context())
	if err != nil {
		slog.Error("list category items failed", "error", err)
		response.InternalServer(c, "Internal server error")
		return
	}
	response.OK(c, "Category list retrieved successfully", result)
}

func (h *CategoryHandler) Create(c *gin.Context) {
	var input domain.CategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid category payload")
		return
	}

	item, err := h.service.Create(c.Request.Context(), input)
	if errors.Is(err, service.ErrInvalidCategoryPayload) {
		response.BadRequest(c, "Invalid category payload")
		return
	}
	if errors.Is(err, service.ErrCategoryExists) {
		response.Conflict(c, "Category already exists")
		return
	}
	if err != nil {
		slog.Error("create category failed", "error", err)
		response.InternalServer(c, "Internal server error")
		return
	}

	response.Created(c, "Category created successfully", item)
}

func (h *CategoryHandler) Update(c *gin.Context) {
	id, err := parseCategoryID(c)
	if err != nil {
		response.BadRequest(c, "Invalid category id")
		return
	}

	var input domain.CategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid category payload")
		return
	}

	item, err := h.service.Update(c.Request.Context(), id, input)
	if errors.Is(err, service.ErrInvalidCategoryID) {
		response.BadRequest(c, "Invalid category id")
		return
	}
	if errors.Is(err, service.ErrInvalidCategoryPayload) {
		response.BadRequest(c, "Invalid category payload")
		return
	}
	if errors.Is(err, service.ErrCategoryNotFound) {
		response.NotFound(c, "Category not found")
		return
	}
	if errors.Is(err, service.ErrCategoryExists) {
		response.Conflict(c, "Category already exists")
		return
	}
	if err != nil {
		slog.Error("update category failed", "error", err)
		response.InternalServer(c, "Internal server error")
		return
	}

	response.OK(c, "Category updated successfully", item)
}

func (h *CategoryHandler) Delete(c *gin.Context) {
	id, err := parseCategoryID(c)
	if err != nil {
		response.BadRequest(c, "Invalid category id")
		return
	}

	err = h.service.Delete(c.Request.Context(), id)
	if errors.Is(err, service.ErrInvalidCategoryID) {
		response.BadRequest(c, "Invalid category id")
		return
	}
	if errors.Is(err, service.ErrCategoryNotFound) {
		response.NotFound(c, "Category not found")
		return
	}
	if errors.Is(err, service.ErrCategoryInUse) {
		response.Conflict(c, "Category is still used by entries")
		return
	}
	if err != nil {
		slog.Error("delete category failed", "error", err)
		response.InternalServer(c, "Internal server error")
		return
	}

	response.OK(c, "Category deleted successfully", nil)
}

func parseCategoryID(c *gin.Context) (int64, error) {
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
