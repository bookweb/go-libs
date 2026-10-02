package logs

import (
	"context"
	"io"
	"log/slog"
	"math"
	"time"

	"github.com/bookweb/go-libs/contexts"
)

type LogKey string

const (
	LogKey_Time    LogKey = "time"
	LogKey_Level   LogKey = "level"
	LogKey_Message LogKey = "msg"    // or message
	LogKey_Source  LogKey = "source" // or caller
)

type LogLevel string

const (
	LogLevel_Debug LogLevel = "debug"
	LogLevel_Info  LogLevel = "info"
	LogLevel_Warn  LogLevel = "warn"
	LogLevel_Error LogLevel = "error"
)

type LoggerOptions struct {
	AppName  string
	LogLevel LogLevel

	CustomWriters []io.Writer

	MaskingFields map[string]string
}

const (
	Attr_TraceId string = "traceId"
)

var (
	_minTimeInt64 = time.Unix(0, math.MinInt64)
	_maxTimeInt64 = time.Unix(0, math.MaxInt64)
)

func GetTraceIdFromContext(ctx context.Context) string {
	switch valType := ctx.Value(contexts.ContextKey__TraceId).(type) {
	case string:
		return valType
	default:
		return ""
	}
}

func String(key string, val string) slog.Attr {
	return slog.String(key, val)
}

func Int64(key string, val int64) slog.Attr {
	return slog.Int64(key, val)
}

func Int32(key string, val int32) slog.Attr {
	return slog.Int(key, int(val))
}

func Int(key string, val int) slog.Attr {
	return slog.Int(key, int(val))
}

func Float64(key string, val float64) slog.Attr {
	return slog.Float64(key, val)
}

func Float32(key string, val float32) slog.Attr {
	return slog.Float64(key, float64(val))
}

func Bool(key string, val bool) slog.Attr {
	return slog.Bool(key, val)
}

func Time(key string, val time.Time) slog.Attr {
	return slog.Time(key, val)
}

func Struct(key string, val any) slog.Attr {
	return slog.Group(key, val)
}

func Any(key string, val any) slog.Attr {
	return slog.Any(key, val)
}

func Error(err error) slog.Attr {
	return slog.Any("error", err)
}

func NamedError(key string, err error) slog.Attr {
	return slog.Any(key, err)
}

func TraceId(val string) slog.Attr {
	return slog.String("traceId", val)
}

func TraceIdFromContext(ctx context.Context) slog.Attr {
	return slog.String("traceId", GetTraceIdFromContext(ctx))
}
