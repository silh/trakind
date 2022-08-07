package logger

import "go.uber.org/zap"

var log *zap.SugaredLogger

func init() {
	config := zap.NewDevelopmentConfig()
	config.Encoding = "console"
	config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	logger, err := config.Build()
	if err != nil {
		zap.S().Fatalw("Failed to create logger", "err", err)
	}
	log = logger.Sugar()
}

// Logger returns the single logger used everywhere.
func Logger() *zap.SugaredLogger {
	return log
}
