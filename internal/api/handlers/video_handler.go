package handlers

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/mrkbwp/gotube/internal/api/responses"
	"github.com/mrkbwp/gotube/internal/domain/services"
	"github.com/mrkbwp/gotube/pkg/pagination"
	"github.com/mrkbwp/gotube/pkg/validator"
	"net/http"
	"path/filepath"
	"strings"
)

// VideoHandler обработчик для видео-API
type VideoHandler struct {
	videoService services.VideoService
	validator    *validator.Validator
}

// NewVideoHandler создает новый VideoHandler
func NewVideoHandler(videoService services.VideoService, validator *validator.Validator) *VideoHandler {
	return &VideoHandler{
		videoService: videoService,
		validator:    validator,
	}
}

// UploadVideo загружает новое видео
// @Summary Загрузка видео
// @Description Загружает новое видео в систему, название и описание берутся из имени файла
// @Tags videos
// @Accept multipart/form-data
// @Produce json
// @Param video formData file true "Видеофайл"
// @Security BearerAuth
// @Success 201 {object} entity.Video
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/videos [post]
func (h *VideoHandler) UploadVideo(c echo.Context) error {
	userID := c.Get("userID").(uuid.UUID)

	// Получаем файл из запроса
	file, fileHeader, err := c.Request().FormFile("video")
	if err != nil {
		return responses.Error(c, http.StatusBadRequest, "Video file is required")
	}
	defer file.Close()

	// Извлекаем название и расширение файла
	filename := fileHeader.Filename
	ext := filepath.Ext(filename)
	title := strings.TrimSuffix(filename, ext) // Название без расширения

	// Описание оставляем пустым или используем то же имя файла
	description := ""

	ctx := c.Request().Context()

	// Вызываем сервис для загрузки видео
	video, err := h.videoService.UploadVideo(
		ctx,
		userID,
		uuid.MustParse("c6d76596-9407-6e93-d8ae-1f2103e3f33d"),
		file,
		filename,
		title,
		description,
	)
	if err != nil {
		return responses.Error(c, http.StatusInternalServerError, "Failed to upload video"+err.Error())
	}

	return responses.JSON(c, http.StatusCreated, video)
}

// GetVideoByCode возвращает информацию о видео по коду
// @Summary Получение видео по коду
// @Description Возвращает информацию о видео по его коду
// @Tags videos
// @Produce json
// @Param code path string true "Код видео"
// @Success 200 {object} entity.Video
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/videos/{code} [get]
func (h *VideoHandler) GetVideoByCode(c echo.Context) error {
	videoCode := c.Param("code")
	ctx := c.Request().Context()

	video, err := h.videoService.GetVideoByCode(ctx, videoCode)
	if err != nil {
		return responses.Error(c, http.StatusNotFound, "Video not found")
	}

	// Получаем список файлов с разными качествами
	files, err := h.videoService.GetVideoFiles(ctx, video)
	if err != nil {
		return responses.Error(c, http.StatusInternalServerError, "Failed to get video files"+err.Error())
	}

	video.Files = files

	userID, ok := c.Get("userID").(uuid.UUID)
	userIP := c.RealIP()
	if ok {
		go h.videoService.ViewVideo(ctx, video.ID, &userID, userIP)
	} else {
		go h.videoService.ViewVideo(ctx, video.ID, nil, userIP)
	}

	return responses.JSON(c, http.StatusOK, video)
}

// GetNewVideos возвращает список новых видео с пагинацией
// @Summary Список новых видео
// @Description Возвращает список новых видео с пагинацией
// @Tags videos
// @Produce json
// @Param page query int false "Номер страницы"
// @Param limit query int false "Количество на странице"
// @Success 200 {object} responses.PaginatedResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/videos/new [get]
func (h *VideoHandler) GetNewVideos(c echo.Context) error {
	paginationParams := pagination.ExtractPaginationParams(c)
	ctx := c.Request().Context()

	videos, total, err := h.videoService.GetNewVideos(ctx, paginationParams.Page, paginationParams.Limit)
	if err != nil {
		return responses.Error(c, http.StatusInternalServerError, "Failed to get videos")
	}

	return responses.JSON(c, http.StatusOK, responses.PaginatedResponse{
		Data:  videos,
		Page:  paginationParams.Page,
		Limit: paginationParams.Limit,
		Total: total,
	})
}

// GetPopularVideos возвращает список популярных видео
// @Summary Список популярных видео
// @Description Возвращает список популярных видео с пагинацией
// @Tags videos
// @Produce json
// @Param page query int false "Номер страницы"
// @Param limit query int false "Количество на странице"
// @Success 200 {object} responses.PaginatedResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/videos/popular [get]
func (h *VideoHandler) GetPopularVideos(c echo.Context) error {
	paginationParams := pagination.ExtractPaginationParams(c)
	ctx := c.Request().Context()

	videos, total, err := h.videoService.GetPopularVideos(ctx, paginationParams.Page, paginationParams.Limit)
	if err != nil {
		return responses.Error(c, http.StatusInternalServerError, "Failed to get videos")
	}

	return responses.JSON(c, http.StatusOK, responses.PaginatedResponse{
		Data:  videos,
		Page:  paginationParams.Page,
		Limit: paginationParams.Limit,
		Total: total,
	})
}

// LikeVideo обработчик для лайка видео
// @Summary Лайк видео
// @Description Добавляет лайк к видео от текущего пользователя
// @Tags videos
// @Produce json
// @Param id path string true "ID видео"
// @Security BearerAuth
// @Success 200 {object} responses.SuccessResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/videos/{id}/like [post]
func (h *VideoHandler) LikeVideo(c echo.Context) error {
	userID := c.Get("userID").(uuid.UUID)
	videoID := c.Param("id")
	ctx := c.Request().Context()

	err := h.videoService.LikeVideo(ctx, uuid.MustParse(videoID), userID)
	if err != nil {
		return responses.Error(c, http.StatusInternalServerError, "Failed to like video")
	}

	return responses.Success(c, "Video liked successfully")
}

// DislikeVideo обработчик для дизлайка видео
// @Summary Дизлайк видео
// @Description Добавляет дизлайк к видео от текущего пользователя
// @Tags videos
// @Produce json
// @Param id path string true "ID видео"
// @Security BearerAuth
// @Success 200 {object} responses.SuccessResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/videos/{id}/dislike [post]
func (h *VideoHandler) DislikeVideo(c echo.Context) error {
	userID := c.Get("userID").(uuid.UUID)
	videoID := c.Param("id")
	ctx := c.Request().Context()

	err := h.videoService.DislikeVideo(ctx, uuid.MustParse(videoID), userID)
	if err != nil {
		return responses.Error(c, http.StatusInternalServerError, "Failed to dislike video")
	}

	return responses.Success(c, "Video disliked successfully")
}

// GetUserVideos возвращает список видео пользователя
// @Summary Список видео пользователя
// @Description Возвращает список видео конкретного пользователя с пагинацией
// @Tags videos
// @Produce json
// @Param user_id path string true "ID пользователя"
// @Param page query int false "Номер страницы"
// @Param limit query int false "Количество на странице"
// @Success 200 {object} responses.PaginatedResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/users/{user_id}/videos [get]
func (h *VideoHandler) GetUserVideos(c echo.Context) error {
	userID := c.Param("user_id")
	paginationParams := pagination.ExtractPaginationParams(c)
	ctx := c.Request().Context()

	videos, total, err := h.videoService.GetUserVideos(ctx, uuid.MustParse(userID), paginationParams.Page, paginationParams.Limit)
	if err != nil {
		return responses.Error(c, http.StatusInternalServerError, "Failed to get user videos")
	}

	return responses.JSON(c, http.StatusOK, responses.PaginatedResponse{
		Data:  videos,
		Page:  paginationParams.Page,
		Limit: paginationParams.Limit,
		Total: total,
	})
}

// UpdateVideo обновляет информацию о видео
// @Summary Обновление видео
// @Description Обновляет информацию о существующем видео
// @Tags videos
// @Accept json
// @Produce json
// @Param id path string true "ID видео"
// @Param video body requests.UpdateVideoRequest true "Данные для обновления видео"
// @Security BearerAuth
// @Success 200 {object} entity.Video
// @Failure 400 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/videos/{code} [put]
func (h *VideoHandler) UpdateVideo(c echo.Context) error {
	userID := c.Get("userID").(uuid.UUID)
	videoCode := c.Param("code")
	ctx := c.Request().Context()

	video, err := h.videoService.GetVideoByCode(ctx, videoCode)
	if err != nil {
		return responses.Error(c, http.StatusNotFound, "Video not found")
	}

	if video.UserID != userID {
		return responses.Error(c, http.StatusForbidden, "You don't have permission to update this video")
	}

	var request struct {
		Title       string    `json:"title"`
		Description string    `json:"description"`
		CategoryID  uuid.UUID `json:"category_id"`
	}

	if err := c.Bind(&request); err != nil {
		return responses.Error(c, http.StatusBadRequest, "Invalid request data")
	}

	if request.Title == "" {
		return responses.Error(c, http.StatusBadRequest, "Title is required")
	}

	updatedVideo, err := h.videoService.UpdateVideo(
		ctx,
		video.ID,
		request.CategoryID,
		request.Title,
		request.Description,
	)
	if err != nil {
		return responses.Error(c, http.StatusInternalServerError, "Failed to update video")
	}

	return responses.JSON(c, http.StatusOK, updatedVideo)
}

// DeleteVideo удаляет видео
// @Summary Удаление видео
// @Description Удаляет видео по ID
// @Tags videos
// @Produce json
// @Param id path string true "ID видео"
// @Security BearerAuth
// @Success 200 {object} responses.SuccessResponse
// @Failure 403 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/videos/{id} [delete]
func (h *VideoHandler) DeleteVideo(c echo.Context) error {
	userID := c.Get("userID").(uuid.UUID)
	videoID := c.Param("id")
	videoIDuuid := uuid.MustParse(videoID)
	ctx := c.Request().Context()

	video, err := h.videoService.GetVideoByID(ctx, videoIDuuid)
	if err != nil {
		return responses.Error(c, http.StatusNotFound, "Video not found")
	}

	if video.UserID != userID {
		return responses.Error(c, http.StatusForbidden, "You don't have permission to delete this video")
	}

	err = h.videoService.DeleteVideo(ctx, videoIDuuid)
	if err != nil {
		return responses.Error(c, http.StatusInternalServerError, "Failed to delete video")
	}

	return responses.Success(c, "Video deleted successfully")
}
