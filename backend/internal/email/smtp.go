package email

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"time"

	"github.com/eaguilar88/deu/internal/config"
	"github.com/wneessen/go-mail"
	"go.uber.org/zap"
)

const implicitTLSPort = 465

type SMTPClient struct {
	from      string
	client    *mail.Client
	templates map[Template]*template.Template
	logger    *zap.Logger
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

	templates, err := parseTemplates()
	if err != nil {
		return nil, err
	}

	return &SMTPClient{
		from:      from,
		client:    c,
		templates: templates,
		logger:    logger,
	}, nil
}

// Attachment is a file attached to an email.
type Attachment struct {
	Name        string
	ContentType string
	Data        []byte
}

// SendTemplate renders the given email template with data and sends it to the recipient as HTML.
func (s *SMTPClient) SendTemplate(ctx context.Context, to string, tmpl Template, data any) error {
	return s.SendTemplateWithAttachments(ctx, to, tmpl, data, nil)
}

// SendTemplateWithAttachments renders the given email template with data and sends it to the
// recipient as HTML, with the given files attached.
func (s *SMTPClient) SendTemplateWithAttachments(ctx context.Context, to string, tmpl Template, data any, attachments []Attachment) error {
	subject, body, err := render(s.templates, tmpl, data)
	if err != nil {
		s.logger.Error("failed to render email template", zap.Error(err), zap.String("template", string(tmpl)))
		return err
	}

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
	message.SetBodyString(mail.TypeTextHTML, body)

	for _, a := range attachments {
		if err := message.AttachReader(a.Name, bytes.NewReader(a.Data), mail.WithFileContentType(mail.ContentType(a.ContentType))); err != nil {
			s.logger.Error("failed to attach file to email", zap.Error(err), zap.String("attachment", a.Name))
			return err
		}
	}

	mailCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := s.client.DialAndSendWithContext(mailCtx, message); err != nil {
		s.logger.Error("failed to send email", zap.Error(err))
		return err
	}
	return nil
}
