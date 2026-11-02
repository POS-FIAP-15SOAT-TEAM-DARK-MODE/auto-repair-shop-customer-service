package logger

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestServiceName_DefaultAndOverride(t *testing.T) {
	t.Setenv("SERVICE_NAME", "")
	assert.Equal(t, defaultServiceName, serviceName())

	t.Setenv("SERVICE_NAME", "customer-service-stg")
	assert.Equal(t, "customer-service-stg", serviceName())
}

func TestRequest_KeepsBaseFieldsAndRequestID(t *testing.T) {
	core, logs := observer.New(zapcore.InfoLevel)
	prev := globalLog
	globalLog = &customLogger{zap.New(core).With(zap.String("service", defaultServiceName))}
	t.Cleanup(func() { globalLog = prev })

	ctx, reqID := Request(context.Background(), "upstream-id")
	Of(ctx).Info("customer.created", zap.String("customer_id", "c-1"))

	assert.Equal(t, "upstream-id", reqID)
	entries := logs.All()
	if assert.Len(t, entries, 1) {
		assert.Equal(t, "customer.created", entries[0].Message)
		fields := entries[0].ContextMap()
		assert.Equal(t, "upstream-id", fields["request_id"])
		assert.Equal(t, defaultServiceName, fields["service"])
		assert.Equal(t, "c-1", fields["customer_id"])
	}
}

func TestRedactQuery(t *testing.T) {
	tests := map[string]struct {
		in   string
		want string
	}{
		"empty":                   {in: "", want: ""},
		"no sensitive params":     {in: "page=2&size=10", want: "page=2&size=10"},
		"document is redacted":    {in: "document=12345678901", want: "document=[REDACTED]"},
		"encoded key is redacted": {in: "%64ocument=123&page=1", want: "%64ocument=[REDACTED]&page=1"},
		"cpf and cnpj":            {in: "cpf=1&CNPJ=2", want: "cpf=[REDACTED]&CNPJ=[REDACTED]"},
		"token in query":          {in: "access_token=abc&jwt=def", want: "access_token=[REDACTED]&jwt=[REDACTED]"},
		"bad escape is redacted":  {in: "%zz=1&page=1", want: "[REDACTED]&page=1"},
		"key without value":       {in: "document&page=1", want: "document=[REDACTED]&page=1"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, RedactQuery(tc.in))
		})
	}
}
