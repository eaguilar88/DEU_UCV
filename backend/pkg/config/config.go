package config

import (
	"fmt"
	"log"
	"os"

	"github.com/caarlos0/env/v10"
	"go.uber.org/zap"
)

type FlavorName string

const (
	FlavorDev  FlavorName = "dev"
	FlavorProd FlavorName = "prod"
)

type DeuConfig struct {
	HTTPPort         int                `env:"HTTP_SERVE_PORT" envDefault:"80" envWhitelisted:"true"`
	JWTEncryptionKey string             `env:"JWT_SIGNING_KEY,required"`
	FilePath         string             `env:"FILE_PATH,required"`
	TTL              uint32             `env:"TOKEN_TTL" envDefault:"3600"`
	Flavor           string             `env:"FLAVOR" envDefault:"dev"`
	Database         DatabaseConfig     `envPrefix:"POSTGRES_"`
	Email            EmailConfig        `envPrefix:"EMAIL_"`
	BlackBlazeB2     BlackBlazeB2Config `envPrefix:"BLACKBLAZE_B2_"`
}

func (s *DeuConfig) IsProd() bool {
	return s.Flavor == string(FlavorProd)
}

type DatabaseConfig struct {
	Hostname string `env:"HOST" envDefault:"localhost" envWhitelisted:"true"`
	Name     string `env:"DB" envDefault:"deu" envWhitelisted:"true"`
	User     string `env:"USER" envDefault:"root" envWhitelisted:"true"`
	Password string `env:"PASSWORD" envDefault:"root" envWhitelisted:"true"`
	Port     int    `env:"PORT" envDefault:"3306" envWhitelisted:"true"`
}

type EmailConfig struct {
	Server   string `env:"SERVER" envDefault:"localhost" envWhitelisted:"true"`
	Name     string `env:"NAME" envDefault:"deu" envWhitelisted:"true"`
	User     string `env:"USERNAME" envDefault:"root" envWhitelisted:"true"`
	Password string `env:"PASSWORD" envDefault:"root" envWhitelisted:"true"`
	Port     int    `env:"PORT" envDefault:"3306" envWhitelisted:"true"`
}

type BlackBlazeB2Config struct {
	BucketName     string `env:"BUCKET_NAME" envDefault:"deu-ucv" envWhitelisted:"true"`
	Endpoint       string `env:"ENDPOINT" envDefault:"s3.us-east-005.backblazeb2.com" envWhitelisted:"true"`
	KeyName        string `env:"KEY_NAME" envDefault:"deu" envWhitelisted:"true"`
	Region         string `env:"REGION" envDefault:"us-east-005" envWhitelisted:"true"`
	ApplicationKey string `env:"APPLICATION_KEY" envDefault:"" envWhitelisted:"true"`
	KeyID          string `env:"KEY_ID" envDefault:"" envWhitelisted:"true"`
}

func (cfg DatabaseConfig) String() string {
	return fmt.Sprintf("{Hostname: %s Name: %s User: %s Password: {private} Port: %d}",
		cfg.Hostname, cfg.Name, cfg.User, cfg.Port)
}

// Read current server config - specific for each application
func Read(logger *zap.Logger) (DeuConfig, error) {
	var config DeuConfig

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

	if err := env.Parse(&config.BlackBlazeB2); err != nil {
		logger.Error("failed to parse blackblaze b2 configuration", zap.Error(err))
		return config, err
	}

	if config.IsProd() {
		config.BlackBlazeB2.ApplicationKey = readSecret("/run/secrets/blackblaze_b2_application_key")
		config.BlackBlazeB2.KeyID = readSecret("/run/secrets/blackblaze_b2_key_id")
	}

	return config, nil
}

// Read Docker Secrets from file
func readSecret(filePath string) string {
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Failed to read secret from %s: %v", filePath, err)
	}
	return string(data)
}
