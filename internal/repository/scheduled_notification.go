package repository

import (
	"context"
	"github.com/broswen/notifi/internal/db"
	"github.com/broswen/notifi/internal/entity"
	"github.com/jackc/pgx/v4/pgxpool"
	"time"
)

type ScheduledNotificationRepository interface {
	ListScheduled(ctx context.Context, period time.Duration, offset, limit int64) ([]entity.Notification, error)
}

type ScheduledNotificationSqlRepository struct {
	pool *pgxpool.Pool
}

func NewScheduledNotificationSqlRepository(pool *pgxpool.Pool) (ScheduledNotificationRepository, error) {
	return ScheduledNotificationSqlRepository{
		pool: pool,
	}, nil
}

func (r ScheduledNotificationSqlRepository) ListScheduled(ctx context.Context, period time.Duration, offset, limit int64) ([]entity.Notification, error) {
	//p is the maximum time we are willing to send notifications early
	p := time.Now().Add(period)
	//TODO verify proper filtering order with compound index
	rows, err := r.pool.Query(ctx, `select id, email_destination, sms_destination, content, schedule, deleted_at, created_at, modified_at, delivered_at from notification 
		where delivered_at is null
		and deleted_at is null 
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
