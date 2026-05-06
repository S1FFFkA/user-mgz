package service

import (
	"context"
	"testing"

	trm "github.com/avito-tech/go-transaction-manager/trm/v2"
)

type fakeTRM struct{}

func (fakeTRM) Do(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

func (fakeTRM) DoWithSettings(ctx context.Context, _ trm.Settings, fn func(context.Context) error) error {
	return fn(ctx)
}

func TestTxManager_Do(t *testing.T) {
	var ran bool
	m := NewTxManager(fakeTRM{})
	err := m.Do(context.Background(), func(ctx context.Context) error {
		ran = true
		return nil
	})
	if err != nil || !ran {
		t.Fatalf("err=%v ran=%v", err, ran)
	}
}
