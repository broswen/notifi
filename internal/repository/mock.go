package repository

import (
	"context"
	"github.com/broswen/notifi/internal/entity"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m MockRepository) Get(ctx context.Context, id string) (entity.Notification, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(entity.Notification), args.Error(1)
}

func (m MockRepository) List(ctx context.Context, offset, limit int64) ([]entity.Notification, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]entity.Notification), args.Error(1)
}

func (m MockRepository) Save(ctx context.Context, n entity.Notification) (entity.Notification, error) {
	args := m.Called(ctx, n)
	return args.Get(0).(entity.Notification), args.Error(1)
}

func (m MockRepository) Update(ctx context.Context, n entity.Notification) (entity.Notification, error) {
	args := m.Called(ctx, n)
	return args.Get(0).(entity.Notification), args.Error(1)
}

func (m MockRepository) Delete(ctx context.Context, id string) (entity.Notification, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(entity.Notification), args.Error(1)
}