package service

import (
	"context"

	"github.com/avito-tech/go-transaction-manager/trm/v2"
)

// TxManagerInterface — одна транзакция БД вокруг нескольких вызовов репозитория.
type TxManagerInterface interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

// TxManager обёртка над github.com/avito-tech/go-transaction-manager.
type TxManager struct {
	manager trm.Manager
}

func NewTxManager(manager trm.Manager) *TxManager {
	return &TxManager{manager: manager}
}

func (m *TxManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	return m.manager.Do(ctx, fn)
}

var _ TxManagerInterface = (*TxManager)(nil)
