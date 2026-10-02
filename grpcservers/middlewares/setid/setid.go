package setid

import (
	"context"
	"reflect"

	"github.com/bookweb/go-libs/contexts"
	"github.com/bookweb/go-libs/uuids"
	"google.golang.org/grpc"
)

func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		valueOf := reflect.ValueOf(req)
		typeOf := reflect.TypeOf(req)
		if typeOf.Kind() != reflect.Ptr {
			return handler(ctx, req)
		}
		ele := valueOf.Elem()
		if !ele.IsValid() {
			// logs.Logger.Warn("element is invalid", ele)
			// logs.Logger.Warn("element is invalid")
			return handler(ctx, req)
		}
		traceId := ele.FieldByName(string(contexts.ContextKey__TraceId))
		if traceId.Kind() != reflect.String {
			return handler(ctx, req)
		}
		if traceId.String() != "" {
			newCtx := context.WithValue(ctx, contexts.ContextKey__TraceId, traceId.String())
			return handler(newCtx, req)
		}
		id := uuids.UuidV7String()
		if traceId.CanSet() {
			traceId.SetString(id)
		}
		newCtx := context.WithValue(ctx, contexts.ContextKey__TraceId, id)

		return handler(newCtx, req)
	}
}
