package testutil

import (
	"context"

	"github.com/S1FFFkA/user-mgz/internal/service"
)

// PassthroughTxManager вызывает fn с тем же ctx (для юнит-тестов без реальной БД-транзакции).
type PassthroughTxManager struct{}

func (PassthroughTxManager) Do(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

var _ service.TxManagerInterface = (PassthroughTxManager{})
