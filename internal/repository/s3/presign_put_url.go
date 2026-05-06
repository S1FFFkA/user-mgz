package s3

import (
	"context"
	"fmt"
	"time"
)

func (r *Repository) PresignPutURL(ctx context.Context, objectKey, contentType string, contentLength int64, expiresIn time.Duration) (string, error) {
	_ = contentType
	_ = contentLength

	u, err := r.client.PresignedPutObject(ctx, r.bucket, objectKey, expiresIn)
	if err != nil {
		return "", fmt.Errorf("presign put url: %w", err)
	}
	return u.String(), nil
}
