package main

import (
	"github.com/broswen/notifi/internal/destination"
	"github.com/broswen/notifi/internal/entity"
	"github.com/broswen/notifi/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func TestHandleDelivery_Email(t *testing.T) {
	nr := &repository.MockRepository{}
	sms := &destination.MockDestination{}
	email := &destination.MockDestination{}
	l := &destination.MockDestination{}
	skipDelivery := ""
	handler := HandleDelivery(nr, sms, email, l, nil, time.Hour, skipDelivery)
	n := entity.Notification{
		ID: "abc123",
		Destination: entity.Destination{
			Email: "test@example.com",
		},
		Content:   "test message",
		CreatedAt: time.Time{},
	}
	nr.On("Ping", mock.Anything).Return(nil)
	email.On("Deliver", n).Return(nil)
	now := time.Now()
	n2 := entity.Notification{
		ID: "abc123",
		Destination: entity.Destination{
			Email: "test@example.com",
		},
		Content:     "test message",
		Schedule:    nil,
		DeletedAt:   nil,
		CreatedAt:   time.Time{},
		ModifiedAt:  time.Time{},
		DeliveredAt: &now,
	}
	nr.On("Update", mock.Anything, mock.Anything).Return(n2, nil)
	err := handler(n)
	assert.NoError(t, err)
	nr.AssertExpectations(t)
	sms.AssertExpectations(t)
	email.AssertExpectations(t)
	l.AssertExpectations(t)
}

func TestHandleDelivery_Sms(t *testing.T) {
	nr := &repository.MockRepository{}
	sms := &destination.MockDestination{}
	email := &destination.MockDestination{}
	l := &destination.MockDestination{}
	skipDelivery := ""
	handler := HandleDelivery(nr, sms, email, l, nil, time.Hour, skipDelivery)
	n := entity.Notification{
		ID: "abc123",
		Destination: entity.Destination{
			SMS: "1234567890",
		},
		Content:   "test message",
		CreatedAt: time.Time{},
	}
	nr.On("Ping", mock.Anything).Return(nil)
	sms.On("Deliver", n).Return(nil)
	now := time.Now()
	n2 := entity.Notification{
		ID: "abc123",
		Destination: entity.Destination{
			SMS: "1234567890",
		},
		Content:     "test message",
		Schedule:    nil,
		DeletedAt:   nil,
		CreatedAt:   time.Time{},
		ModifiedAt:  time.Time{},
		DeliveredAt: &now,
	}
	nr.On("Update", mock.Anything, mock.Anything).Return(n2, nil)
	err := handler(n)
	assert.NoError(t, err)
	nr.AssertExpectations(t)
	sms.AssertExpectations(t)
	email.AssertExpectations(t)
	l.AssertExpectations(t)
}
