package config

import (
	"fmt"

	"github.com/caarlos0/env/v10"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/joho/godotenv"
)

type Server struct {
	HTTPPort         int    `env:"HTTP_SERVE_PORT" envDefault:"80" envWhitelisted:"true"`
	JWTEncryptionKey string `env:"JWT_SIGNING_KEY,required"`
	FilePath         string `env:"FILE_PATH,required"`
	Database         DatabaseConfig
	Email            EmailConfig
}

type DatabaseConfig struct {
	Hostname string `env:"POSTGRES_HOST" envDefault:"localhost" envWhitelisted:"true"`
	Name     string `env:"POSTGRES_DB" envDefault:"deu" envWhitelisted:"true"`
	User     string `env:"POSTGRES_USERNAME" envDefault:"root" envWhitelisted:"true"`
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
func Read(logger log.Logger) (Server, error) {
	var config Server
	// Loading the environment variables from '.env' file.
	err := godotenv.Load("./.env")
	if err != nil {
		_ = level.Error(logger).Log("msg", "failed to load env file", "error", err)
	}

	if err := env.Parse(&config); err != nil {
		_ = level.Error(logger).Log("msg", "failed to parse configuration", "error", err)
		return config, err
	}

	if err := env.Parse(&config.Database); err != nil {
		_ = level.Error(logger).Log("msg", "failed to parse database configuration", "error", err)
		return config, err
	}

	if err := env.Parse(&config.Email); err != nil {
		_ = level.Error(logger).Log("msg", "failed to parse email configuration", "error", err)
		return config, err
	}

	logger.Log("msg", "Config successfully loaded")
	return config, nil
}
