package s3

import (
	"github.com/S1FFFkA/user-mgz/internal/repository"
	"github.com/minio/minio-go/v7"
)

// Repository — MinIO/S3 в слое репозитория (инфраструктура).
type Repository struct {
	client *minio.Client
	bucket string
}

func New(client *minio.Client, bucket string) *Repository {
	return &Repository{
		client: client,
		bucket: bucket,
	}
}

var _ repository.S3RepositoryInterface = (*Repository)(nil)
