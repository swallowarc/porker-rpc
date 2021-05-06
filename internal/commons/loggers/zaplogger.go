package loggers

import (
	"log"

	"go.uber.org/zap"
)

func NewZapLogger(develop bool) *zap.Logger {
	if !develop {
		return getInstance(zap.NewProductionConfig())
	}
	c := zap.NewDevelopmentConfig()
	c.DisableStacktrace = true
	return getInstance(c)
}

// getInstance get zap logger instance.
func getInstance(config zap.Config) *zap.Logger {
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}
	zapLogger, err := config.Build()
	if err != nil {
		log.Fatalf("create logger failed: %v", err)
	}
	return zapLogger
}
