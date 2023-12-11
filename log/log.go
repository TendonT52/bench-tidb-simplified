package log

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

func New() (*zap.Logger, error) {
	if logger != nil {
		return logger, nil
	}

	var level zapcore.Level
	switch *logLevel {
	case "error":
		level = zap.ErrorLevel
	case "warn":
		level = zap.WarnLevel
	case "info":
		level = zap.InfoLevel
	case "debug":
		level = zap.DebugLevel
	default:
		level = zap.InfoLevel
	}

	customTimeEncoder := func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("15:04:05"))
	}

	config := zap.Config{
		Level:       zap.NewAtomicLevelAt(level),
		Development: true,
		Encoding:    "console",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     customTimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
	var err error
	logger, err = config.Build()
	if err != nil {
		return nil, fmt.Errorf("error building logger: %w", err)
	}
	logger.Sugar().Infof("Log level: %s", level.String())
	return logger, nil
}
