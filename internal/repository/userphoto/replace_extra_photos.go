package repository

import (
	"context"
	"fmt"

	"github.com/S1FFFkA/user-mgz/internal/domain"
	repocore "github.com/S1FFFkA/user-mgz/internal/repository"
	"github.com/google/uuid"
)

func (r *Repository) ReplaceExtraPhotos(ctx context.Context, userID uuid.UUID, photos []domain.UserPhoto) error {
	if len(photos) == 0 {
		return fmt.Errorf("replace extra photos: photos list must not be empty")
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx replace extra photos: %w", err)
	}
	defer tx.Rollback(ctx)

	if err = repocore.ReplaceExtraPhotosTx(ctx, tx, userID, photos); err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit replace extra photos: %w", err)
	}
	return nil
}
