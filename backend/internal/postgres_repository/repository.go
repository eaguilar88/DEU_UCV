package repository

import (
	"context"
	"database/sql"
	"errors"

	"go.uber.org/zap"
)

const (
	pgErrorCodeUniqueViolation = "23505"
)

var (
	ErrDatabaseError = errors.New("database error")
	ErrScanError     = errors.New("scan error")
	ErrInvalidQuery  = errors.New("invalid query")
)

type scannable interface {
	Scan(dest ...interface{}) error
}

type preparer interface {
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

type PostgresRepository struct {
	db           *sql.DB
	documentsDir string
	logger       *zap.Logger
}

func NewRepository(connection *sql.DB, directory string, logger *zap.Logger) *PostgresRepository {
	return &PostgresRepository{
		db:           connection,
		documentsDir: directory,
		logger:       logger,
	}
}

// toNullString converts a string to sql.NullString.
// Returns a valid NullString if the value is non-empty.
func toNullString(s string) sql.NullString {
	return sql.NullString{
		String: s,
		Valid:  s != "",
	}
}

// toNullBool converts a bool to sql.NullBool with Valid always true.
func toNullBool(b bool) sql.NullBool {
	return sql.NullBool{
		Bool:  b,
		Valid: true,
	}
}
