package email

import (
	"context"
	"time"

	"github.com/eaguilar88/deu/internal/config"
	"github.com/mailgun/mailgun-go/v5"
	"go.uber.org/zap"
)

type MailgunClient struct {
	from   string
	domain string
	client *mailgun.Client
	logger *zap.Logger
}

func NewMailgunClient(config config.EmailConfig, logger *zap.Logger) *MailgunClient {
	c := mailgun.NewMailgun(config.APIKey)
	return &MailgunClient{
		from:   config.From,
		domain: config.Domain,
		client: c,
	}
}

func (m *MailgunClient) Send(ctx context.Context, to string, subject string, body string) error {
	message := mailgun.NewMessage(m.domain, m.from, subject, body, to)
	mailCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	_, err := m.client.Send(mailCtx, message)
	if err != nil {
		m.logger.Error("failed to send email", zap.Error(err))
		return err
	}
	return nil
}
