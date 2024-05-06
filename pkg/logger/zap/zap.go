package zap

import (
	"github.com/Neeeooshka/alice-skill.git/pkg/logger"
	"go.uber.org/zap"
)

type zapLogger struct {
	logger *zap.Logger
}

var Log = &zapLogger{logger: zap.NewNop()}

func (l *zapLogger) init(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	zl, err := cfg.Build()
	if err != nil {
		return err
	}
	l.logger = zl
	return nil
}

func (l *zapLogger) Log(rq logger.RequestData, rs logger.ResponseData) {
	l.logger.Info("receive new request",
		zap.String("URI", rq.URI),
		zap.String("method", rq.Method),
		zap.Duration("duration", rq.Duration),
		zap.Int("status", rs.Status),
		zap.Int("size", rs.Size),
	)
}

func (l *zapLogger) Debug(msg string, fields ...zap.Field) {
	l.logger.Debug(msg, fields...)
}

func (l *zapLogger) String(key string, val string) zap.Field {
	return zap.String(key, val)
}

func (l *zapLogger) Error(err error) zap.Field {
	return zap.Error(err)
}

func (l *zapLogger) Info(msg string, fields ...zap.Field) {
	l.logger.Info(msg, fields...)
}

func NewZapLogger(level string) (*zapLogger, error) {
	return Log, Log.init(level)
}
