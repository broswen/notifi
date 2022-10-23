package repository

import (
	"context"
	"github.com/broswen/notifi/internal/db"
	"github.com/broswen/notifi/internal/entity"
	"github.com/jackc/pgx/v4/pgxpool"
)

type NotificationRepository interface {
	Get(ctx context.Context, id string) (entity.Notification, error)
	Save(ctx context.Context, n entity.Notification) (entity.Notification, error)
	Update(ctx context.Context, n entity.Notification) (entity.Notification, error)
	Delete(ctx context.Context, id string) (entity.Notification, error)
	List(ctx context.Context, offset, limit int64) ([]entity.Notification, error)
}

type NotificationSqlRepository struct {
	pool *pgxpool.Pool
}

func NewNotificationSqlRepository(pool *pgxpool.Pool) (NotificationSqlRepository, error) {
	return NotificationSqlRepository{
		pool: pool,
	}, nil
}

func (r NotificationSqlRepository) Get(ctx context.Context, id string) (entity.Notification, error) {
	n := entity.Notification{
		Destination: entity.Destination{},
	}
	err := db.PgError(r.pool.QueryRow(ctx, `select id, email_destination, sms_destination, content, schedule, deleted_at, created_at, modified_at, delivered_at from notification where id = $1;`,
		id).Scan(&n.ID, &n.Destination.Email, &n.Destination.SMS, &n.Content, &n.Schedule, &n.DeletedAt, &n.CreatedAt, &n.ModifiedAt, &n.DeliveredAt))

	return n, err
}

func (r NotificationSqlRepository) List(ctx context.Context, offset, limit int64) ([]entity.Notification, error) {
	n := entity.Notification{
		Destination: entity.Destination{},
	}
	rows, err := r.pool.Query(ctx, `select id, email_destination, sms_destination, content, schedule, deleted_at, created_at, modified_at, delivered_at from notification offset $1 limit $2;`, offset, limit)
	err = db.PgError(err)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	notifications := make([]entity.Notification)
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

func (r NotificationSqlRepository) Save(ctx context.Context, n entity.Notification) (entity.Notification, error) {
	savedNotification := entity.Notification{
		Destination: entity.Destination{},
	}
	err := db.PgError(r.pool.QueryRow(ctx, `insert into notification (id, email_destination, sms_destination, content, schedule, deleted_at, created_at, modified_at) values ($1, $2, $3, $4, $5, $6, $7, $8) returning id, email_destination, sms_destination, content, schedule, deleted_at, created_at, modified_at, delivered_at;`,
		n.ID, n.Destination.Email, n.Destination.SMS, n.Content, n.Schedule, n.DeletedAt, n.CreatedAt, n.ModifiedAt).Scan(&n.ID, &n.Destination.Email, &n.Destination.SMS, &n.Content, &n.Schedule, &n.DeletedAt, &n.CreatedAt, &n.ModifiedAt, &n.DeliveredAt))

	return savedNotification, err
}

func (r NotificationSqlRepository) Update(ctx context.Context, n entity.Notification) (entity.Notification, error) {
	updatedNotification := entity.Notification{
		Destination: entity.Destination{},
	}
	err := db.PgError(r.pool.QueryRow(ctx, `update notification set email_destination = $2, sms_destination = $3, content = $4, schedule = $5, deleted_at = $6, delivered_at = $7 where id = $1 returning id, email_destination, sms_destination, content, schedule, deleted_at, created_at, modified_at, delivered_at;`,
		n.ID, n.Destination.Email, n.Destination.SMS, n.Content, n.Schedule, n.DeletedAt, n.DeliveredAt).Scan(&n.ID, &n.Destination.Email, &n.Destination.SMS, &n.Content, &n.Schedule, &n.DeletedAt, &n.CreatedAt, &n.ModifiedAt, &n.DeliveredAt))

	return updatedNotification, err
}

func (r NotificationSqlRepository) Delete(ctx context.Context, id string) (entity.Notification, error) {
	n := entity.Notification{
		Destination: entity.Destination{},
	}
	err := db.PgError(r.pool.QueryRow(ctx, `update notification set deleted_at = now() where id = $1;`,
		id).Scan(&n.ID, &n.Destination.Email, &n.Destination.SMS, &n.Content, &n.Schedule, &n.DeletedAt, &n.CreatedAt, &n.ModifiedAt, &n.DeliveredAt))

	return n, err
}
