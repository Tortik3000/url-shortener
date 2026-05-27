package logger

import "go.uber.org/zap"

type zapLogger struct {
	z *zap.Logger
}

func NewZap(z *zap.Logger) *zapLogger {
	return &zapLogger{
		z: z,
	}
}

func (l *zapLogger) Debug(msg string, fields ...Field) { l.z.Debug(msg, toZapFields(fields)...) }
func (l *zapLogger) Info(msg string, fields ...Field)  { l.z.Info(msg, toZapFields(fields)...) }
func (l *zapLogger) Warn(msg string, fields ...Field)  { l.z.Warn(msg, toZapFields(fields)...) }
func (l *zapLogger) Error(msg string, fields ...Field) { l.z.Error(msg, toZapFields(fields)...) }
func (l *zapLogger) Fatal(msg string, fields ...Field) { l.z.Fatal(msg, toZapFields(fields)...) }

func (l *zapLogger) With(fields ...Field) Logger {
	return &zapLogger{z: l.z.With(toZapFields(fields)...)}
}

func (l *zapLogger) Sync() error {
	return l.z.Sync()
}

func toZapFields(fields []Field) []zap.Field {
	out := make([]zap.Field, 0, len(fields))
	for _, f := range fields {
		if err, ok := f.Value.(error); ok {
			out = append(out, zap.NamedError(f.Key, err))
			continue
		}
		out = append(out, zap.Any(f.Key, f.Value))
	}
	return out
}
