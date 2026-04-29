package observability

import (
	"github.com/bitik/backend/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(cfg config.ObservabilityConfig) (*zap.Logger, error) {
	level := zapcore.InfoLevel
	if err := level.UnmarshalText([]byte(cfg.LogLevel)); err != nil {
		return nil, err
	}

	zapConfig := zap.NewProductionConfig()
	zapConfig.Level = zap.NewAtomicLevelAt(level)
	zapConfig.Encoding = "json"

	if cfg.LogLevel == "debug" {
		zapConfig.Development = true
	}

	return zapConfig.Build()
}
