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
	Ping(ctx context.Context) error
}

type NotificationSqlRepository struct {
	pool *pgxpool.Pool
}

func NewNotificationSqlRepository(pool *pgxpool.Pool) (*NotificationSqlRepository, error) {
	return &NotificationSqlRepository{
		pool: pool,
	}, nil
}

func (r *NotificationSqlRepository) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}

func (r *NotificationSqlRepository) Get(ctx context.Context, id string) (entity.Notification, error) {
	n := entity.Notification{
		Destination: entity.Destination{},
	}
	err := db.PgError(r.pool.QueryRow(ctx, `select id, email_destination, sms_destination, content, schedule, deleted_at, created_at, modified_at, delivered_at, submitted_at from notification where id = $1 and deleted_at is null;`,
		id).Scan(&n.ID, &n.Destination.Email, &n.Destination.SMS, &n.Content, &n.Schedule, &n.DeletedAt, &n.CreatedAt, &n.ModifiedAt, &n.DeliveredAt, &n.SubmittedAt))

	if err != nil {
		switch err {
		case db.ErrNotFound:
			return n, ErrNotificationNotFound{err.Error()}
		case db.ErrInvalidData:
			return n, ErrInvalidData{err.Error()}
		default:
			return n, ErrUnknown{err}
		}
	}

	return n, nil
}

func (r *NotificationSqlRepository) List(ctx context.Context, offset, limit int64) ([]entity.Notification, error) {
	rows, err := r.pool.Query(ctx, `select id, email_destination, sms_destination, content, schedule, deleted_at, created_at, modified_at, delivered_at from notification where deleted_at is null offset $1 limit $2;`, offset, limit)
	err = db.PgError(err)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	notifications := make([]entity.Notification, 0)
	for rows.Next() {
		n := entity.Notification{}
		err = rows.Scan(&n.ID, &n.Destination.Email, &n.Destination.SMS, &n.Content, &n.Schedule, &n.DeletedAt, &n.CreatedAt, &n.ModifiedAt, &n.DeliveredAt, &n.SubmittedAt)
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

func (r *NotificationSqlRepository) Save(ctx context.Context, n entity.Notification) (entity.Notification, error) {
	savedNotification := entity.Notification{
		Destination: entity.Destination{},
	}
	err := db.PgError(r.pool.QueryRow(ctx, `insert into notification (id, email_destination, sms_destination, content, schedule, deleted_at, created_at, modified_at, delivered_at, submitted_at) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) returning id, email_destination, sms_destination, content, schedule, deleted_at, created_at, modified_at, delivered_at, submitted_at;`,
		n.ID, n.Destination.Email, n.Destination.SMS, n.Content, n.Schedule, n.DeletedAt, n.CreatedAt, n.ModifiedAt, n.DeliveredAt, n.SubmittedAt).Scan(&n.ID, &n.Destination.Email, &n.Destination.SMS, &n.Content, &n.Schedule, &n.DeletedAt, &n.CreatedAt, &n.ModifiedAt, &n.DeliveredAt, &n.SubmittedAt))

	if err != nil {
		switch err {
		case db.ErrNotFound:
			return n, ErrNotificationNotFound{err.Error()}
		case db.ErrInvalidData:
			return n, ErrInvalidData{err.Error()}
		default:
			return n, ErrUnknown{err}
		}
	}

	return savedNotification, nil
}

func (r *NotificationSqlRepository) Update(ctx context.Context, n entity.Notification) (entity.Notification, error) {
	updatedNotification := entity.Notification{
		Destination: entity.Destination{},
	}
	err := db.PgError(r.pool.QueryRow(ctx, `update notification set email_destination = $2, sms_destination = $3, content = $4, schedule = $5, deleted_at = $6, delivered_at = $7 where id = $1 returning id, email_destination, sms_destination, content, schedule, deleted_at, created_at, modified_at, delivered_at, submitted_at;`,
		n.ID, n.Destination.Email, n.Destination.SMS, n.Content, n.Schedule, n.DeletedAt, n.DeliveredAt).Scan(&n.ID, &n.Destination.Email, &n.Destination.SMS, &n.Content, &n.Schedule, &n.DeletedAt, &n.CreatedAt, &n.ModifiedAt, &n.DeliveredAt, &n.SubmittedAt))

	if err != nil {
		switch err {
		case db.ErrNotFound:
			return n, ErrNotificationNotFound{err.Error()}
		case db.ErrInvalidData:
			return n, ErrInvalidData{err.Error()}
		default:
			return n, ErrUnknown{err}
		}
	}

	return updatedNotification, nil
}

func (r *NotificationSqlRepository) Delete(ctx context.Context, id string) (entity.Notification, error) {
	n := entity.Notification{
		Destination: entity.Destination{},
	}
	err := db.PgError(r.pool.QueryRow(ctx, `update notification set deleted_at = now() where id = $1 returning id, email_destination, sms_destination, content, schedule, deleted_at, created_at, modified_at, delivered_at;`,
		id).Scan(&n.ID, &n.Destination.Email, &n.Destination.SMS, &n.Content, &n.Schedule, &n.DeletedAt, &n.CreatedAt, &n.ModifiedAt, &n.DeliveredAt))

	if err != nil {
		switch err {
		case db.ErrNotFound:
			return n, ErrNotificationNotFound{err.Error()}
		case db.ErrInvalidData:
			return n, ErrInvalidData{err.Error()}
		default:
			return n, ErrUnknown{err}
		}
	}

	return n, nil
}
