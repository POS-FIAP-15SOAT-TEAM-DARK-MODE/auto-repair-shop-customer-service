package logger

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/env"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type customLogger struct {
	*zap.Logger
}

type ctxKey struct{}

// defaultServiceName is the value of the `service` field when SERVICE_NAME is not set.
const defaultServiceName = "customer-service"

var (
	globalLog  *customLogger
	dsnPattern *regexp.Regexp
)

func init() {
	initGlobalLogger()
}

func initGlobalLogger() {
	globalLog = &customLogger{initLogger()}
	dsnPattern = regexp.MustCompile(`://([^:/\s]+):([^@/\s]+)@`)
}

func initLogger() *zap.Logger {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.MessageKey = "message"
	cfg.Level = getLogLevel(env.GetString("LOG_LEVEL", "INFO"))

	l, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	return l.With(zap.String("service", serviceName()))
}

func serviceName() string {
	return env.GetString("SERVICE_NAME", defaultServiceName)
}

func getLogLevel(level string) zap.AtomicLevel {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return zap.NewAtomicLevelAt(zap.DebugLevel)
	case "PANIC":
		return zap.NewAtomicLevelAt(zap.PanicLevel)
	default:
		return zap.NewAtomicLevelAt(zap.InfoLevel)
	}
}

func Global() *customLogger {
	return globalLog
}

// Request attaches a per-request logger to ctx, tagging every log line with
// a request_id. If incomingID is non-empty (e.g. forwarded by an upstream
// API Gateway or a client-supplied X-Request-Id), it's reused as-is so log
// lines for the same request stay correlated across services; otherwise a
// new one is generated.
func Request(ctx context.Context, incomingID string) (context.Context, string) {
	reqId := incomingID
	if reqId == "" {
		reqId = uuid.New().String()
	}
	l := globalLog.Logger.With(
		zap.String("request_id", reqId),
	)
	return withContext(ctx, &customLogger{l}), reqId
}

func Of(ctx context.Context) *customLogger {
	if ctx == nil {
		return Global()
	}

	if l, ok := ctx.Value(ctxKey{}).(*customLogger); ok {
		return l
	}
	return Global()
}

func withContext(ctx context.Context, l *customLogger) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}

func (l *customLogger) Info(msg string, fields ...zap.Field) {
	l.Logger.Info(msg, sanitizeFields(fields)...)
}

func (l *customLogger) Warn(msg string, fields ...zap.Field) {
	l.Logger.Warn(msg, sanitizeFields(fields)...)
}

func (l *customLogger) Error(err error, fields ...zap.Field) {
	fields = append(fields, zap.Error(err))
	l.Logger.Error(err.Error(), sanitizeFields(fields)...)
}

func (l *customLogger) Fatal(msg string, fields ...zap.Field) {
	l.Logger.Fatal(msg, sanitizeFields(fields)...)
}

func (l *customLogger) Debug(msg string, fields ...zap.Field) {
	l.Logger.Debug(msg, sanitizeFields(fields)...)
}

func (l *customLogger) Sync() {
	if l != nil {
		if err := l.Logger.Sync(); err != nil {
			err = fmt.Errorf("failed to sync logger: %w", err)
			l.Error(err)
		}
	}
}

func sanitizeFields(fields []zap.Field) []zap.Field {
	sanitized := make([]zap.Field, 0, len(fields))
	for _, field := range fields {
		key := strings.ToLower(strings.TrimSpace(field.Key))
		if isSensitiveKey(key) {
			sanitized = append(sanitized, zap.String(field.Key, "[REDACTED]"))
			continue
		}
		if field.Type == zapcore.StringType {
			value := redactDSNPassword(field.String)
			sanitized = append(sanitized, zap.String(field.Key, value))
			continue
		}
		sanitized = append(sanitized, field)
	}
	return sanitized
}

// RedactQuery returns a raw URL query with the values of sensitive parameters
// (CPF, CNPJ, tokens) replaced by [REDACTED]. Other pairs are kept as received.
func RedactQuery(rawQuery string) string {
	if rawQuery == "" {
		return ""
	}

	pairs := strings.Split(rawQuery, "&")
	for i, pair := range pairs {
		rawKey, _, _ := strings.Cut(pair, "=")
		key, err := url.QueryUnescape(rawKey)
		if err != nil {
			pairs[i] = "[REDACTED]"
			continue
		}
		if isSensitiveKey(strings.ToLower(key)) {
			pairs[i] = rawKey + "=[REDACTED]"
		}
	}
	return strings.Join(pairs, "&")
}

func isSensitiveKey(key string) bool {
	return key == "password" ||
		key == "senha" ||
		key == "pwd" ||
		key == "token" ||
		key == "jwt" ||
		key == "access_token" ||
		key == "x-internal-token" ||
		key == "authorization" ||
		key == "cpf" ||
		key == "cnpj" ||
		key == "document" ||
		key == "documento"
}

func redactDSNPassword(value string) string {
	return dsnPattern.ReplaceAllString(value, `://$1:[REDACTED]@`)
}
