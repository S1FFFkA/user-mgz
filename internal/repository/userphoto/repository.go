package userphoto

import (
	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/S1FFFkA/user-mgz/internal/repository"
)

type UserPhotoRepository struct {
	pool   *pgxpool.Pool
	getter *trmpgx.CtxGetter
}

func New(pool *pgxpool.Pool) *UserPhotoRepository {
	return &UserPhotoRepository{
		pool:   pool,
		getter: trmpgx.DefaultCtxGetter,
	}
}

var _ repository.UserPhotoRepositoryInterface = (*UserPhotoRepository)(nil)
