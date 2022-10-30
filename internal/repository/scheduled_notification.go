package repository

import (
	"context"
	"github.com/broswen/notifi/internal/db"
	"github.com/broswen/notifi/internal/entity"
	"github.com/jackc/pgx/v4/pgxpool"
	"time"
)

type ScheduledNotificationRepository interface {
	MarkSubmitted(ctx context.Context, id string) (entity.Notification, error)
	ListScheduled(ctx context.Context, period time.Duration, offset, limit int64) ([]entity.Notification, error)
}

type ScheduledNotificationSqlRepository struct {
	pool *pgxpool.Pool
}

func NewScheduledNotificationSqlRepository(pool *pgxpool.Pool) (ScheduledNotificationRepository, error) {
	return &ScheduledNotificationSqlRepository{
		pool: pool,
	}, nil
}

func (r *ScheduledNotificationSqlRepository) ListScheduled(ctx context.Context, period time.Duration, offset, limit int64) ([]entity.Notification, error) {
	//p is the maximum time we are willing to send notifications early
	p := time.Now().Add(period)
	rows, err := r.pool.Query(ctx, `select id, email_destination, sms_destination, content, schedule, deleted_at, created_at, modified_at, delivered_at from notification 
		where delivered_at is null
		and deleted_at is null 
-- 		only submit if it hasn't been submitted or the previous submission was over 5 minutes ago
		and (submitted_at is null or submitted_at < (now() - interval '5 min'))
		and schedule is not null
		and schedule < $1
		order by schedule asc
		offset $2 limit $3;`, p, offset, limit)
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

func (r *ScheduledNotificationSqlRepository) MarkSubmitted(ctx context.Context, id string) (entity.Notification, error) {
	un := entity.Notification{
		Destination: entity.Destination{},
	}
	err := db.PgError(r.pool.QueryRow(ctx, `update notification set submitted_at = now() where id = $1 returning id, email_destination, sms_destination, content, schedule, deleted_at, created_at, modified_at, delivered_at, submitted_at;`,
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
