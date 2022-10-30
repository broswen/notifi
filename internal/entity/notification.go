package entity

import "time"

type Notification struct {
	ID          string      `json:"id"`
	Destination Destination `json:"destination"`
	Content     string      `json:"content"`
	Schedule    *time.Time  `json:"schedule"`
	DeletedAt   *time.Time  `json:"deleted_at"`
	CreatedAt   time.Time   `json:"created_at"`
	ModifiedAt  time.Time   `json:"modified_at"`
	DeliveredAt *time.Time  `json:"delivered_at"`
	SubmittedAt *time.Time  `json:"submitted_at"`
}

type Destination struct {
	Email string `json:"email"`
	SMS   string `json:"sms"`
}

func (n *Notification) Scheduled() bool {
	return n.Schedule != nil
}
