package loggers

import (
	"context"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"go.uber.org/zap"
)

type Map map[string]interface{}

func LoggerToContext(ctx context.Context, zapLogger *zap.Logger) context.Context {
	return ctxzap.ToContext(ctx, zapLogger)
}

func Logger(ctx context.Context) *zap.Logger {
	return ctxzap.Extract(ctx)
}

func With(ctx context.Context, keyValues Map) context.Context {
	oldTags := grpc_ctxtags.Extract(ctx).Values()
	if oldTags == nil {
		oldTags = make(map[string]interface{})
	}

	tags := grpc_ctxtags.NewTags()

	for k, v := range oldTags {
		tags.Set(k, v)
	}
	for k, v := range keyValues {
		tags.Set(k, v)
	}
	return grpc_ctxtags.SetInContext(ctx, tags)
}
