package email

import (
	"context"
	"fmt"
	"time"

	"github.com/eaguilar88/deu/internal/config"
	"github.com/wneessen/go-mail"
	"go.uber.org/zap"
)

const implicitTLSPort = 465

type SMTPClient struct {
	from   string
	client *mail.Client
	logger *zap.Logger
}

func NewSMTPClient(config config.EmailConfig, logger *zap.Logger) (*SMTPClient, error) {
	opts := []mail.Option{
		mail.WithPort(config.Port),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(config.Username),
		mail.WithPassword(config.APIKey),
		mail.WithTimeout(30 * time.Second),
	}
	if config.Port == implicitTLSPort {
		opts = append(opts, mail.WithSSLPort(false))
	} else {
		opts = append(opts, mail.WithTLSPortPolicy(mail.TLSMandatory))
	}

	c, err := mail.NewClient(config.Domain, opts...)
	if err != nil {
		return nil, fmt.Errorf("creating smtp client: %w", err)
	}

	from := config.From
	if from == "" {
		from = config.Username
	}
	if err := mail.NewMsg().From(from); err != nil {
		return nil, fmt.Errorf("invalid sender address %q: %w", from, err)
	}

	return &SMTPClient{
		from:   from,
		client: c,
		logger: logger,
	}, nil
}

func (s *SMTPClient) Send(ctx context.Context, to string, subject string, body string) error {
	message := mail.NewMsg()
	if err := message.From(s.from); err != nil {
		s.logger.Error("invalid sender address", zap.Error(err), zap.String("from", s.from))
		return err
	}
	if err := message.To(to); err != nil {
		s.logger.Error("invalid recipient address", zap.Error(err), zap.String("to", to))
		return err
	}
	message.Subject(subject)
	message.SetBodyString(mail.TypeTextPlain, body)

	mailCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := s.client.DialAndSendWithContext(mailCtx, message); err != nil {
		s.logger.Error("failed to send email", zap.Error(err))
		return err
	}
	return nil
}
