package handlers

import (
	"github.com/google/uuid"
	"github.com/mrkbwp/gotube/pkg/constants"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/mrkbwp/gotube/internal/api/responses"
	"github.com/mrkbwp/gotube/internal/domain/entity"
	"github.com/mrkbwp/gotube/internal/domain/services"
)

// CategoryHandler обработчик для API категорий
type CategoryHandler struct {
	categoryService services.CategoryService
}

// NewCategoryHandler создает новый CategoryHandler
func NewCategoryHandler(categoryService services.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
	}
}

// GetCategories возвращает список всех категорий
// @Summary Список категорий
// @Description Возвращает список всех доступных категорий
// @Tags categories
// @Produce json
// @Success 200 {array} entity.Category
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/categories [get]
func (h *CategoryHandler) GetCategories(c echo.Context) error {
	ctx := c.Request().Context()

	categories, err := h.categoryService.GetAllCategories(ctx)
	if err != nil {
		return responses.Error(c, http.StatusInternalServerError, "Failed to get categories")
	}

	return responses.JSON(c, http.StatusOK, categories)
}

// GetCategoryByID возвращает категорию по ID
// @Summary Получение категории
// @Description Возвращает категорию по её ID
// @Tags categories
// @Produce json
// @Param id path string true "ID категории"
// @Success 200 {object} entity.Category
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/categories/{id} [get]
func (h *CategoryHandler) GetCategoryByID(c echo.Context) error {
	categoryID := c.Param("id")
	categoryIDuuid := uuid.MustParse(categoryID)
	ctx := c.Request().Context()

	category, err := h.categoryService.GetCategoryByID(ctx, categoryIDuuid)
	if err != nil {
		if err == constants.ErrNotFound {
			return responses.Error(c, http.StatusNotFound, "Category not found")
		}
		return responses.Error(c, http.StatusInternalServerError, "Failed to get category")
	}

	return responses.JSON(c, http.StatusOK, category)
}

// CreateCategory создает новую категорию
// @Summary Создание категории
// @Description Создает новую категорию (только для администраторов)
// @Tags categories
// @Accept json
// @Produce json
// @Param category body entity.Category true "Данные категории"
// @Security BearerAuth
// @Success 201 {object} entity.Category
// @Failure 400 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/admin/categories [post]
func (h *CategoryHandler) CreateCategory(c echo.Context) error {
	var category entity.Category
	if err := c.Bind(&category); err != nil {
		return responses.Error(c, http.StatusBadRequest, "Invalid request data")
	}

	if category.Name == "" {
		return responses.Error(c, http.StatusBadRequest, "Category name is required")
	}

	ctx := c.Request().Context()
	err := h.categoryService.CreateCategory(ctx, &category)
	if err != nil {
		return responses.Error(c, http.StatusInternalServerError, "Failed to create category")
	}

	return responses.JSON(c, http.StatusCreated, category)
}

// UpdateCategory обновляет категорию
// @Summary Обновление категории
// @Description Обновляет существующую категорию (только для администраторов)
// @Tags categories
// @Accept json
// @Produce json
// @Param id path string true "ID категории"
// @Param category body entity.Category true "Данные категории"
// @Security BearerAuth
// @Success 200 {object} entity.Category
// @Failure 400 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/admin/categories/{id} [put]
func (h *CategoryHandler) UpdateCategory(c echo.Context) error {
	categoryID := c.Param("id")
	categoryIDuuid := uuid.MustParse(categoryID)

	var category entity.Category
	if err := c.Bind(&category); err != nil {
		return responses.Error(c, http.StatusBadRequest, "Invalid request data")
	}

	category.ID = uuid.MustParse(categoryID)

	if category.Name == "" {
		return responses.Error(c, http.StatusBadRequest, "Category name is required")
	}

	ctx := c.Request().Context()

	// Проверяем существование категории
	_, err := h.categoryService.GetCategoryByID(ctx, categoryIDuuid)
	if err != nil {
		if err == constants.ErrNotFound {
			return responses.Error(c, http.StatusNotFound, "Category not found")
		}
		return responses.Error(c, http.StatusInternalServerError, "Failed to get category")
	}

	err = h.categoryService.UpdateCategory(ctx, &category)
	if err != nil {
		return responses.Error(c, http.StatusInternalServerError, "Failed to update category")
	}

	return responses.JSON(c, http.StatusOK, category)
}

// DeleteCategory удаляет категорию
// @Summary Удаление категории
// @Description Удаляет категорию по ID (только для администраторов)
// @Tags categories
// @Produce json
// @Param id path string true "ID категории"
// @Security BearerAuth
// @Success 200 {object} responses.SuccessResponse
// @Failure 403 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/admin/categories/{id} [delete]
func (h *CategoryHandler) DeleteCategory(c echo.Context) error {
	ctx := c.Request().Context()
	categoryID := c.Param("id")
	categoryIDuuid := uuid.MustParse(categoryID)

	// Проверяем существование категории
	_, err := h.categoryService.GetCategoryByID(ctx, categoryIDuuid)
	if err != nil {
		if err == constants.ErrNotFound {
			return responses.Error(c, http.StatusNotFound, "Category not found")
		}
		return responses.Error(c, http.StatusInternalServerError, "Failed to get category")
	}

	err = h.categoryService.DeleteCategory(ctx, categoryIDuuid)
	if err != nil {
		return responses.Error(c, http.StatusInternalServerError, "Failed to delete category")
	}

	return responses.Success(c, "Category deleted successfully")
}
