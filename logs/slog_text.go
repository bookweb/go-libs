package logs

import (
	"io"
	"log/slog"
)

// CreateSlogTextLogger helps to create slog Text logger
func CreateSlogTextLogger(writer io.Writer, opts ...Option) *slog.Logger {
	handlerOptions := &slog.HandlerOptions{}
	return slog.New(slog.NewTextHandler(writer, handlerOptions))
}

func NewSimpleTextLogger(opts ...Option) *slog.Logger {
	return CreateSlogTextLogger(GetDefaultSlogWriter(), opts...)
}
