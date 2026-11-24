// Package backups makes the daily PostgreSQL backups: it dumps the database with pg_dump, uploads
// the dump to Backblaze B2 and deletes the dumps older than the retention period.
package backups

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/eaguilar88/deu/internal/entities"
	"go.uber.org/zap"
)

const (
	dumpExtension   = ".dump"
	dumpContentType = "application/octet-stream"
	timestampLayout = "20060102_150405"
)

// Dumper writes a full dump of the database to w.
type Dumper interface {
	Dump(ctx context.Context, w io.Writer) error
}

// Storage defines the file storage operations required by the backups service.
type Storage interface {
	PutObject(ctx context.Context, objectKey, contentType string, body io.Reader) error
	ListObjects(ctx context.Context, prefix string) ([]entities.StoredObject, error)
	DeleteFile(ctx context.Context, objectKey string) error
}

// Config sets where the dumps are stored and for how long.
type Config struct {
	// DatabaseName is used to name the dump files.
	DatabaseName string
	// Prefix is the storage key prefix for the dumps, e.g. "backups/postgres/".
	Prefix string
	// Retention is how long a dump is kept. Zero or negative keeps every dump.
	Retention time.Duration
}

// Result describes a finished backup run.
type Result struct {
	Key     string
	Size    int64
	Deleted []string
}

type Service struct {
	dumper  Dumper
	storage Storage
	config  Config
	now     func() time.Time
	tempDir string
	logger  *zap.Logger
}

func NewService(dumper Dumper, storage Storage, config Config, logger *zap.Logger) *Service {
	if !strings.HasSuffix(config.Prefix, "/") && config.Prefix != "" {
		config.Prefix += "/"
	}
	return &Service{
		dumper:  dumper,
		storage: storage,
		config:  config,
		now:     time.Now,
		tempDir: os.TempDir(),
		logger:  logger,
	}
}

// Run dumps the database, uploads the dump and then removes the expired dumps. The dump is first
// written to a temporary file so a failed pg_dump never leaves a partial backup in the bucket, and
// so the upload body is seekable (the S3 client signs the payload).
func (s *Service) Run(ctx context.Context) (Result, error) {
	key := s.objectKey(s.now())

	tmp, err := os.CreateTemp(s.tempDir, "deu-backup-*"+dumpExtension)
	if err != nil {
		return Result{}, fmt.Errorf("creating temporary dump file: %w", err)
	}
	defer func() {
		if err := tmp.Close(); err != nil && !errors.Is(err, os.ErrClosed) {
			s.logger.Warn("failed to close temporary dump", zap.Error(err))
		}
		if err := os.Remove(tmp.Name()); err != nil {
			s.logger.Warn("failed to remove temporary dump", zap.String("path", tmp.Name()), zap.Error(err))
		}
	}()

	if err := s.dumper.Dump(ctx, tmp); err != nil {
		return Result{}, fmt.Errorf("dumping the database: %w", err)
	}

	info, err := tmp.Stat()
	if err != nil {
		return Result{}, fmt.Errorf("reading dump size: %w", err)
	}
	if info.Size() == 0 {
		return Result{}, errors.New("dumping the database: pg_dump produced an empty file")
	}
	if _, err := tmp.Seek(0, io.SeekStart); err != nil {
		return Result{}, fmt.Errorf("rewinding dump file: %w", err)
	}

	if err := s.storage.PutObject(ctx, key, dumpContentType, tmp); err != nil {
		return Result{}, fmt.Errorf("uploading dump %s: %w", key, err)
	}
	s.logger.Info("database backup uploaded", zap.String("key", key), zap.Int64("bytes", info.Size()))

	result := Result{Key: key, Size: info.Size()}

	// A cleanup failure doesn't undo the backup that was just stored; it is reported and the
	// expired dumps are deleted on the next run.
	deleted, err := s.deleteExpired(ctx, key)
	result.Deleted = deleted
	if err != nil {
		return result, fmt.Errorf("deleting expired backups: %w", err)
	}
	return result, nil
}

// objectKey names a dump after the database and the UTC time it was taken, so keys sort by date.
func (s *Service) objectKey(at time.Time) string {
	return fmt.Sprintf("%s%s_%s%s", s.config.Prefix, s.config.DatabaseName, at.UTC().Format(timestampLayout), dumpExtension)
}

// deleteExpired removes the dumps older than the retention period. It never deletes the dump just
// uploaded (keep), so a misconfigured clock or retention can't leave the bucket without backups.
func (s *Service) deleteExpired(ctx context.Context, keep string) ([]string, error) {
	if s.config.Retention <= 0 {
		return nil, nil
	}

	objects, err := s.storage.ListObjects(ctx, s.config.Prefix)
	if err != nil {
		return nil, err
	}

	cutoff := s.now().Add(-s.config.Retention)
	sort.Slice(objects, func(i, j int) bool { return objects[i].Key < objects[j].Key })

	var deleted []string
	var errs []error
	for _, obj := range objects {
		if obj.Key == keep || !strings.HasSuffix(obj.Key, dumpExtension) || !obj.LastModified.Before(cutoff) {
			continue
		}
		if err := s.storage.DeleteFile(ctx, obj.Key); err != nil {
			errs = append(errs, fmt.Errorf("deleting %s: %w", obj.Key, err))
			continue
		}
		deleted = append(deleted, obj.Key)
		s.logger.Info("expired database backup deleted", zap.String("key", obj.Key))
	}
	return deleted, errors.Join(errs...)
}
