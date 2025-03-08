package main

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/mrkbwp/gotube/docs"
	"github.com/mrkbwp/gotube/internal/api/handlers"
	apiMiddleware "github.com/mrkbwp/gotube/internal/api/middleware"
	"github.com/mrkbwp/gotube/internal/infrastructure/repositories"
	"github.com/mrkbwp/gotube/internal/infrastructure/services"
	"github.com/mrkbwp/gotube/internal/infrastructure/storage"
	"github.com/mrkbwp/gotube/pkg/config"
	"github.com/mrkbwp/gotube/pkg/constants"
	"github.com/mrkbwp/gotube/pkg/jwt"
	"github.com/mrkbwp/gotube/pkg/kafka"
	"github.com/mrkbwp/gotube/pkg/logger"
	"github.com/mrkbwp/gotube/pkg/postgres"
	"github.com/mrkbwp/gotube/pkg/redis"
	"github.com/mrkbwp/gotube/pkg/validator"
	echoSwagger "github.com/swaggo/echo-swagger"

	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// @title GoTube API
// @version 1.0
// @description Api видеохостинга GoTube
// @host localhost:8111
// @BasePath /api/v1
func main() {
	// Контекст
	ctx := context.Background()

	// Инициализируем логгер
	log := logger.NewLogger(true)

	// Загружаем конфигурацию из переменных окружения
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config: %v", err)
	}

	// Инициализируем соединение с базой данных
	db, err := postgres.NewPostgresDB(cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to postgres: %v", err)
	}
	defer db.Close()

	// Инициализируем Redis
	redisClient, err := redis.NewRedisClient(cfg.Redis)
	if err != nil {
		log.Fatal("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	// Инициализируем MinIO клиент
	minioClient, err := storage.NewMinioClient(
		cfg.Minio.Endpoint,
		cfg.Storage.BaseURL,
		cfg.Minio.AccessKey,
		cfg.Minio.SecretKey,
		cfg.Minio.UseSSL,
		cfg.Storage.UseSSL,
	)
	if err != nil {
		log.Fatal("Failed to create MinIO client: %v", err)
	}

	if err := minioClient.EnsureBucketExists(ctx, constants.ThumbnailsBucket); err != nil {
		log.Fatal("Failed to create thumbnails bucket: %v", err)
	}

	// Инициализируем Kafka producer
	kafkaProducer, err := kafka.NewProducer(
		cfg.Kafka.Brokers,
		cfg.Kafka.VideoProcessingTopic,
	)
	if err != nil {
		log.Fatal("Failed to connect kafka: %v", err)
	}
	defer kafkaProducer.Close()

	// Инициализируем валидатор
	validator := validator.NewValidator()

	// Инициализируем сервисы безопасности
	jwtService := jwt.NewJWTService(
		cfg.Auth.AccessTokenSecret,
		cfg.Auth.RefreshTokenSecret,
		cfg.Auth.AccessTokenDuration,
		cfg.Auth.RefreshTokenDuration,
	)
	passwordService := jwt.NewPasswordService()

	// Инициализируем репозитории
	userRepo := repositories.NewUserRepository(db)
	tokenRepo := repositories.NewTokenRepository(db)
	videoRepo := repositories.NewVideoRepository(db)
	commentRepo := repositories.NewCommentRepository(db)
	categoryRepo := repositories.NewCategoryRepository(db)

	// Инициализируем бизнес-логику
	authService := services.NewAuthService(userRepo, tokenRepo, passwordService, jwtService)
	videoService := services.NewVideoService(videoRepo, minioClient, kafkaProducer, redisClient, cfg.Storage.ShardCount)
	commentService := services.NewCommentService(commentRepo, videoService)
	categoryService := services.NewCategoryService(categoryRepo, redisClient)

	// Инициализируем HTTP обработчики
	authHandler := handlers.NewAuthHandler(authService, validator)
	videoHandler := handlers.NewVideoHandler(videoService, validator)
	commentHandler := handlers.NewCommentHandler(commentService, validator)
	categoryHandler := handlers.NewCategoryHandler(categoryService)

	// Конвертация
	conversionService := services.NewConversionService(
		videoRepo,
		minioClient,
		"/tmp/video-conversion",
	)

	// Запуск очереди конвертации
	conversionService.StartConversionQueue()
	defer conversionService.StopConversionQueue()

	// Создаем Echo-сервер
	e := echo.New()

	// Настраиваем middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Добавляем аутентификационное middleware
	authMiddleware := apiMiddleware.AuthMiddleware(jwtService)

	// Роут для swagger UI
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// Настройка роутов
	apiV1 := e.Group("/api/v1")

	// Аутентификация
	apiV1.POST("/auth/register", authHandler.Register)
	apiV1.POST("/auth/login", authHandler.Login)
	apiV1.POST("/auth/refresh", authHandler.Refresh)
	apiV1.POST("/auth/logout", authHandler.Logout)

	// Публичные эндпоинты видео
	apiV1.GET("/videos/new", videoHandler.GetNewVideos)
	apiV1.GET("/videos/popular", videoHandler.GetPopularVideos)
	apiV1.GET("/videos/:code", videoHandler.GetVideoByCode)

	// Категории
	apiV1.GET("/categories", categoryHandler.GetCategories)

	// Комментарии (чтение)
	apiV1.GET("/videos/:code/comments", commentHandler.GetVideoComments)

	// Видео пользователя (чтение)
	apiV1.GET("/api/users/:user_id/videos", videoHandler.GetUserVideos)

	// Защищенные маршруты (требуют аутентификации)
	apiV1auth := apiV1
	apiV1auth.Use(authMiddleware)

	// Видео (операции записи)
	apiV1auth.POST("/videos", videoHandler.UploadVideo)
	apiV1auth.PUT("/videos/:code", videoHandler.UpdateVideo)
	apiV1auth.DELETE("/videos/:code", videoHandler.DeleteVideo)

	// Реакции на видео
	apiV1auth.POST("/videos/:id/like", videoHandler.LikeVideo)
	apiV1auth.POST("/videos/:id/dislike", videoHandler.DislikeVideo)

	// Комментарии (операции записи)
	apiV1auth.POST("/videos/:code/comments", commentHandler.AddComment)
	apiV1auth.DELETE("/comments/:id", commentHandler.DeleteComment)

	// Получение информации для юзера о видео
	apiV1auth.GET("/videos/user/:code", videoHandler.GetVideoUserByCode)

	// Запускаем сервер с graceful shutdown
	go func() {
		if err := e.Start(fmt.Sprintf(":%d", cfg.Server.Port)); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.Timeout)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Fatal("Server shutdown failed: %v", err)
	}

	log.Info("Server stopped gracefully")
}
