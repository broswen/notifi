package producer

import (
	"github.com/broswen/notifi/internal/entity"
	"github.com/stretchr/testify/mock"
)

type MockProducer struct {
	mock.Mock
}

func (m MockProducer) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m MockProducer) Submit(notification entity.Notification) error {
	args := m.Called(notification)
	return args.Error(0)
}
