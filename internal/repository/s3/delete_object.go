package s3

import (
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
)

func (r *Repository) DeleteObject(ctx context.Context, objectKey string) error {
	err := r.client.RemoveObject(ctx, r.bucket, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("delete object: %w", err)
	}
	return nil
}
