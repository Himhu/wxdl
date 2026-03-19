package utils

import (
	"strings"

	"backend/internal/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(cfg config.LogConfig, env string) (*zap.Logger, error) {
	var zapConfig zap.Config
	if strings.EqualFold(env, "production") {
		zapConfig = zap.NewProductionConfig()
	} else {
		zapConfig = zap.NewDevelopmentConfig()
	}

	level := new(zapcore.Level)
	if err := level.UnmarshalText([]byte(strings.ToLower(cfg.Level))); err != nil {
		*level = zapcore.InfoLevel
	}
	zapConfig.Level = zap.NewAtomicLevelAt(*level)

	if cfg.Encoding != "" {
		zapConfig.Encoding = cfg.Encoding
	}

	return zapConfig.Build()
}

func SyncLogger(logger *zap.Logger) {
	if logger != nil {
		_ = logger.Sync()
	}
}
