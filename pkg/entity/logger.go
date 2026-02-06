package entity

import (
	"context"
	"time"

	logrus "github.com/sirupsen/logrus"
	gorm_logger "gorm.io/gorm/logger"
)

type Logger struct {
	LogLevel                  gorm_logger.LogLevel
	SlowThreshold             time.Duration
	IgnoreRecordNotFoundError bool
}

func (l *Logger) LogMode(level gorm_logger.LogLevel) gorm_logger.Interface {
	newlogger := *l
	newlogger.LogLevel = level
	return &newlogger
}

func (l *Logger) Info(ctx context.Context, s string, args ...any) {
	logrus.WithContext(ctx).Infof(s, args...)
}

func (l *Logger) Warn(ctx context.Context, s string, args ...any) {
	logrus.WithContext(ctx).Warnf(s, args...)
}

func (l *Logger) Error(ctx context.Context, s string, args ...any) {
	logrus.WithContext(ctx).Errorf(s, args...)
}

func (l *Logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if logrus.GetLevel() >= logrus.TraceLevel {
		sql, _ := fc()
		elapsed := time.Since(begin)
		if elapsed > l.SlowThreshold {
			logrus.WithContext(ctx).Tracef("%s [%s]", sql, elapsed)
		}
	}
}
