package repository

import (
	repocore "github.com/S1FFFkA/user-mgz/internal/repository"
	"github.com/minio/minio-go/v7"
)

type Repository struct {
	client *minio.Client
	bucket string
}

func NewRepository(client *minio.Client, bucket string) *Repository {
	return &Repository{
		client: client,
		bucket: bucket,
	}
}

var _ repocore.S3Repository = (*Repository)(nil)
