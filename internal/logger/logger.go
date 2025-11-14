package logger

import (
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"mPR/internal/config"
)

func New(cfg config.Config) *zap.Logger {
	level := parse(cfg.Log.Level)

	var zapCfg zap.Config
	if cfg.App.Env == "prod" {
		zapCfg = zap.NewProductionConfig()
	} else {
		zapCfg = zap.NewDevelopmentConfig()
	}

	zapCfg.Level = zap.NewAtomicLevelAt(level)

	logger, err := zapCfg.Build(zap.AddCaller())
	if err != nil {
		log.Fatalf("Error create logger: %v", err)
	}

	return logger
}

func parse(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		log.Println("Wrong level log, use default: info")
		return zapcore.InfoLevel
	}
}
