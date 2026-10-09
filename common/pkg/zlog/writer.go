package zlog

import (
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
	"go.uber.org/zap"
)

type Field = logx.LogField

func ErrorField(err error) Field {
	return logx.Field("error", err)
}

func DataField(data any) Field {
	return logx.Field("data", data)
}

func ParamField(data any) Field {
	return logx.Field("param", data)
}

type ZapWriter struct {
	logger *zap.Logger
}

func NewZapWriter(logger *zap.Logger) logx.Writer {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &ZapWriter{logger: logger}
}

func (w *ZapWriter) Alert(value any) {
	w.logger.Error(fmt.Sprint(value))
}

func (w *ZapWriter) Close() error {
	return w.logger.Sync()
}

func (w *ZapWriter) Debug(value any, fields ...logx.LogField) {
	w.logger.Debug(fmt.Sprint(value), toZapFields(fields...)...)
}

func (w *ZapWriter) Error(value any, fields ...logx.LogField) {
	w.logger.Error(fmt.Sprint(value), toZapFields(fields...)...)
}

func (w *ZapWriter) Info(value any, fields ...logx.LogField) {
	w.logger.Info(fmt.Sprint(value), toZapFields(fields...)...)
}

func (w *ZapWriter) Severe(value any) {
	w.logger.Error(fmt.Sprint(value), zap.String("severity", "severe"))
}

func (w *ZapWriter) Slow(value any, fields ...logx.LogField) {
	w.logger.Warn(fmt.Sprint(value), toZapFields(fields...)...)
}

func (w *ZapWriter) Stack(value any) {
	w.logger.Error(fmt.Sprint(value), zap.Stack("stack"))
}

func (w *ZapWriter) Stat(value any, fields ...logx.LogField) {
	w.logger.Info(fmt.Sprint(value), toZapFields(fields...)...)
}

func toZapFields(fields ...logx.LogField) []zap.Field {
	result := make([]zap.Field, 0, len(fields))
	for _, field := range fields {
		result = append(result, zap.Any(field.Key, field.Value))
	}
	return result
}
