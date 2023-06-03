package gormslog

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/exp/slog"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	gormlogger "gorm.io/gorm/logger"
)

type Logger struct {
	LogLevel                  gormlogger.LogLevel
	SlowThreshold             time.Duration
	SkipCallerLookup          bool
	IgnoreRecordNotFoundError bool
}

func New(log *slog.Logger) Logger {
	// slog.SetDefault(log)
	return Logger{
		LogLevel:                  gormlogger.Warn,
		SlowThreshold:             100 * time.Millisecond,
		SkipCallerLookup:          false,
		IgnoreRecordNotFoundError: false,
	}
}

func (l Logger) SetAsDefault() {
	gormlogger.Default = l
}

func (l Logger) LogMode(level gormlogger.LogLevel) logger.Interface {
	return Logger{
		LogLevel:                  level,
		SlowThreshold:             l.SlowThreshold,
		SkipCallerLookup:          l.SkipCallerLookup,
		IgnoreRecordNotFoundError: l.IgnoreRecordNotFoundError,
	}

}

func (l Logger) Info(ctx context.Context, str string, args ...interface{}) {
	if l.LogLevel < gormlogger.Info {
		return
	}
	slog.InfoCtx(ctx, str, args...)
}

func (l Logger) Warn(ctx context.Context, str string, args ...interface{}) {
	if l.LogLevel < gormlogger.Warn {
		return
	}
	slog.WarnCtx(ctx, str, args...)
}

func (l Logger) Error(ctx context.Context, str string, args ...interface{}) {
	if l.LogLevel < gormlogger.Error {
		return
	}
	slog.ErrorCtx(ctx, str, args...)
}

func (l Logger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.LogLevel <= 0 {
		return
	}
	elapsed := time.Since(begin)

	switch {
	case err != nil && l.LogLevel >= gormlogger.Error && (!l.IgnoreRecordNotFoundError || !errors.Is(err, gorm.ErrRecordNotFound)):
		sql, rows := fc()
		slog.ErrorCtx(ctx, fmt.Sprint(err), "elapsed", elapsed, "rows", rows, "sql", sql)
	case l.SlowThreshold != 0 && elapsed > l.SlowThreshold && l.LogLevel >= gormlogger.Warn:
		sql, rows := fc()
		slog.WarnCtx(ctx, fmt.Sprint(err), "elapsed", elapsed, "rows", rows, "sql", sql)

	case l.LogLevel >= gormlogger.Info:
		sql, rows := fc()
		slog.DebugCtx(ctx, fmt.Sprint(err), "elapsed", elapsed, "rows", rows, "sql", sql)
	}
}
