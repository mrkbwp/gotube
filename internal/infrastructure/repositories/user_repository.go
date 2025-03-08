package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/mrkbwp/gotube/pkg/constants"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/mrkbwp/gotube/internal/domain/entity"
	"github.com/mrkbwp/gotube/internal/domain/repositories"
)

// UserRepository реализует интерфейс UserRepository
type UserRepository struct {
	db *sqlx.DB
}

// NewUserRepository создает новый экземпляр UserRepository
func NewUserRepository(db *sqlx.DB) repositories.UserRepository {
	return &UserRepository{
		db: db,
	}
}

// Create создает нового пользователя
func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO users (id, username, email, password_hash, avatar, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	user.ID = uuid.New()

	var id string
	err := r.db.QueryRowContext(
		ctx,
		query,
		user.ID,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.Avatar,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&id)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID возвращает пользователя по ID
func (r *UserRepository) GetByID(ctx context.Context, id string) (*entity.User, error) {
	query := `
		SELECT id, username, email, password_hash, avatar, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user entity.User
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, constants.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetByEmail возвращает пользователя по email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `
		SELECT id, username, email, password_hash, avatar, role, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user entity.User
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, constants.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

// Update обновляет данные пользователя
func (r *UserRepository) Update(ctx context.Context, user *entity.User) error {
	query := `
		UPDATE users
		SET username = $2, email = $3, avatar = $4, role = $5, updated_at = $6
		WHERE id = $1
	`

	user.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(
		ctx,
		query,
		user.ID,
		user.Username,
		user.Email,
		user.Avatar,
		user.Role,
		user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// UpdatePassword обновляет пароль пользователя
func (r *UserRepository) UpdatePassword(ctx context.Context, id, passwordHash string) error {
	query := `
		UPDATE users
		SET password_hash = $2, updated_at = $3
		WHERE id = $1
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		id,
		passwordHash,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// Delete удаляет пользователя
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	query := `
		DELETE FROM users
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}
