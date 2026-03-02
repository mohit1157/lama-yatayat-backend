package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(serviceName string, debug bool) *zap.Logger {
	var cfg zap.Config
	if debug {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		cfg = zap.NewProductionConfig()
	}

	logger, _ := cfg.Build()
	return logger.With(zap.String("service", serviceName))
}
