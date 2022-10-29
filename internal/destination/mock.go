package destination

import (
	"github.com/broswen/notifi/internal/entity"
	"github.com/stretchr/testify/mock"
)

type MockDestination struct {
	mock.Mock
}

func (m *MockDestination) Deliver(notification entity.Notification) error {
	args := m.Called(notification)
	return args.Error(0)
}
