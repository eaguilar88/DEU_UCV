// Command backup makes the PostgreSQL backups: it dumps the database with pg_dump, uploads the
// dump to Backblaze B2 under BACKUP_PREFIX and deletes the dumps older than BACKUP_RETENTION_DAYS.
// It runs once on start and then every day at BACKUP_RUN_AT (BACKUP_TIMEZONE).
//
// With -once it makes a single backup and exits with a non-zero status if it failed.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata" // the alpine image has no zoneinfo for time.LoadLocation

	"github.com/eaguilar88/deu/internal/backups"
	"github.com/eaguilar88/deu/internal/config"
	"github.com/eaguilar88/deu/internal/schedule"
	"github.com/eaguilar88/deu/internal/storage"
	"go.uber.org/zap"
)

func main() {
	once := flag.Bool("once", false, "make a single backup and exit")
	flag.Parse()

	logger, err := config.NewLogger()
	if err != nil {
		os.Exit(1)
	}
	defer logger.Sync() //nolint: errcheck

	if err := run(logger, *once); err != nil {
		logger.Error("backup stopped", zap.Error(err))
		os.Exit(1)
	}
}

func run(logger *zap.Logger, once bool) error {
	cfg, err := config.ReadBackup(logger)
	if err != nil {
		return fmt.Errorf("parsing configuration: %w", err)
	}
	hour, minute, loc, err := schedule.ParseDaily(cfg.RunAt, cfg.TimeZone)
	if err != nil {
		return fmt.Errorf("parsing BACKUP_RUN_AT/BACKUP_TIMEZONE: %w", err)
	}

	b2, err := storage.NewB2Client(
		cfg.BlackBlazeB2.BucketName,
		cfg.BlackBlazeB2.KeyID,
		cfg.BlackBlazeB2.ApplicationKey,
		cfg.BlackBlazeB2.Endpoint,
		cfg.BlackBlazeB2.Region,
		"",
		logger,
	)
	if err != nil {
		return fmt.Errorf("creating b2 client: %w", err)
	}

	dumper := backups.PgDump{
		Host:     cfg.Database.Hostname,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		Database: cfg.Database.Name,
	}
	svc := backups.NewService(dumper, b2, backups.Config{
		DatabaseName: cfg.Database.Name,
		Prefix:       cfg.Prefix,
		Retention:    time.Duration(cfg.RetentionDays) * 24 * time.Hour,
	}, logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	backup := func() error {
		runCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
		defer cancel()
		result, err := svc.Run(runCtx)
		if err != nil {
			return err
		}
		logger.Info("backup finished",
			zap.String("key", result.Key),
			zap.Int64("bytes", result.Size),
			zap.Int("expired_deleted", len(result.Deleted)))
		return nil
	}

	if once {
		return backup()
	}

	logger.Info("backup service started",
		zap.String("run_at", cfg.RunAt),
		zap.String("time_zone", cfg.TimeZone),
		zap.Int("retention_days", cfg.RetentionDays))
	for {
		// A failed run is logged and retried on the next scheduled run.
		if err := backup(); err != nil {
			logger.Error("backup run failed", zap.Error(err))
		}

		next := schedule.NextRun(time.Now(), hour, minute, loc)
		logger.Info("next backup run", zap.Time("at", next))
		timer := time.NewTimer(time.Until(next))
		select {
		case <-ctx.Done():
			timer.Stop()
			logger.Info("backup service stopped")
			return nil
		case <-timer.C:
		}
	}
}
