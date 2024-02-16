package config

import (
	"github.com/go-kit/log"
)

type Server struct {
	HTTPPort         int    `env:"HTTP_SERVE_PORT" envDefault:"80" envWhitelisted:"true"`
	JWTEncryptionKey string `env:"JWT_SIGNING_KEY,required"`
}

// Read current server config - specific for each application
func Read(logger log.Logger) (Server, error) {
	var config Server

	// for _, target := range []interface{}{
	// 	&config,
	// } {
	// 	if err := env.Parse(target); err != nil {
	// 		return config, err
	// 	}
	// }

	logger.Log("msg", "Config successfully loaded")
	return config, nil
}
