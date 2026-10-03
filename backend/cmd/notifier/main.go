// Command notifier sends the reminder emails (pending requests to coordinators, yearly renewal to
// extension groups). It runs once on start and then every day at NOTIFIER_RUN_AT. Every run is
// safe to repeat: deu.notifications_log keeps a reminder from being resent before
// NOTIFIER_RESEND_INTERVAL.
//
// With -once it sends the due reminders and exits.
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata" // the alpine image has no zoneinfo for time.LoadLocation

	"github.com/eaguilar88/deu/internal/config"
	"github.com/eaguilar88/deu/internal/email"
	"github.com/eaguilar88/deu/internal/notifications"
	repository "github.com/eaguilar88/deu/internal/postgres_repository"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func main() {
	once := flag.Bool("once", false, "send the due reminders and exit")
	flag.Parse()

	logger, err := config.NewLogger()
	if err != nil {
		os.Exit(1)
	}
	defer logger.Sync() //nolint: errcheck

	if err := run(logger, *once); err != nil {
		logger.Error("notifier stopped", zap.Error(err))
		os.Exit(1)
	}
}

func run(logger *zap.Logger, once bool) error {
	cfg, err := config.ReadNotifier(logger)
	if err != nil {
		return fmt.Errorf("parsing configuration: %w", err)
	}
	loc, err := time.LoadLocation(cfg.TimeZone)
	if err != nil {
		return fmt.Errorf("loading time zone %q: %w", cfg.TimeZone, err)
	}
	runAt, err := time.Parse("15:04", cfg.RunAt)
	if err != nil {
		return fmt.Errorf("parsing NOTIFIER_RUN_AT %q: %w", cfg.RunAt, err)
	}

	db, err := connectToDB(cfg.Database)
	if err != nil {
		return fmt.Errorf("connecting to the db: %w", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		return fmt.Errorf("pinging the db: %w", err)
	}

	mailClient, err := email.NewSMTPClient(cfg.Email, logger)
	if err != nil {
		return fmt.Errorf("creating smtp client: %w", err)
	}

	repo := repository.NewRepository(db, "", logger)
	svc := notifications.NewService(repo, mailClient, notifications.Config{
		PendingRequestAge: cfg.PendingRequestAge,
		RenewalWindow:     cfg.RenewalWindow,
		ResendInterval:    cfg.ResendInterval,
	}, logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if once {
		return svc.Run(ctx)
	}

	logger.Info("notifier started", zap.String("run_at", cfg.RunAt), zap.String("time_zone", cfg.TimeZone))
	for {
		// A failed run is retried on the next one; reminders already sent are not repeated.
		if err := svc.Run(ctx); err != nil {
			logger.Error("notifier run finished with errors", zap.Error(err))
		}

		next := nextRun(time.Now(), runAt.Hour(), runAt.Minute(), loc)
		logger.Info("next notifier run", zap.Time("at", next))
		timer := time.NewTimer(time.Until(next))
		select {
		case <-ctx.Done():
			timer.Stop()
			logger.Info("notifier stopped")
			return nil
		case <-timer.C:
		}
	}
}

// nextRun returns the first time after now that is hour:minute in loc.
func nextRun(now time.Time, hour, minute int, loc *time.Location) time.Time {
	local := now.In(loc)
	next := time.Date(local.Year(), local.Month(), local.Day(), hour, minute, 0, 0, loc)
	if !next.After(local) {
		next = next.AddDate(0, 0, 1)
	}
	return next
}

func connectToDB(conf config.DatabaseConfig) (*sql.DB, error) {
	connection := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		conf.User,
		conf.Password,
		conf.Hostname,
		conf.Port,
		conf.Name,
	)
	return sql.Open("postgres", connection)
}
