package logger

import (
	"context"
	"os"

	"github.com/snisid/platform/internal/platform/tracing"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

func init() {
	// 1. Loki & OTel Compatible JSON Encoder
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "@timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	encoder := zapcore.NewJSONEncoder(encoderConfig)

	// 2. Output strictly to stdout for K8s Fluent-bit/Promtail ingestion
	stdout := zapcore.Lock(os.Stdout)

	// 3. Sampling: prevent logging DoS during massive traffic spikes
	// Drop 50% of logs after 100 identical logs within a second
	core := zapcore.NewSamplerWithOptions(
		zapcore.NewCore(encoder, stdout, zap.InfoLevel),
		timeTick,
		100, // Initial burst
		2,   // Thereafter drop 1 of 2 (50%)
	)

	Log = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
}

const timeTick = 1 // 1 second for sampling window

// Secret masks sensitive PII/Authentication data in logs
func Secret(key, val string) zap.Field {
	if val == "" {
		return zap.String(key, "")
	}
	return zap.String(key, "***REDACTED***")
}

// injectTracing attempts to pull standard trace headers
func injectTracing(ctx context.Context, fields []zap.Field) []zap.Field {
	if ctx == nil {
		return fields
	}

	// 1. Extract SNISID Correlation ID
	if corrID := tracing.ExtractCorrelationID(ctx); corrID != "" {
		fields = append(fields, zap.String("correlation_id", corrID))
	}

	// 2. Extract OpenTelemetry Trace ID & Span ID
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasTraceID() {
		fields = append(fields, zap.String("trace_id", spanCtx.TraceID().String()))
	}
	if spanCtx.HasSpanID() {
		fields = append(fields, zap.String("span_id", spanCtx.SpanID().String()))
	}

	return fields
}

func Info(ctx context.Context, msg string, fields ...zap.Field) {
	fields = injectTracing(ctx, fields)
	Log.Info(msg, fields...)
}

func Warn(ctx context.Context, msg string, fields ...zap.Field) {
	fields = injectTracing(ctx, fields)
	Log.Warn(msg, fields...)
}

func Debug(ctx context.Context, msg string, fields ...zap.Field) {
	fields = injectTracing(ctx, fields)
	Log.Debug(msg, fields...)
}

func Error(ctx context.Context, msg string, err error, fields ...zap.Field) {
	fields = append(fields, zap.Error(err))
	fields = injectTracing(ctx, fields)
	Log.Error(msg, fields...)
}

func Fatal(ctx context.Context, msg string, err error, fields ...zap.Field) {
	fields = append(fields, zap.Error(err))
	fields = injectTracing(ctx, fields)
	Log.Fatal(msg, fields...)
}

func GetLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}
