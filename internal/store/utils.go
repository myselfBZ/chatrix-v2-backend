package store

import (
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrNotFound          = errors.New("resource not found")
	ErrAlreadyExists     = errors.New("resource already exists")
	ErrConstraintMessage = errors.New("operation violated a database constraint")
	ErrInternal          = errors.New("an internal storage error occurred")
)


// mapError converts database-specific errors into domain-specific errors
func mapError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			return ErrAlreadyExists
		case "23503": // foreign_key_violation
			return ErrConstraintMessage
		}
	}

	slog.Info(err.Error())

	return ErrInternal
}
