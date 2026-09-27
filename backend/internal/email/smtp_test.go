package email

import (
	"testing"

	"github.com/eaguilar88/deu/internal/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewSMTPClient(t *testing.T) {
	tests := []struct {
		name     string
		from     string
		username string
		wantErr  bool
	}{
		{name: "valid from address", from: "deu@example.com", username: "deu@example.com"},
		{name: "empty from falls back to username", from: "", username: "deu@example.com"},
		{name: "display name only is rejected", from: "Sistema de Reservas Espacios Universitarios", username: "deu@example.com", wantErr: true},
		{name: "empty from and username is rejected", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.EmailConfig{
				Domain:   "smtp.gmail.com",
				Port:     587,
				Username: tt.username,
				APIKey:   "secret",
				From:     tt.from,
			}

			client, err := NewSMTPClient(cfg, zap.NewNop())
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, client)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, client)
		})
	}
}
