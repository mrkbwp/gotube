package handlers

import (
	"github.com/google/uuid"
	"github.com/mrkbwp/gotube/internal/api/requests"
	"github.com/mrkbwp/gotube/pkg/constants"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/mrkbwp/gotube/internal/api/responses"
	"github.com/mrkbwp/gotube/internal/domain/entity"
	"github.com/mrkbwp/gotube/internal/domain/services"
	"github.com/mrkbwp/gotube/pkg/pagination"
	"github.com/mrkbwp/gotube/pkg/validator"
)

// CommentHandler обработчик для API комментариев
type CommentHandler struct {
	commentService services.CommentService
	validator      *validator.Validator
}

// NewCommentHandler создает новый CommentHandler
func NewCommentHandler(commentService services.CommentService, validator *validator.Validator) *CommentHandler {
	return &CommentHandler{
		commentService: commentService,
		validator:      validator,
	}
}

// GetVideoComments возвращает комментарии к видео
// @Summary Получение комментариев к видео
// @Description Возвращает список комментариев к видео с пагинацией
// @Tags comments
// @Produce json
// @Param code path string true "ID видео"
// @Param page query int false "Номер страницы"
// @Param limit query int false "Количество на странице"
// @Success 200 {object} responses.PaginatedResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/videos/{code}/comments [get]
func (h *CommentHandler) GetVideoComments(c echo.Context) error {
	videoCode := c.Param("code")
	paginationParams := pagination.ExtractPaginationParams(c)
	ctx := c.Request().Context()

	comments, total, err := h.commentService.GetVideoComments(ctx, videoCode, paginationParams.Page, paginationParams.Limit)
	if err != nil {
		return responses.Error(c, http.StatusInternalServerError, "Не удалось получить комментарии")
	}

	return responses.JSON(c, http.StatusOK, responses.PaginatedResponse{
		Data:  comments,
		Page:  paginationParams.Page,
		Limit: paginationParams.Limit,
		Total: total,
	})
}

// AddComment добавляет новый комментарий к видео
// @Summary Добавление комментария
// @Description Добавляет новый комментарий к видео
// @Tags comments
// @Accept json
// @Produce json
// @Param id path string true "ID видео"
// @Param comment body requests.CommentRequest true "Данные комментария"
// @Security BearerAuth
// @Success 201 {object} entity.Comment
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/videos/{id}/comments [post]
func (h *CommentHandler) AddComment(c echo.Context) error {
	userID := c.Get("userID").(uuid.UUID)
	videoID := c.Param("id")

	var request requests.CommentRequest
	if err := c.Bind(&request); err != nil {
		return responses.Error(c, http.StatusBadRequest, "Ошибка в данных запроса")
	}

	if request.Text == "" {
		return responses.Error(c, http.StatusBadRequest, "Текст комментария обязателен")
	}

	var parentId uuid.UUID
	if request.ParentID != nil {
		parentId = uuid.MustParse(*request.ParentID)
	}

	ctx := c.Request().Context()
	comment := &entity.Comment{
		VideoID:  uuid.MustParse(videoID),
		UserID:   userID,
		Text:     request.Text,
		ParentID: &parentId,
	}

	err := h.commentService.AddComment(ctx, comment)
	if err != nil {
		return responses.Error(c, http.StatusInternalServerError, "Не удалось добавить комментарий")
	}

	return responses.JSON(c, http.StatusCreated, comment)
}

// DeleteComment удаляет комментарий
// @Summary Удаление комментария
// @Description Удаляет комментарий по ID
// @Tags comments
// @Produce json
// @Param id path string true "ID комментария"
// @Security BearerAuth
// @Success 200 {object} responses.SuccessResponse
// @Failure 403 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/comments/{id} [delete]
func (h *CommentHandler) DeleteComment(c echo.Context) error {
	userID := c.Get("userID").(uuid.UUID)
	commentID := c.Param("id")
	commentIDuuid := uuid.MustParse(commentID)
	ctx := c.Request().Context()

	comment, err := h.commentService.GetCommentByID(ctx, commentIDuuid)
	if err != nil {
		if err == constants.ErrNotFound {
			return responses.Error(c, http.StatusNotFound, "Комментарий не найден")
		}
		return responses.Error(c, http.StatusInternalServerError, "Ошибка получения комментария")
	}

	if comment.UserID != userID {
		return responses.Error(c, http.StatusForbidden, "Нет прав для удаления комментария")
	}

	err = h.commentService.DeleteComment(ctx, commentIDuuid)
	if err != nil {
		return responses.Error(c, http.StatusInternalServerError, "Ошибка удаления комментария")
	}

	return responses.Success(c, "Комментарий удален")
}

// UpdateComment обновляет комментарий
// @Summary Обновление комментария
// @Description Обновляет текст существующего комментария
// @Tags comments
// @Accept json
// @Produce json
// @Param id path string true "ID комментария"
// @Param comment body requests.CommentRequest true "Данные комментария"
// @Security BearerAuth
// @Success 200 {object} entity.Comment
// @Failure 400 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/comments/{id} [put]
func (h *CommentHandler) UpdateComment(c echo.Context) error {
	userID := c.Get("userID").(uuid.UUID)
	commentID := c.Param("id")
	commentIDuuid := uuid.MustParse(commentID)

	var request struct {
		Text string `json:"text"`
	}

	if err := c.Bind(&request); err != nil {
		return responses.Error(c, http.StatusBadRequest, "Ошибка в данных запроса")
	}

	if request.Text == "" {
		return responses.Error(c, http.StatusBadRequest, "Текст комментария обязателен")
	}

	ctx := c.Request().Context()
	comment, err := h.commentService.GetCommentByID(ctx, commentIDuuid)
	if err != nil {
		if err == constants.ErrNotFound {
			return responses.Error(c, http.StatusNotFound, "Комментарий не найден")
		}
		return responses.Error(c, http.StatusInternalServerError, "Ошибка получения комментария")
	}

	if comment.UserID != userID {
		return responses.Error(c, http.StatusForbidden, "Нет прав для редактирования комментария")
	}

	comment.Text = request.Text
	err = h.commentService.UpdateComment(ctx, comment)
	if err != nil {
		return responses.Error(c, http.StatusInternalServerError, "Ошибка обновления комментария")
	}

	return responses.JSON(c, http.StatusOK, comment)
}
