package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/mrkbwp/gotube/pkg/constants"
	"github.com/mrkbwp/gotube/pkg/sqlutil"
	"reflect"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/mrkbwp/gotube/internal/domain/entity"
	"github.com/mrkbwp/gotube/internal/domain/repositories"
)

type VideoRepository struct {
	db *sqlx.DB
}

func NewVideoRepository(db *sqlx.DB) repositories.VideoRepository {
	return &VideoRepository{db: db}
}

func (r *VideoRepository) Create(ctx context.Context, video *entity.Video) error {
	// Получаем список полей из структуры Video
	fields, err := sqlutil.GetFields(video)
	if err != nil {
		return fmt.Errorf("failed to get fields: %w", err)
	}

	// Получаем значения полей
	values, err := sqlutil.GetValues(video)
	if err != nil {
		return fmt.Errorf("failed to get values: %w", err)
	}

	// Создаем SQL-запрос через Squirrel
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := builder.
		Insert(constants.VideosTable).
		Columns(fields...).
		Values(values...). // Используем распакованные значения
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	// Выполняем запрос
	var id uuid.UUID
	err = r.db.QueryRowContext(ctx, query, args...).Scan(&id)
	if err != nil {
		return fmt.Errorf("failed to create video: %w", err)
	}

	video.ID = id
	return nil
}

func (r *VideoRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Video, error) {
	// Получаем список полей из структуры Video
	fields, err := sqlutil.GetFields(&entity.Video{})
	if err != nil {
		return nil, fmt.Errorf("failed to get fields: %w", err)
	}

	// Создаем SQL-запрос через Squirrel
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := builder.
		Select(fields...).
		From(constants.VideosTable).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	// Выполняем запрос
	video := &entity.Video{}
	err = r.db.GetContext(ctx, video, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, constants.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get video: %w", err)
	}

	return video, nil
}

func (r *VideoRepository) GetByCode(ctx context.Context, code string) (*entity.Video, error) {
	// Получаем список полей из структуры Video
	fields, err := sqlutil.GetFields(&entity.Video{})
	if err != nil {
		return nil, fmt.Errorf("failed to get fields: %w", err)
	}

	// Создаем SQL-запрос через Squirrel
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := builder.
		Select(fields...).
		From(constants.VideosTable).
		Where("video_code = ?", code).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	// Выполняем запрос
	video := &entity.Video{}
	err = r.db.GetContext(ctx, video, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, constants.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get video: %w", err)
	}

	return video, nil
}

func (r *VideoRepository) GetNewVideos(ctx context.Context, page, limit int) ([]*entity.Video, int64, error) {
	fields, err := sqlutil.GetFields(&entity.Video{})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get fields: %w", err)
	}

	sb := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	query := sb.
		Select(fields...).
		From(constants.VideosTable).
		Where("status = ?", constants.VideoStatusReady).
		Where("deleted_at IS NULL").
		Where("is_blocked = ?", false).
		Where("is_private = ?", false).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64((page - 1) * limit))

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build query: %w", err)
	}

	videos := []*entity.Video{}
	err = r.db.SelectContext(ctx, &videos, sqlQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get videos: %w", err)
	}

	// Подсчёт общего количества
	countQuery := sb.
		Select("COUNT(*)").
		From(constants.VideosTable).
		Where("status = ?", constants.VideoStatusReady).
		Where("deleted_at IS NULL").
		Where("is_blocked = ?", false).
		Where("is_private = ?", false)

	countSql, countArgs, err := countQuery.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build count query: %w", err)
	}

	var total int64
	err = r.db.GetContext(ctx, &total, countSql, countArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}

	return videos, total, nil
}

func (r *VideoRepository) GetPopularVideos(ctx context.Context, page, limit int) ([]*entity.Video, int64, error) {
	fields, err := sqlutil.GetFields(&entity.Video{})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get fields: %w", err)
	}

	sb := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	query := sb.
		Select(fields...).
		From(constants.VideosTable).
		Where("status = ?", constants.VideoStatusReady).
		Where("deleted_at IS NULL").
		Where("is_blocked = ?", false).
		Where("is_private = ?", false).
		OrderBy("views DESC, likes DESC").
		Limit(uint64(limit)).
		Offset(uint64((page - 1) * limit))

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build query: %w", err)
	}

	videos := []*entity.Video{}
	err = r.db.SelectContext(ctx, &videos, sqlQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get videos: %w", err)
	}

	// Подсчёт общего количества
	countQuery := sb.
		Select("COUNT(*)").
		From(constants.VideosTable).
		Where("status = ?", constants.VideoStatusReady).
		Where("deleted_at IS NULL").
		Where("is_blocked = ?", false).
		Where("is_private = ?", false)

	countSql, countArgs, err := countQuery.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build count query: %w", err)
	}

	var total int64
	err = r.db.GetContext(ctx, &total, countSql, countArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}

	return videos, total, nil
}

func (r *VideoRepository) GetUserVideos(ctx context.Context, userID uuid.UUID, page, limit int) ([]*entity.Video, int64, error) {
	fields, err := sqlutil.GetFields(&entity.Video{}) // Получаем поля из структуры
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get fields: %w", err)
	}

	sb := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar) // Указываем формат плейсхолдеров

	offset := (page - 1) * limit

	// Построение SELECT-запроса через Squirrel
	query := sb.
		Select(fields...). // Динамически добавляем поля
		From(constants.VideosTable).
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset))

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build query: %w", err)
	}

	videos := []*entity.Video{}
	err = r.db.SelectContext(ctx, &videos, sqlQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get videos: %w", err)
	}

	// Построение COUNT-запроса
	countQuery := sb.
		Select("COUNT(*)").
		From(constants.VideosTable).
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL")

	countSql, countArgs, err := countQuery.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build count query: %w", err)
	}

	var total int64
	err = r.db.GetContext(ctx, &total, countSql, countArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}

	return videos, total, nil
}

func (r *VideoRepository) Update(ctx context.Context, video *entity.Video) error {
	sb := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	query, args, err := sb.
		Update(constants.VideosTable).
		Set("title", video.Title).
		Set("description", video.Description).
		Set("category_id", video.CategoryID).
		Set("updated_at", time.Now()).
		Where("id = ?", video.ID).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update video: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return constants.ErrNotFound
	}

	return nil
}

func (r *VideoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	sb := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	query, args, err := sb.
		Update(constants.VideosTable).
		Set("deleted_at", time.Now()).
		Set("updated_at", time.Now()).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build delete query: %w", err)
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute delete query: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return constants.ErrNotFound
	}

	return nil
}

func (r *VideoRepository) IncrementViews(ctx context.Context, id uuid.UUID) error {
	sb := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	query, args, err := sb.
		Update(constants.VideosTable).
		Set("views", "views + 1").
		Set("updated_at", time.Now()).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to increment views: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return constants.ErrNotFound
	}

	return nil
}

func (r *VideoRepository) GetUserReaction(ctx context.Context, videoID, userID uuid.UUID) (*entity.VideoReaction, error) {
	sb := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	// Получаем список полей из структуры VideoReaction
	fields, err := sqlutil.GetFields(&entity.VideoReaction{})
	if err != nil {
		return nil, fmt.Errorf("failed to get fields: %w", err)
	}

	// Формируем SELECT-запрос через Squirrel
	query, args, err := sb.
		Select(fields...).
		From(constants.VideoReactionsTable).
		Where("video_id = ?", videoID).
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	reaction := &entity.VideoReaction{}
	err = r.db.GetContext(ctx, reaction, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, constants.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get reaction: %w", err)
	}

	return reaction, nil
}

func (r *VideoRepository) CreateReaction(ctx context.Context, reaction *entity.VideoReaction) error {
	reaction.ID = uuid.New()

	if reaction.CreatedAt.IsZero() {
		reaction.CreatedAt = time.Now()
	}

	sb := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	// Получаем список полей из структуры
	fields, err := sqlutil.GetFields(reaction)
	if err != nil {
		return fmt.Errorf("failed to get fields: %w", err)
	}

	// Формируем INSERT-запрос
	insertQuery, insertArgs, err := sb.
		Insert(constants.VideoReactionsTable).
		Columns(fields...).
		Values(reaction).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build insert query: %w", err)
	}

	// Формируем часть ON CONFLICT
	onConflictQuery := `
        ON CONFLICT (video_id, user_id) DO UPDATE SET
            type = EXCLUDED.type,
            created_at = EXCLUDED.created_at,
            deleted_at = NULL
    `

	fullQuery := insertQuery + onConflictQuery

	// Выполняем запрос
	_, err = r.db.ExecContext(ctx, fullQuery, insertArgs...)
	if err != nil {
		return fmt.Errorf("failed to execute reaction query: %w", err)
	}

	return nil
}

func (r *VideoRepository) UpdateReaction(ctx context.Context, reaction *entity.VideoReaction) error {
	if reaction.CreatedAt.IsZero() {
		reaction.CreatedAt = time.Now()
	}

	sb := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	// Получаем список полей из структуры
	fields, err := sqlutil.GetFields(reaction)
	if err != nil {
		return fmt.Errorf("failed to get fields: %w", err)
	}

	// Исключаем поле id из обновления

	// Формируем SET-часть через рефлексию
	values := make(map[string]interface{})
	for _, field := range fields {
		val := reflect.Indirect(reflect.ValueOf(reaction)).FieldByName(strings.Title(field)).Interface()
		values[field] = val
	}

	delete(values, "id")

	// Билдим запрос
	query, args, err := sb.
		Update(constants.VideoReactionsTable).
		SetMap(values).
		Where("video_id = ?", reaction.VideoID).
		Where("user_id = ?", reaction.UserID).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update reaction: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return constants.ErrNotFound
	}

	return nil
}

func (r *VideoRepository) DeleteReaction(ctx context.Context, videoID, userID uuid.UUID) error {
	sb := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	query, args, err := sb.
		Update(constants.VideoReactionsTable).
		Set("deleted_at", time.Now()).
		Where("video_id = ?", videoID).
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build delete reaction query: %w", err)
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute delete reaction: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return constants.ErrNotFound
	}

	return nil
}

func (r *VideoRepository) GetVideoFiles(ctx context.Context, videoID uuid.UUID) ([]*entity.VideoFile, error) {
	query := `
        SELECT 
            vf.video_id,
            vf.quality_id,
            vf.file_format,
            vf.file_size,
            vf.width,
            vf.height,
            vf.bitrate,
            vf.status,
            vq.name as quality_name
        FROM video_files vf
        JOIN video_qualities vq ON vq.id = vf.quality_id
        WHERE vf.video_id = $1 
        AND vf.status = 'completed'
        ORDER BY vq.height ASC
    `

	var files []*entity.VideoFile
	err := r.db.SelectContext(ctx, &files, query, videoID)
	if err != nil {
		return nil, fmt.Errorf("failed to get video files: %w", err)
	}

	return files, nil
}

func (r *VideoRepository) GetVideosForConversion(ctx context.Context, limit int) ([]*entity.Video, error) {
	query := `
        SELECT * FROM videos 
        WHERE status = $1
        AND deleted_at IS NULL
        LIMIT $2
    `

	var videos []*entity.Video
	err := r.db.SelectContext(ctx, &videos, query, string(constants.VideoStatusUploaded), limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get videos: %w", err)
	}

	return videos, nil
}

func (r *VideoRepository) GetVideoQualities(ctx context.Context) ([]*entity.VideoQuality, error) {
	query := `
        SELECT * FROM video_qualities
        WHERE is_active = true
        ORDER BY height ASC
    `

	var qualities []*entity.VideoQuality
	err := r.db.SelectContext(ctx, &qualities, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get qualities: %w", err)
	}

	return qualities, nil
}

func (r *VideoRepository) CreateVideoFile(ctx context.Context, file *entity.VideoFile) error {
	query := `
        INSERT INTO video_files (
            video_id, quality_id, file_format, file_size, 
            width, height, bitrate, status, 
            created_at, updated_at
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
        )
    `

	_, err := r.db.ExecContext(ctx, query,
		file.VideoID, file.QualityID, file.Format, file.FileSize,
		file.Width, file.Height, file.Bitrate, file.Status,
		file.CreatedAt, file.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create video file: %w", err)
	}

	return nil
}

func (r *VideoRepository) UpdateStatus(ctx context.Context, videoID uuid.UUID, status string) error {
	query := `
        UPDATE videos 
        SET status = $1, updated_at = NOW()
        WHERE id = $2
    `

	result, err := r.db.ExecContext(ctx, query, status, videoID)
	if err != nil {
		return fmt.Errorf("failed to update video status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return constants.ErrNotFound
	}

	return nil
}

func (r *VideoRepository) UpdateThumbnailAndDuration(ctx context.Context, videoID uuid.UUID, thumbnailURL string, duration int) error {
	query := `
        UPDATE videos 
        SET thumbnail_url = $1,
            duration = $2,
            updated_at = NOW()
        WHERE id = $3 
        AND deleted_at IS NULL
    `

	result, err := r.db.ExecContext(ctx, query, thumbnailURL, duration, videoID)
	if err != nil {
		return fmt.Errorf("failed to update video thumbnail and duration: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return constants.ErrNotFound
	}

	return nil
}
