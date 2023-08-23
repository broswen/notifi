package repository

import (
	"context"
	"time"

	"github.com/broswen/notifi/internal/db"
	"github.com/broswen/notifi/internal/entity"
)

type ScheduledNotificationRepository interface {
	MarkSubmitted(ctx context.Context, id string) (entity.Notification, error)
	ListScheduled(ctx context.Context, period time.Duration, limit int64) ([]entity.Notification, error)
}

type ScheduledNotificationSqlRepository struct {
	db db.Conn
}

func NewScheduledNotificationSqlRepository(db db.Conn) ScheduledNotificationRepository {
	return &ScheduledNotificationSqlRepository{
		db: db,
	}
}

// ListScheduled selects and locks a set of notifications
func (r *ScheduledNotificationSqlRepository) ListScheduled(ctx context.Context, period time.Duration, limit int64) ([]entity.Notification, error) {
	//p is the maximum time we are willing to send notifications early
	p := time.Now().Add(period)
	rows, err := r.db.QueryContext(ctx, `select id, email_destination, sms_destination, content, schedule, deleted_at, created_at, modified_at, delivered_at from notification 
		where delivered_at is null
		and deleted_at is null 
		and submitted_at is null
		and schedule is not null
		and schedule < $1
		order by schedule asc
		limit $2 for update skip locked;`, p, limit)
	err = db.PgError(err)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	notifications := make([]entity.Notification, 0)
	for rows.Next() {
		n := entity.Notification{}
		err = rows.Scan(&n.ID, &n.Destination.Email, &n.Destination.SMS, &n.Content, &n.Schedule, &n.DeletedAt, &n.CreatedAt, &n.ModifiedAt, &n.DeliveredAt)
		if err != nil {
			return notifications, err
		}
		notifications = append(notifications, n)
	}
	if rows.Err() != nil {
		return notifications, err
	}
	return notifications, err
}

// MarkSubmitted marks a noitifcation as submitted
func (r *ScheduledNotificationSqlRepository) MarkSubmitted(ctx context.Context, id string) (entity.Notification, error) {
	un := entity.Notification{
		Destination: entity.Destination{},
	}
	err := db.PgError(r.db.QueryRowContext(ctx, `update notification set submitted_at = now() where id = $1 returning id, email_destination, sms_destination, content, schedule, deleted_at, created_at, modified_at, delivered_at, submitted_at;`,
		id).
		Scan(&un.ID, &un.Destination.Email, &un.Destination.SMS, &un.Content, &un.Schedule, &un.DeletedAt, &un.CreatedAt, &un.ModifiedAt, &un.DeliveredAt, &un.SubmittedAt))

	if err != nil {
		switch err {
		case db.ErrNotFound:
			return un, ErrNotificationNotFound{err.Error()}
		case db.ErrInvalidData:
			return un, ErrInvalidData{err.Error()}
		default:
			return un, ErrUnknown{err}
		}
	}

	return un, nil
}
