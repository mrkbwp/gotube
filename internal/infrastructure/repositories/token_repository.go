package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/mrkbwp/gotube/pkg/constants"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/mrkbwp/gotube/internal/domain/entity"
	"github.com/mrkbwp/gotube/internal/domain/repositories"
)

// TokenRepository реализует интерфейс TokenRepository
type TokenRepository struct {
	db *sqlx.DB
}

// NewTokenRepository создает новый экземпляр TokenRepository
func NewTokenRepository(db *sqlx.DB) repositories.TokenRepository {
	return &TokenRepository{
		db: db,
	}
}

// Create создает новый refresh токен
func (r *TokenRepository) Create(ctx context.Context, token *entity.Token) error {
	query := `
        INSERT INTO tokens (id, user_id, refresh_token, user_agent, client_ip, is_blocked, expires_at, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `

	token.ID = uuid.New()

	_, err := r.db.ExecContext(
		ctx,
		query,
		token.ID,
		token.UserID,
		token.RefreshToken,
		token.UserAgent,
		token.ClientIP,
		token.IsBlocked,
		token.ExpiresAt,
		token.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create token: %w", err)
	}

	return nil
}

// GetByRefreshToken находит токен по значению refresh токена
func (r *TokenRepository) GetByRefreshToken(ctx context.Context, refreshToken string) (*entity.Token, error) {
	query := `
        SELECT id, user_id, refresh_token, user_agent, client_ip, is_blocked, expires_at, created_at
        FROM tokens
        WHERE refresh_token = $1
    `

	var token entity.Token
	err := r.db.GetContext(ctx, &token, query, refreshToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, constants.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	return &token, nil
}

// DeleteByUserID удаляет все токены пользователя
func (r *TokenRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `
        DELETE FROM tokens
        WHERE user_id = $1
    `

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete tokens: %w", err)
	}

	return nil
}

// DeleteByID удаляет токен по ID
func (r *TokenRepository) DeleteByID(ctx context.Context, tokenID uuid.UUID) error {
	query := `
        DELETE FROM tokens
        WHERE id = $1
    `

	_, err := r.db.ExecContext(ctx, query, tokenID)
	if err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}

	return nil
}

// UpdateBlockStatus обновляет статус блокировки токена
func (r *TokenRepository) UpdateBlockStatus(ctx context.Context, tokenID uuid.UUID, isBlocked bool) error {
	query := `
        UPDATE tokens
        SET is_blocked = $2
        WHERE id = $1
    `

	_, err := r.db.ExecContext(ctx, query, tokenID, isBlocked)
	if err != nil {
		return fmt.Errorf("failed to update token status: %w", err)
	}

	return nil
}
