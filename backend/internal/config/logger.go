package config

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger() (*zap.Logger, error) {
	logConfig := zap.Config{
		Level:            zap.NewAtomicLevelAt(zap.DebugLevel), // Adjust log level as needed
		Encoding:         "json",                               // Use "console" for human-readable logs
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:   "message",
			LevelKey:     "level",
			TimeKey:      "time",
			CallerKey:    "caller",
			EncodeTime:   zapcore.RFC3339TimeEncoder,
			EncodeLevel:  zapcore.LowercaseLevelEncoder,
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
		DisableStacktrace: true, // This removes stack traces for all logs
	}
	return logConfig.Build()
}
