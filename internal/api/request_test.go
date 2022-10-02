package api

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestValidateRequest(t *testing.T) {
	now := time.Now()
	tests := []struct {
		request NotificationRequest
		err     error
	}{
		{
			request: NotificationRequest{
				Content:  "test message",
				Schedule: nil,
				Email:    "",
				SMS:      "",
			},
			err: ErrBadRequest.WithError(errors.New("a single destination must be specified")),
		},
		{
			request: NotificationRequest{
				Content:  "test message",
				Schedule: nil,
				Email:    "test@example.com",
				SMS:      "+18005551234",
			},
			err: ErrBadRequest.WithError(errors.New("a single destination must be specified")),
		},
		{
			request: NotificationRequest{
				Content:  "test message",
				Schedule: nil,
				Email:    "test@example.com",
				SMS:      "",
			},
			err: nil,
		},

		{
			request: NotificationRequest{
				Content:  "test message",
				Schedule: nil,
				Email:    "@example.com",
				SMS:      "",
			},
			err: ErrBadRequest.WithError(errors.New("invalid email destination")),
		},

		{
			request: NotificationRequest{
				Content:  "test message",
				Schedule: &now,
				Email:    "test@example.com",
				SMS:      "",
			},
			err: ErrBadRequest.WithError(errors.New("scheduled time must not be in the past")),
		},
	}

	for _, test := range tests {
		err := test.request.Validate()
		if test.err == nil {
			assert.NoError(t, err)
		} else {
			assert.Equal(t, test.err, err)
		}
	}
}
