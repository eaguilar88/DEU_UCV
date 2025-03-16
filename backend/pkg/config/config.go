package config

import (
	"fmt"

	"github.com/caarlos0/env/v10"
	"go.uber.org/zap"
)

type Server struct {
	HTTPPort         int    `env:"HTTP_SERVE_PORT" envDefault:"80" envWhitelisted:"true"`
	JWTEncryptionKey string `env:"JWT_SIGNING_KEY,required"`
	FilePath         string `env:"FILE_PATH,required"`
	TTL              uint32 `env:"TOKEN_TTL" envDefault:"3600"`
	Database         DatabaseConfig
	Email            EmailConfig
}

type DatabaseConfig struct {
	Hostname string `env:"POSTGRES_HOST" envDefault:"localhost" envWhitelisted:"true"`
	Name     string `env:"POSTGRES_DB" envDefault:"deu" envWhitelisted:"true"`
	User     string `env:"POSTGRES_USER" envDefault:"root" envWhitelisted:"true"`
	Password string `env:"POSTGRES_PASSWORD" envDefault:"root" envWhitelisted:"true"`
	Port     int    `env:"POSTGRES_PORT" envDefault:"3306" envWhitelisted:"true"`
}

type EmailConfig struct {
	Server   string `env:"EMAIL_SERVER" envDefault:"localhost" envWhitelisted:"true"`
	Name     string `env:"EMAIL_NAME" envDefault:"deu" envWhitelisted:"true"`
	User     string `env:"EMAIL_USERNAME" envDefault:"root" envWhitelisted:"true"`
	Password string `env:"EMAIL_PASSWORD" envDefault:"root" envWhitelisted:"true"`
	Port     int    `env:"EMAIL_PORT" envDefault:"3306" envWhitelisted:"true"`
}

func (cfg DatabaseConfig) String() string {
	return fmt.Sprintf("{Hostname: %s Name: %s User: %s Password: {private} Port: %d}",
		cfg.Hostname, cfg.Name, cfg.User, cfg.Port)
}

// Read current server config - specific for each application
func Read(logger *zap.Logger) (Server, error) {
	var config Server

	if err := env.Parse(&config); err != nil {
		logger.Error("failed to parse configuration", zap.Error(err))
		return config, err
	}

	if err := env.Parse(&config.Database); err != nil {
		logger.Error("failed to parse database configuration", zap.Error(err))
		return config, err
	}

	if err := env.Parse(&config.Email); err != nil {
		logger.Error("failed to parse email configuration", zap.Error(err))
		return config, err
	}

	return config, nil
}
