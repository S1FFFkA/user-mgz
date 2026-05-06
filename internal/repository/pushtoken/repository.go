package pushtoken

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PushToken struct {
	ID       int64
	UserID   uuid.UUID
	Provider string
	Token    string
	DeviceID string
	Platform string
}

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]PushToken, error) {
	const q = `
SELECT id, user_id, provider, token, device_id, platform
FROM user_push_tokens
WHERE user_id = $1`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("list push tokens by user id: %w", err)
	}
	defer rows.Close()

	out := make([]PushToken, 0)
	for rows.Next() {
		var p PushToken
		if err = rows.Scan(&p.ID, &p.UserID, &p.Provider, &p.Token, &p.DeviceID, &p.Platform); err != nil {
			return nil, fmt.Errorf("scan push token: %w", err)
		}
		out = append(out, p)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate push tokens: %w", err)
	}
	return out, nil
}

func (r *Repository) Upsert(ctx context.Context, userID uuid.UUID, provider, token, deviceID, platform string) error {
	const q = `
INSERT INTO user_push_tokens (user_id, provider, token, device_id, platform, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
ON CONFLICT (user_id, device_id) DO UPDATE SET
	provider = EXCLUDED.provider,
	token = EXCLUDED.token,
	platform = EXCLUDED.platform,
	updated_at = NOW()`
	_, err := r.pool.Exec(ctx, q, userID, provider, token, deviceID, platform)
	if err != nil {
		return fmt.Errorf("upsert push token: %w", err)
	}
	return nil
}

func (r *Repository) DeleteByUserAndDevice(ctx context.Context, userID uuid.UUID, deviceID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM user_push_tokens WHERE user_id = $1 AND device_id = $2`, userID, deviceID)
	if err != nil {
		return fmt.Errorf("delete push token: %w", err)
	}
	return nil
}
