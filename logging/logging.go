package logging

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	l    *zap.Logger
	once sync.Once
)

// InitConfig create singlton of Config struct
func InitLogger(level int) *zap.Logger {
	once.Do(func() {
		var zapLevel zapcore.Level
		switch level {
		case 0:
			zapLevel = zapcore.DebugLevel
		case 1:
			zapLevel = zapcore.InfoLevel
		case 2:
			zapLevel = zapcore.WarnLevel
		case 3:
			zapLevel = zapcore.ErrorLevel
		case 4:
			zapLevel = zapcore.DPanicLevel
		case 5:
			zapLevel = zapcore.PanicLevel
		case 6:
			zapLevel = zapcore.FatalLevel
		default:
			panic("invalid log level choice")
		}

		config := zap.NewProductionConfig()
		// config.EncoderConfig.TimeKey = "timestamp"
		// config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
		config.Level.SetLevel(zapLevel)

		logger, err := config.Build()
		if err != nil {
			panic(err)
		}

		// logger, _ := zap.NewDevelopment()
		// logger, _ := zap.NewProduction()
		defer logger.Sync() // flushes buffer, if any

		l = logger
	})

	return l
}
