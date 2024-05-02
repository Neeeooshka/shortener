package main

import (
	"github.com/Neeeooshka/alice-skill.git/internal/logger"
	"github.com/sirupsen/logrus"
	"os"
)

type logrusLogger struct {
	logger *logrus.Logger
}

func newLogrusLogger(level logrus.Level) *logrusLogger {
	logrusLogger := &logrusLogger{logrus.New()}
	logrusLogger.logger.SetOutput(os.Stdout)
	logrusLogger.logger.SetLevel(level)

	return logrusLogger
}

func (l *logrusLogger) Log(rq logger.RequestData, rs logger.ResponseData) {
	l.logger.WithFields(logrus.Fields{
		"URI":      rq.URI,
		"method":   rq.Method,
		"duration": rq.Duration,
		"status":   rs.Status,
		"size":     rs.Size,
	}).Info("receive new request")
}
