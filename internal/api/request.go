package api

import (
	"errors"
	"github.com/broswen/notifi/internal/entity"
	"github.com/segmentio/ksuid"
	"regexp"
	"time"
)

var emailRegexp = regexp.MustCompile("^[-0-9a-zA-Z]+@[-0-9a-zA-Z]+\\.[a-zA-Z]{2,3}$")
var phoneRegexp = regexp.MustCompile("(?:\\d{1}\\s)?\\(?(\\d{3})\\)?-?\\s?(\\d{3})-?\\s?(\\d{4})")

type NotificationRequest struct {
	Content  string     `json:"content"`
	Schedule *time.Time `json:"schedule"`
	Email    string     `json:"email"`
	SMS      string     `json:"sms"`
}

func (nr *NotificationRequest) Validate() error {
	//TODO allow schedules slightly in the past with a buffer time
	if nr.Schedule != nil && nr.Schedule.Before(time.Now()) {
		return ErrBadRequest.WithError(errors.New("scheduled time must not be in the past"))
	}

	if nr.Email == "" && nr.SMS == "" {
		return ErrBadRequest.WithError(errors.New("a single destination must be specified"))
	}
	if nr.Email != "" && nr.SMS != "" {
		return ErrBadRequest.WithError(errors.New("a single destination must be specified"))
	}

	if nr.Email != "" && !emailRegexp.MatchString(nr.Email) {
		return ErrBadRequest.WithError(errors.New("invalid email destination"))
	}

	if nr.SMS != "" && !phoneRegexp.MatchString(nr.SMS) {
		return ErrBadRequest.WithError(errors.New("invalid sms destination"))
	}

	return nil
}

func (nr NotificationRequest) IntoEntity() entity.Notification {
	k := ksuid.New()
	//TODO calculate partition key here?
	n := entity.Notification{
		ID: k.String(),
		Destination: entity.Destination{
			Email: nr.Email,
			SMS:   nr.SMS,
		},
		Content:    nr.Content,
		Schedule:   nr.Schedule,
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	}
	return n
}
