package loggers

import (
	"context"
	"testing"
)

func TestLogger(t *testing.T) {
	ctx := LoggerToContext(context.Background(), NewZapLogger(true))

	ctx = With(ctx, Map{"key1": "value", "key2": 12345})
	Logger(ctx).Info("log with1 ...")

	With(ctx, Map{"key99": "value"}) // no output

	ctx = With(ctx, Map{"key3": "test", "key4": []byte("aabbcc")})
	Logger(ctx).Info("log with2 ...")
}
