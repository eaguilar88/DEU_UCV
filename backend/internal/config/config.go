package config

import (
	"fmt"
	"log"
	"os"
	"strings"

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
	Hostname string `env:"HOST,required"`
	Name     string `env:"DB,required"`
	User     string `env:"USER,required"`
	Password string `env:"PASSWORD,required"`
	Port     int    `env:"PORT,required"`
}

type EmailConfig struct {
	Domain string `env:"DOMAIN"   envDefault:"smtp.gmail.com"`
	APIKey string `env:"API_KEY"  envDefault:"DEU"`
	From   string `env:"FROM"`
	Port   int    `env:"PORT"     envDefault:"587"`
}

type BlackBlazeB2Config struct {
	BucketName     string `env:"BUCKET_NAME"              envDefault:"deu"`
	Endpoint       string `env:"ENDPOINT"                 envDefault:"s3.us-west-002.backblazeb2.com"`
	KeyName        string `env:"KEY_NAME"                 envDefault:"deu"`
	Region         string `env:"REGION"                   envDefault:"us-west-002"`
	ApplicationKey string `env:"APPLICATION_KEY"`
	KeyID          string `env:"KEY_ID"`
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

	if config.IsProd() {
		config.BlackBlazeB2.ApplicationKey = readSecret("/run/secrets/b2_application_key")
		config.BlackBlazeB2.KeyID = readSecret("/run/secrets/b2_key_id")
	}

	return config, nil
}

// Read Docker Secrets from file
func readSecret(filePath string) string {
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Failed to read secret from %s: %v", filePath, err)
	}
	return strings.TrimSpace(string(data))
}
