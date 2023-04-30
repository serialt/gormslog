package gormslog

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/exp/slog"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Handler struct {
	Debug                 bool
	SkipErrRecordNotFound bool
	SlowThreshold         time.Duration
}

var _ logger.Interface = (*Handler)(nil)

func (h *Handler) LogMode(logger.LogLevel) logger.Interface {
	return h
}

func (h *Handler) Info(ctx context.Context, str string, args ...interface{}) {
	slog.InfoCtx(ctx, str, args...)
}

func (h *Handler) Warn(ctx context.Context, str string, args ...interface{}) {
	slog.WarnCtx(ctx, str, args...)
}

func (h *Handler) Error(ctx context.Context, str string, args ...interface{}) {
	slog.ErrorCtx(ctx, str, args...)
}

func (h *Handler) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin)
	sql, _ := fc()

	if err != nil && !(errors.Is(err, gorm.ErrRecordNotFound) && h.SkipErrRecordNotFound) {
		slog.ErrorCtx(ctx, fmt.Sprintf("%s [%s]", sql, elapsed), slog.String("error", err.Error()))
		return
	}

	if h.SlowThreshold != 0 && elapsed > h.SlowThreshold {
		slog.WarnCtx(ctx, fmt.Sprintf("%s [%s]", sql, elapsed))
		return
	}

	if h.Debug {
		slog.DebugCtx(ctx, fmt.Sprintf("%s [%s]", sql, elapsed))
	}
}
