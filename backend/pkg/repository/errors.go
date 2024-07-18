package repository

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	errDatabase = errors.New("database error")
)

var (
	errScan                 = fmt.Errorf("%w: scan error", errDatabase)
	errBadQuery             = fmt.Errorf("%w: bad query error", errDatabase)
	errBadLastInsertID      = fmt.Errorf("%w: bad lastID", errDatabase)
	errUniqueIndexViolation = fmt.Errorf("%w: duplicated entry", errDatabase)
	errRowsNotFound         = fmt.Errorf("%w: rows not found", errDatabase)
)

type ErrQuery struct {
	cause      error
	queryCause error
}

// NewQueryError creates a new ErrQuery, used only for query errors
func NewQueryError(cause, queryCause error) ErrQuery {
	return ErrQuery{
		cause:      cause,
		queryCause: queryCause,
	}
}

func (e ErrQuery) StatusCode() int {
	return http.StatusInternalServerError
}

func (e ErrQuery) Unwrap() error {
	return e.cause
}

func (e ErrQuery) Error() string {
	return fmt.Sprintf("%s: %s", e.cause, e.queryCause)
}
