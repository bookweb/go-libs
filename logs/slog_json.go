package logs

import (
	"io"
	"log/slog"
)

// CreateSlogJSONLogger helps to create slog JSON logger
func CreateSlogJSONLogger(writer io.Writer, opts ...Option) *slog.Logger {
	handlerOptions := &slog.HandlerOptions{}
	return slog.New(slog.NewJSONHandler(writer, handlerOptions))
}
