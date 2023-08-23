package db

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog/log"
)

type Conn interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

func InitDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

var (
	ErrNotFound     = errors.New("not found")
	ErrInvalidData  = errors.New("invalid data")
	ErrKeyNotUnique = errors.New("key not unique")
	ErrUnknown      = errors.New("unknown error")
)

func PgError(err error) error {
	if err == nil {
		return nil
	}
	log.Error().Err(err)

	var pgError *pgconn.PgError
	if errors.As(err, &pgError) {
		log.Warn().
			Str("code", pgError.Code).
			Str("message", pgError.Message).
			Str("detail", pgError.Detail).
			Str("constraint", pgError.ConstraintName).
			Str("table", pgError.TableName).
			Str("column", pgError.ColumnName)

		//https://www.postgresql.org/docs/11/errcodes-appendix.html
		//convert postgres error codes to user friendly errors
		switch {
		case strings.HasPrefix(pgError.Code, "22"):
			return ErrInvalidData
		case strings.HasPrefix(pgError.Code, "23"):
			return ErrKeyNotUnique
		default:
			return ErrUnknown
		}
	}

	//pgx returns ErrNoRows
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}

	return err
}
