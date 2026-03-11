package repository

import (
	repocore "github.com/S1FFFkA/user-mgz/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

var _ repocore.UserPhotoRepository = (*Repository)(nil)
