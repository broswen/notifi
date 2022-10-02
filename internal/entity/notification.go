package entity

import "time"

type Notification struct {
	ID          string      `json:"id"`
	Destination Destination `json:"destination"`
	Content     string      `json:"content"`
	Schedule    *time.Time  `json:"schedule"`
}

type Destination struct {
	Email string `json:"email"`
	SMS   string `json:"sms"`
}

func (n *Notification) Scheduled() bool {
	return n.Schedule != nil
}
