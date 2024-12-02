package repository

import (
	"context"
	"database/sql"

	"github.com/go-kit/log"
)

const (
	pgErrorCodeUniqueViolation = "23505"
	pgErrorCodeNoData          = "02000"
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
	logger       log.Logger
}

func NewRepository(connection *sql.DB, directory string, logger log.Logger) *PostgresRepository {
	return &PostgresRepository{
		db:           connection,
		documentsDir: directory,
		logger:       logger,
	}
}
